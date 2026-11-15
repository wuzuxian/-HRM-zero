package resignation

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
	hrmuser "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/hrmuser"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_resignation_bill"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/resignation"
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

// Add 新增员工离职单（HR204）：事务内生成单据编号，初始状态 未提交.
// 单据编号/状态列与共享审批底座（HRM-06）对齐，提交/审批走共享端点；
// 审批通过仅记录，员工置离职由定时任务在离职日统一处理.
func (l *Add) Add(req *types.AddRequest) (resp *types.AddResponse, err error) {
	if err = validateResignation(req.EmployeeNo, req.EmployeeName, req.ResignType,
		req.ResignDate, req.LastWorkDate, req.SalarySettleDate); err != nil {
		return nil, err
	}

	user := hrmuser.Current(l.ctx)
	applyDate := req.ApplyDate
	if applyDate == "" {
		applyDate = time.Now().Format(time.DateOnly) // 申请日期默认创建当日
	}

	var billNo string
	if txErr := l.svcCtx.SqlxConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		no, err := bill.GenerateBillNo(ctx, session, BillTypeCode, time.Now())
		if err != nil {
			return err
		}
		billNo = no
		err = l.svcCtx.Model.HrmResignationBill.InsertV2(ctx, session, &hrm_resignation_bill.HrmResignationBill{
			Uuid:             uuid.New().String(),
			BillNo:           no,
			Status:           bill.StatusUnsubmitted,
			EmployeeUuid:     req.EmployeeUuid,
			EmployeeNo:       req.EmployeeNo,
			EmployeeName:     req.EmployeeName,
			OrgUuid:          req.OrgUuid,
			CompanyUuid:      req.CompanyUuid,
			Position:         req.Position,
			Post:             req.Post,
			ResignType:       req.ResignType,
			ResignReason:     req.ResignReason,
			ApplyDate:        applyDate,
			ResignDate:       req.ResignDate,
			LastWorkDate:     req.LastWorkDate,
			HandoverPerson:   req.HandoverPerson,
			SalarySettleDate: req.SalarySettleDate,
			Applied:          0,
			CreateBy:         user.Uuid,
			CreatorName:      user.Username,
		})
		return err
	}); txErr != nil {
		return nil, txErr
	}

	l.Infof("resignation: 员工离职单已创建 billNo=%s employee=%s(%s) resignDate=%s",
		billNo, req.EmployeeName, req.EmployeeNo, req.ResignDate)
	return &types.AddResponse{}, nil
}
