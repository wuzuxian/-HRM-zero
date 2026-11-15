package entry

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_entry_bill"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/entry"
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

// Edit 修改入职单：仅 未提交/审批不通过/已撤销 状态可修改（演示站规则）.
// 单据编号与状态由系统维护，不允许通过编辑改动；已建档（审批通过）的入职单
// 不可再改（employee_uuid 非空锁定）.
func (l *Edit) Edit(req *types.EditRequest) (resp *types.EditResponse, err error) {
	if err = validateEntry(req.Name, req.Gender, req.MaritalStatus, req.HireDate,
		req.ExpectFormalDate, req.ProbationMonths); err != nil {
		return nil, err
	}

	row, err := l.svcCtx.Model.HrmEntryBill.FindOneByUuid(l.ctx, nil, req.Uuid)
	if err != nil {
		if err == hrm_entry_bill.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.EntryNotFoundCode, "入职单不存在")
		}
		return nil, err
	}
	if !canEdit(row.Status) {
		return nil, status.ErrorMessage(errcodes.EntryNotEditableCode,
			"当前单据状态不允许修改（仅 未提交/审批不通过/已撤销 可改）")
	}
	if row.EmployeeUuid != "" {
		return nil, status.ErrorMessage(errcodes.EntryArchivedCode,
			"入职单已自动建档，不允许修改")
	}

	fields := map[string]any{
		"name":               req.Name,
		"phone":              req.Phone,
		"email":              req.Email,
		"gender":             req.Gender,
		"birthday":           req.Birthday,
		"nation":             req.Nation,
		"political_status":   req.PoliticalStatus,
		"marital_status":     req.MaritalStatus,
		"culture_level":      req.CultureLevel,
		"native_place":       req.NativePlace,
		"address":            req.Address,
		"household_address":  req.HouseholdAddress,
		"emergency_contact":  req.EmergencyContact,
		"emergency_phone":    req.EmergencyPhone,
		"bank_name":          req.BankName,
		"bank_account":       req.BankAccount,
		"avatar":             req.Avatar,
		"position":           req.Position,
		"post":               req.Post,
		"probation_months":   req.ProbationMonths,
		"org_uuid":           req.OrgUuid,
		"company_uuid":       req.CompanyUuid,
		"hire_date":          req.HireDate,
		"expect_formal_date": expectFormalOf(req.HireDate, req.ExpectFormalDate, req.ProbationMonths),
		"attachment_desc":    req.AttachmentDesc,
		"remark":             req.Remark,
	}
	if err = l.svcCtx.Model.HrmEntryBill.UpdateFieldsByCondition(l.ctx, nil, fields,
		condition.NewChain().Equal(hrm_entry_bill.Uuid, req.Uuid).Build()...); err != nil {
		return nil, err
	}

	// 三子表快照整体替换（全删全插, 与员工档案 edit 语义一致）
	if txErr := l.svcCtx.SqlxConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		if err := deleteChildren(ctx, session, l.svcCtx, req.Uuid); err != nil {
			return err
		}
		return insertChildren(ctx, session, l.svcCtx, req.Uuid, req.Works, req.Educations, req.Families)
	}); txErr != nil {
		return nil, txErr
	}

	return &types.EditResponse{}, nil
}
