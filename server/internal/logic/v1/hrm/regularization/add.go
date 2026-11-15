package regularization

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
	hrmuser "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/hrmuser"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_regularization_bill"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/regularization"
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

// Add 新增员工转正单（HR202）：事务内生成单据编号，初始状态 未提交.
// 员工快照（工号/姓名/部门/公司/职位/职务/人员状态/入职日期/预计转正日期）
// 由前端选择员工带出（spec 用户故事 20）；单据编号/状态列与共享审批底座
// （HRM-06）对齐，提交/审批走共享端点，审批通过回写见 bill.go.
func (l *Add) Add(req *types.AddRequest) (resp *types.AddResponse, err error) {
	if err = validateRegularization(req.EmployeeNo, req.EmployeeName, req.FormalDate); err != nil {
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
		err = l.svcCtx.Model.HrmRegularizationBill.InsertV2(ctx, session, &hrm_regularization_bill.HrmRegularizationBill{
			Uuid:               uuid.New().String(),
			BillNo:             no,
			Status:             bill.StatusUnsubmitted,
			EmployeeUuid:       req.EmployeeUuid,
			EmployeeNo:         req.EmployeeNo,
			EmployeeName:       req.EmployeeName,
			OrgUuid:            req.OrgUuid,
			CompanyUuid:        req.CompanyUuid,
			Position:           req.Position,
			Post:               req.Post,
			PersonStatus:       req.PersonStatus,
			HireDate:           req.HireDate,
			ExpectedFormalDate: req.ExpectedFormalDate,
			FormalDate:         req.FormalDate,
			Remark:             req.Remark,
			Applied:            0,
			CreateBy:           user.Uuid,
			CreatorName:        user.Username,
		})
		return err
	}); txErr != nil {
		return nil, txErr
	}

	l.Infof("regularization: 员工转正单已创建 billNo=%s employee=%s(%s)", billNo, req.EmployeeName, req.EmployeeNo)
	return &types.AddResponse{}, nil
}
