package overtimebill

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/overtimerule"
	model "github.com/jzero-io/jzero-admin/server/internal/model/hrm_overtime_bill"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/overtimebill"
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

// Edit 修改加班申请单：仅 未提交/审批不通过/已撤销 状态可修改（演示站规则）。
// 单据编号/状态/核定与折算字段由系统维护，编辑只改申请信息（申请时长随起止时间重算，
// 原核定/补偿方式作废清零——提交后须重新核定）.
func (l *Edit) Edit(req *types.EditRequest) (resp *types.EditResponse, err error) {
	applyHours, err := validateOvertime(req.DayType, req.StartTime, req.EndTime)
	if err != nil {
		return nil, err
	}

	row, err := l.svcCtx.Model.HrmOvertimeBill.FindOneByUuid(l.ctx, nil, req.Uuid)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.OvertimeNotFoundCode, "加班单不存在")
		}
		return nil, err
	}
	if !canEdit(row.Status) {
		return nil, status.ErrorMessage(errcodes.OvertimeNotEditableCode,
			"当前单据状态不允许修改（仅 未提交/审批不通过/已撤销 可改）")
	}

	// 编辑重算最小单位（与提交同口径）：按员工当前生效规则解析最小计算单位，
	// 未命中规则按默认 30 分钟.
	minUnit := resolveMinUnitMinutes(l.ctx, l.svcCtx, req.EmployeeUuid, req.DayType)
	if hours, err := CalcOvertimeHours(req.StartTime, req.EndTime, minUnit); err == nil {
		applyHours = hours
	}

	fields := map[string]any{
		"employee_uuid": req.EmployeeUuid,
		"employee_no":   req.EmployeeNo,
		"employee_name": req.EmployeeName,
		"org_uuid":      req.OrgUuid,
		"org_name":      req.OrgName,
		"overtime_date": req.OvertimeDate,
		"day_type":      req.DayType,
		"start_time":    req.StartTime,
		"end_time":      req.EndTime,
		"apply_hours":   applyHours,
		"reason":        req.Reason,
		"confirm_hours": float64(0), // 申请信息变化后原核定作废
		"comp_type":     "",
	}
	if err = l.svcCtx.Model.HrmOvertimeBill.UpdateFieldsByCondition(l.ctx, nil, fields,
		conditionChainByUuid(req.Uuid)...); err != nil {
		return nil, err
	}

	return &types.EditResponse{}, nil
}

// resolveMinUnitMinutes 解析员工 当天类型 的生效加班规则最小计算单位.
// 员工未找到或规则未命中时返回 0（CalcOvertimeHours 内部按默认 30 分钟处理）.
func resolveMinUnitMinutes(ctx context.Context, svcCtx *svc.ServiceContext,
	employeeUuid, dayType string) int64 {
	if employeeUuid == "" {
		return 0
	}
	emp, err := svcCtx.Model.HrmEmployee.FindOneByUuid(ctx, nil, employeeUuid)
	if err != nil {
		return 0
	}
	groupUuid, orgUuid := attendanceScopeOf(ctx, svcCtx, emp)
	resolved, err := overtimerule.ResolveOvertimeRule(ctx, svcCtx, groupUuid, orgUuid, dayType)
	if err != nil || !resolved.Found {
		return 0
	}
	return resolved.MinUnitMinutes
}
