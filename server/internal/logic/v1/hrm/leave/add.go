package leave

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
	hrmuser "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/hrmuser"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_leave_bill"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/leave"
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

// timeNow 可测试钩子（单据编号日期部分用当天）。
var timeNow = time.Now

// Add 新增请假销假单（HR205）：事务内生成单据编号 + 按预计起止时间自动计算预计天数
// （0.5 天步进口径，口径说明见 days.go / 单据列表 notes），初始状态 未提交。
// 提交/审批/撤销走共享单据底座 v1/hrm/bill（类型码 205）。
func (l *Add) Add(req *types.AddRequest) (resp *types.AddResponse, err error) {
	user := hrmuser.Current(l.ctx)
	planDays, err := validateLeave(req.LeaveTypeCode, req.PlanStartTime, req.PlanEndTime)
	if err != nil {
		return nil, err
	}

	lt, err := fetchLeaveTypeFor(l.ctx, l.svcCtx, req.LeaveTypeCode)
	if err != nil {
		return nil, err
	}

	var billNo string
	if txErr := l.svcCtx.SqlxConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		no, err := bill.GenerateBillNo(ctx, session, BillTypeCode, timeNow())
		if err != nil {
			return err
		}
		billNo = no
		return l.svcCtx.Model.HrmLeaveBill.InsertV2(ctx, session, &hrm_leave_bill.HrmLeaveBill{
			Uuid:          uuid.New().String(),
			BillNo:        no,
			Status:        bill.StatusUnsubmitted,
			EmployeeUuid:  req.EmployeeUuid,
			EmployeeNo:    req.EmployeeNo,
			EmployeeName:  req.EmployeeName,
			OrgUuid:       req.OrgUuid,
			OrgName:       req.OrgName,
			LeaveTypeCode: req.LeaveTypeCode,
			LeaveTypeName: lt.TypeName,
			PlanStartTime: req.PlanStartTime,
			PlanEndTime:   req.PlanEndTime,
			PlanDays:      planDays,
			Reason:        req.Reason,
			Settled:       0,
			CreateBy:      user.Uuid,
			CreatorName:   user.Username,
		})
	}); txErr != nil {
		return nil, txErr
	}

	l.Infof("leave: 请假销假单已创建 billNo=%s employee=%s(%s) type=%s planDays=%.1f",
		billNo, req.EmployeeName, req.EmployeeNo, req.LeaveTypeCode, planDays)
	return &types.AddResponse{}, nil
}
