package resignation

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_resignation_bill"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/resignation"
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

// Edit 修改员工离职单：仅 未提交/审批不通过/已撤销 状态可修改（演示站规则）.
// 单据编号与状态由系统维护，不允许通过编辑改动.
func (l *Edit) Edit(req *types.EditRequest) (resp *types.EditResponse, err error) {
	if err = validateResignation(req.EmployeeNo, req.EmployeeName, req.ResignType,
		req.ResignDate, req.LastWorkDate, req.SalarySettleDate); err != nil {
		return nil, err
	}

	row, err := l.svcCtx.Model.HrmResignationBill.FindOneByUuid(l.ctx, nil, req.Uuid)
	if err != nil {
		if err == hrm_resignation_bill.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.ResignationNotFoundCode, "离职单不存在")
		}
		return nil, err
	}
	if !canEdit(row.Status) {
		return nil, status.ErrorMessage(errcodes.ResignationNotEditableCode,
			"当前单据状态不允许修改（仅 未提交/审批不通过/已撤销 可改）")
	}

	fields := map[string]any{
		"employee_uuid":      req.EmployeeUuid,
		"employee_no":        req.EmployeeNo,
		"employee_name":      req.EmployeeName,
		"org_uuid":           req.OrgUuid,
		"company_uuid":       req.CompanyUuid,
		"position":           req.Position,
		"post":               req.Post,
		"resign_type":        req.ResignType,
		"resign_reason":      req.ResignReason,
		"apply_date":         req.ApplyDate,
		"resign_date":        req.ResignDate,
		"last_work_date":     req.LastWorkDate,
		"handover_person":    req.HandoverPerson,
		"salary_settle_date": req.SalarySettleDate,
	}
	if err = l.svcCtx.Model.HrmResignationBill.UpdateFieldsByCondition(l.ctx, nil, fields,
		conditionChainByUuid(req.Uuid)...); err != nil {
		return nil, err
	}

	return &types.EditResponse{}, nil
}
