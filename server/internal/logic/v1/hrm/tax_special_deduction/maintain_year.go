package tax_special_deduction

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_tax_special_deduction"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/tax_special_deduction"
)

type MaintainYear struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewMaintainYear(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *MaintainYear {
	return &MaintainYear{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

func (l *MaintainYear) MaintainYear(req *types.MaintainYearRequest) (resp *types.MaintainYearResponse, err error) {
	if err = validateDeductionCommon(req.EmployeeUuid, req.DeductYear, 1, req.DeductType, req.Amount); err != nil {
		return nil, err
	}

	employee, err := l.svcCtx.Model.HrmEmployee.FindOneByUuid(l.ctx, nil, req.EmployeeUuid)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.TaxSpecialDeductionEmployeeNotFoundCode, "关联员工不存在")
		}
		return nil, err
	}

	months := req.Months
	if len(months) == 0 {
		months = []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	}
	monthSet := make(map[int64]bool, len(months))
	for _, m := range months {
		if !validMonth(m) {
			return nil, status.ErrorMessage(errcodes.TaxSpecialDeductionInvalidCode, "扣除月份必须在 1-12 之间")
		}
		monthSet[m] = true
	}

	var saved int64
	err = l.svcCtx.SqlxConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		saved = 0
		for m := int64(1); m <= 12; m++ {
			if !monthSet[m] {
				continue
			}
			existing, err := l.svcCtx.Model.HrmTaxSpecialDeduction.FindOneByCondition(ctx, session,
				deductionUniqueCondition(req.EmployeeUuid, req.DeductYear, m, req.DeductType)...)
			if err != nil && err != hrm_tax_special_deduction.ErrNotFound {
				return err
			}
			if existing != nil {
				if err = l.svcCtx.Model.HrmTaxSpecialDeduction.UpdateFieldsByCondition(ctx, session, map[string]any{
					string(hrm_tax_special_deduction.Amount):       req.Amount,
					string(hrm_tax_special_deduction.Source):       SourceManual,
					string(hrm_tax_special_deduction.EmployeeName): employee.Name,
				}, deductionUuidCondition(existing.Uuid)...); err != nil {
					return err
				}
			} else {
				if _, err = l.svcCtx.Model.HrmTaxSpecialDeduction.Insert(ctx, session, &hrm_tax_special_deduction.HrmTaxSpecialDeduction{
					Uuid:         uuid.New().String(),
					EmployeeUuid: req.EmployeeUuid,
					EmployeeName: employee.Name,
					DeductYear:   req.DeductYear,
					DeductMonth:  m,
					DeductType:   req.DeductType,
					Amount:       req.Amount,
					Source:       SourceManual,
				}); err != nil {
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

	return &types.MaintainYearResponse{Saved: saved}, nil
}
