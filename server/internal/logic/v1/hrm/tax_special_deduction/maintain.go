package tax_special_deduction

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_tax_special_deduction"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/tax_special_deduction"
)

type Maintain struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewMaintain(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Maintain {
	return &Maintain{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// Maintain 行操作「维护」弹窗提交: 员工 + 扣除年度的一批 (月份, 扣除项, 金额) 格子。
// 已存在的行覆盖金额（部分月份更新——未提交的格子不动），不存在的行插入；整个提交一个事务。
// 同一提交内 (月份, 扣除项) 重复即拒绝。
func (l *Maintain) Maintain(req *types.MaintainRequest) (resp *types.MaintainResponse, err error) {
	if err = validateDeductionCommon(req.EmployeeUuid, req.DeductYear, 1, DeductTypeChildrenEdu, 0); err != nil {
		return nil, err
	}
	if len(req.Rows) == 0 {
		return nil, status.ErrorMessage(errcodes.TaxSpecialDeductionInvalidCode, "维护内容不能为空")
	}

	employee, err := l.svcCtx.Model.HrmEmployee.FindOneByUuid(l.ctx, nil, req.EmployeeUuid)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.TaxSpecialDeductionEmployeeNotFoundCode, "关联员工不存在")
		}
		return nil, err
	}

	seen := make(map[string]bool, len(req.Rows))
	rows := make([]*hrm_tax_special_deduction.HrmTaxSpecialDeduction, 0, len(req.Rows))
	for _, r := range req.Rows {
		if err = validateDeductionCommon(req.EmployeeUuid, req.DeductYear, r.DeductMonth, r.DeductType, r.Amount); err != nil {
			return nil, err
		}
		key := deductionKey(r.DeductMonth, r.DeductType)
		if seen[key] {
			return nil, status.ErrorMessage(errcodes.TaxSpecialDeductionDuplicateCode, "同一年度同月同一扣除项在提交中重复")
		}
		seen[key] = true
		rows = append(rows, &hrm_tax_special_deduction.HrmTaxSpecialDeduction{
			Uuid:         uuid.New().String(),
			EmployeeUuid: req.EmployeeUuid,
			EmployeeName: employee.Name,
			DeductYear:   req.DeductYear,
			DeductMonth:  r.DeductMonth,
			DeductType:   r.DeductType,
			Amount:       r.Amount,
			Source:       SourceManual,
		})
	}

	var saved int64
	err = l.svcCtx.SqlxConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		saved = 0
		for _, row := range rows {
			existing, err := l.svcCtx.Model.HrmTaxSpecialDeduction.FindOneByCondition(ctx, session,
				deductionUniqueCondition(row.EmployeeUuid, row.DeductYear, row.DeductMonth, row.DeductType)...)
			if err != nil && err != hrm_tax_special_deduction.ErrNotFound {
				return err
			}
			if existing != nil {
				// 部分月份更新: 只覆盖金额/来源/姓名快照，未提交格子不动
				if err = l.svcCtx.Model.HrmTaxSpecialDeduction.UpdateFieldsByCondition(ctx, session, map[string]any{
					string(hrm_tax_special_deduction.Amount):       row.Amount,
					string(hrm_tax_special_deduction.Source):       row.Source,
					string(hrm_tax_special_deduction.EmployeeName): employee.Name,
				}, deductionUuidCondition(existing.Uuid)...); err != nil {
					return err
				}
			} else {
				if _, err = l.svcCtx.Model.HrmTaxSpecialDeduction.Insert(ctx, session, row); err != nil {
					return err
				}
			}
			saved++
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &types.MaintainResponse{Saved: saved}, nil
}

// employeeOf 按员工 uuid 取员工（供维护/年度维护/导入复用）。
func employeeOf(l *Maintain, employeeUuid string) (*hrm_employee.HrmEmployee, error) {
	return l.svcCtx.Model.HrmEmployee.FindOneByUuid(l.ctx, nil, employeeUuid)
}
