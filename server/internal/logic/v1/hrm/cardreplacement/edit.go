package cardreplacement

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_card_replacement_bill"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/cardreplacement"
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

// Edit 修改补卡单：仅 未提交/审批不通过/已撤销 状态可修改（演示站规则）。
// 单据编号/状态/原打卡快照由系统维护；编辑改 补卡日期/类型/时间/原因，
// 日期或类型变更时重取原打卡快照.
func (l *Edit) Edit(req *types.EditRequest) (resp *types.EditResponse, err error) {
	if err := validateCardReplacement(req.CardType, req.PunchDate, req.PunchTime); err != nil {
		return nil, err
	}

	row, err := l.svcCtx.Model.HrmCardReplacementBill.FindOneByUuid(l.ctx, nil, req.Uuid)
	if err != nil {
		if err == hrm_card_replacement_bill.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.CardReplacementNotFoundCode, "补卡单不存在")
		}
		return nil, err
	}
	if !canEdit(row.Status) {
		return nil, status.ErrorMessage(errcodes.CardReplacementNotEditableCode,
			"当前单据状态不允许修改（仅 未提交/审批不通过/已撤销 可改）")
	}

	employee, err := l.svcCtx.Model.HrmEmployee.FindOneByCondition(l.ctx, nil,
		condition.NewChain().Equal(hrm_employee.Uuid, req.EmployeeUuid).Build()...)
	if err != nil {
		if err == hrm_employee.ErrNotFound {
			return nil, employeeNotFound(req.EmployeeUuid)
		}
		return nil, err
	}

	origIn, origOut := row.OrigClockInTime, row.OrigClockOutTime
	if row.PunchDate != req.PunchDate {
		origIn, origOut = origClocksOf(l.ctx, nil, l.svcCtx, req.EmployeeUuid, req.PunchDate)
	}

	fields := map[string]any{
		"employee_uuid":       employee.Uuid,
		"employee_no":         employee.EmployeeNo,
		"employee_name":       employee.Name,
		"org_uuid":            req.OrgUuid,
		"org_name":            req.OrgName,
		"punch_date":          req.PunchDate,
		"card_type":           req.CardType,
		"punch_time":          req.PunchTime,
		"reason":              req.Reason,
		"orig_clock_in_time":  origIn,
		"orig_clock_out_time": origOut,
	}
	if err = l.svcCtx.Model.HrmCardReplacementBill.UpdateFieldsByCondition(l.ctx, nil, fields,
		conditionByUuid(req.Uuid)...); err != nil {
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

// conditionByUuid 单据 uuid 条件.
func conditionByUuid(billUuid string) []condition.Condition {
	return condition.NewChain().Equal(hrm_card_replacement_bill.Uuid, billUuid).Build()
}
