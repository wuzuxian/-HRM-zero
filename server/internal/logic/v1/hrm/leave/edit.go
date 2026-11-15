package leave

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_leave_bill"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_leave_type"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/leave"
)

type Edit struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewEdit(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Edit {
	return &Edit{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// Edit 修改请假销假单：仅 未提交/审批不通过/已撤销 状态可修改（演示站规则）。
// 单据编号/状态/实际回填字段由系统维护，编辑只改 预计信息（天数随起止时间重算）.
func (l *Edit) Edit(req *types.EditRequest) (resp *types.EditResponse, err error) {
	planDays, err := validateLeave(req.LeaveTypeCode, req.PlanStartTime, req.PlanEndTime)
	if err != nil {
		return nil, err
	}

	row, err := l.svcCtx.Model.HrmLeaveBill.FindOneByUuid(l.ctx, nil, req.Uuid)
	if err != nil {
		if err == hrm_leave_bill.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.LeaveNotFoundCode, "请假单不存在")
		}
		return nil, err
	}
	if !canEdit(row.Status) {
		return nil, status.ErrorMessage(errcodes.LeaveNotEditableCode,
			"当前单据状态不允许修改（仅 未提交/审批不通过/已撤销 可改）")
	}

	lt, err := fetchLeaveTypeFor(l.ctx, l.svcCtx, req.LeaveTypeCode)
	if err != nil {
		return nil, err
	}

	fields := map[string]any{
		"employee_uuid":   req.EmployeeUuid,
		"employee_no":     req.EmployeeNo,
		"employee_name":   req.EmployeeName,
		"org_uuid":        req.OrgUuid,
		"org_name":        req.OrgName,
		"leave_type_code": req.LeaveTypeCode,
		"leave_type_name": lt.TypeName,
		"plan_start_time": req.PlanStartTime,
		"plan_end_time":   req.PlanEndTime,
		"plan_days":       planDays,
		"reason":          req.Reason,
	}
	if err = l.svcCtx.Model.HrmLeaveBill.UpdateFieldsByCondition(l.ctx, nil, fields,
		conditionChainByUuid(req.Uuid)...); err != nil {
		return nil, err
	}

	return &types.EditResponse{}, nil
}

// canEdit 与共享单据状态机一致的「可编辑」状态集（未提交/审批不通过/已撤销）.
func canEdit(statusCode string) bool {
	switch statusCode {
	case bill.StatusUnsubmitted, bill.StatusRejected, bill.StatusRevoked:
		return true
	}
	return false
}

// fetchLeaveTypeFor 校验假期类型存在且启用（Edit 用；Add 上的方法版本不便于复用）.
func fetchLeaveTypeFor(ctx context.Context, svcCtx *svc.ServiceContext, leaveTypeCode string) (*hrm_leave_type.HrmLeaveType, error) {
	lt, err := svcCtx.Model.HrmLeaveType.FindOneByCondition(ctx, nil,
		conditionByTypeCode(leaveTypeCode).Build()...)
	if err != nil {
		if err == hrm_leave_type.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.LeaveTypeNotFoundCode, "假期类型不存在: "+leaveTypeCode)
		}
		return nil, err
	}
	if lt.Status != "1" {
		return nil, status.ErrorMessage(errcodes.LeaveTypeNotFoundCode, "假期类型已停用: "+lt.TypeName)
	}
	return lt, nil
}
