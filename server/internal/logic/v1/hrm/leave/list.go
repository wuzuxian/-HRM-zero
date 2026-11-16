package leave

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	hrmuser "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/hrmuser"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_leave_bill"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/leave"
)

type List struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewList(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *List {
	return &List{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// List 请假销假单分页列表，支持按单据编号/单据状态/请假类型筛选.
// 默认按「我的单据」过滤（创建人=当前用户，spec 用户故事 4）；
// 管理员级角色（持共享审批码 v1:hrm:bill:approve）可看全部，传 mine=1 时管理员也只看自己的单据.
func (l *List) List(req *types.ListRequest) (resp *types.ListResponse, err error) {
	user := hrmuser.Current(l.ctx)
	onlyMine := req.Mine == 1 || !hrmuser.CanSeeAllBills(l.ctx, l.svcCtx, user)

	chain := condition.NewChain().
		Equal(hrm_leave_bill.CreateBy, user.Uuid, condition.WithSkip(!onlyMine)).
		Like(hrm_leave_bill.BillNo, "%"+req.BillNo+"%", condition.WithSkip(req.BillNo == "")).
		Equal(hrm_leave_bill.Status, req.Status, condition.WithSkip(req.Status == "")).
		Equal(hrm_leave_bill.LeaveTypeCode, req.LeaveTypeCode, condition.WithSkip(req.LeaveTypeCode == "")).
		Page(req.Current, req.Size).
		OrderByDesc(hrm_leave_bill.Id)

	rows, total, err := l.svcCtx.Model.HrmLeaveBill.PageByCondition(l.ctx, nil, chain.Build()...)
	if err != nil {
		return nil, err
	}

	records := make([]types.HrmLeave, 0, len(rows))
	for _, row := range rows {
		records = append(records, leaveRecord(row))
	}

	return &types.ListResponse{PageResponse: types.PageResponse{
		Current: req.Current,
		Size:    req.Size,
		Total:   total,
	}, Records: records}, nil
}

func leaveRecord(row *hrm_leave_bill.HrmLeaveBill) types.HrmLeave {
	return types.HrmLeave{
		Uuid:            row.Uuid,
		BillNo:          row.BillNo,
		Status:          row.Status,
		EmployeeUuid:    row.EmployeeUuid,
		EmployeeNo:      row.EmployeeNo,
		EmployeeName:    row.EmployeeName,
		OrgUuid:         row.OrgUuid,
		OrgName:         row.OrgName,
		LeaveTypeCode:   row.LeaveTypeCode,
		LeaveTypeName:   row.LeaveTypeName,
		PlanStartTime:   row.PlanStartTime,
		PlanEndTime:     row.PlanEndTime,
		PlanDays:        row.PlanDays,
		Reason:          row.Reason,
		ActualStartTime: row.ActualStartTime,
		ActualEndTime:   row.ActualEndTime,
		ActualDays:      row.ActualDays,
		Settled:         row.Settled,
		CreatorUuid:     row.CreateBy,
		CreatorName:     row.CreatorName,
		CreateTime:      row.CreateTime.Format("2006-01-02 15:04:05"),
		UpdateTime:      row.UpdateTime.Format("2006-01-02 15:04:05"),
	}
}
