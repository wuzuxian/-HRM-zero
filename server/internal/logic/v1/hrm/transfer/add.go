package transfer

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
	hrmuser "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/hrmuser"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_transfer_bill"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/transfer"
)

type Add struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewAdd(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Add {
	return &Add{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// Add 新增人事调动单（HR203）：事务内生成单据编号，初始状态 未提交.
// 单据编号/状态列与共享审批底座（HRM-06）对齐，提交/审批走共享端点.
func (l *Add) Add(req *types.AddRequest) (resp *types.AddResponse, err error) {
	if err = validateTransfer(req.EmployeeNo, req.EmployeeName, req.ChangeType,
		req.EffectiveImmediately, req.EffectiveDate); err != nil {
		return nil, err
	}

	user := hrmuser.Current(l.ctx)

	var billNo string
	if txErr := l.svcCtx.SqlxConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		no, err := bill.GenerateBillNo(ctx, session, BillTypeCode, time.Now())
		if err != nil {
			return err
		}
		billNo = no
		err = l.svcCtx.Model.HrmTransferBill.InsertV2(ctx, session, &hrm_transfer_bill.HrmTransferBill{
			Uuid:                 uuid.New().String(),
			BillNo:               no,
			Status:               bill.StatusUnsubmitted,
			EmployeeUuid:         req.EmployeeUuid,
			EmployeeNo:           req.EmployeeNo,
			EmployeeName:         req.EmployeeName,
			ChangeType:           req.ChangeType,
			OriginalCompany:      req.OriginalCompany,
			OriginalOrg:          req.OriginalOrg,
			OriginalPosition:     req.OriginalPosition,
			OriginalPost:         req.OriginalPost,
			NewCompany:           req.NewCompany,
			NewOrg:               req.NewOrg,
			NewPosition:          req.NewPosition,
			NewPost:              req.NewPost,
			NewSalary:            req.NewSalary,
			TransferReason:       req.TransferReason,
			EffectiveImmediately: req.EffectiveImmediately,
			EffectiveDate:        req.EffectiveDate,
			Applied:              0,
			CreateBy:             user.Uuid,
			CreatorName:          user.Username,
		})
		return err
	}); txErr != nil {
		return nil, txErr
	}

	l.Infof("transfer: 人事调动单已创建 billNo=%s employee=%s(%s)", billNo, req.EmployeeName, req.EmployeeNo)
	return &types.AddResponse{}, nil
}
