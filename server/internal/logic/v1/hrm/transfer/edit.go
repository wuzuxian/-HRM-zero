package transfer

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_transfer_bill"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/transfer"
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

// Edit 修改人事调动单：仅 未提交/审批不通过/已撤销 状态可修改（演示站规则）.
// 单据编号与状态由系统维护，不允许通过编辑改动.
func (l *Edit) Edit(req *types.EditRequest) (resp *types.EditResponse, err error) {
	if err = validateTransfer(req.EmployeeNo, req.EmployeeName, req.ChangeType,
		req.EffectiveImmediately, req.EffectiveDate); err != nil {
		return nil, err
	}

	row, err := l.svcCtx.Model.HrmTransferBill.FindOneByUuid(l.ctx, nil, req.Uuid)
	if err != nil {
		if err == hrm_transfer_bill.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.TransferNotFoundCode, "调动单不存在")
		}
		return nil, err
	}
	if !canEdit(row.Status) {
		return nil, status.ErrorMessage(errcodes.TransferNotEditableCode,
			"当前单据状态不允许修改（仅 未提交/审批不通过/已撤销 可改）")
	}

	fields := map[string]any{
		"employee_uuid":         req.EmployeeUuid,
		"employee_no":           req.EmployeeNo,
		"employee_name":         req.EmployeeName,
		"change_type":           req.ChangeType,
		"original_company":      req.OriginalCompany,
		"original_org":          req.OriginalOrg,
		"original_position":     req.OriginalPosition,
		"original_post":         req.OriginalPost,
		"new_company":           req.NewCompany,
		"new_org":               req.NewOrg,
		"new_position":          req.NewPosition,
		"new_post":              req.NewPost,
		"new_salary":            req.NewSalary,
		"transfer_reason":       req.TransferReason,
		"effective_immediately": req.EffectiveImmediately,
		"effective_date":        req.EffectiveDate,
	}
	if err = l.svcCtx.Model.HrmTransferBill.UpdateFieldsByCondition(l.ctx, nil, fields,
		conditionChainByUuid(req.Uuid)...); err != nil {
		return nil, err
	}

	return &types.EditResponse{}, nil
}
