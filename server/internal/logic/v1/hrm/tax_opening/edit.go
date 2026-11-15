package tax_opening

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_tax_opening"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/tax_opening"
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

func (l *Edit) Edit(req *types.EditRequest) (resp *types.EditResponse, err error) {
	if err = validateOpeningCommon(req.EmployeeUuid, req.TaxYear, req.UntilMonth,
		req.IncomeTotal, req.SpecialDeductionTotal, req.SpecialExtraTotal, req.ExemptTotal,
		req.OtherDeductionTotal, req.TaxableIncomeTotal, req.TaxPaidTotal); err != nil {
		return nil, err
	}

	row, err := l.svcCtx.Model.HrmTaxOpening.FindOneByUuid(l.ctx, nil, req.Uuid)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.TaxOpeningNotFoundCode, "个税累计期初不存在")
		}
		return nil, err
	}

	// 员工快照随换随更；唯一守卫（员工+主体+年度）排除自身
	employeeName := row.EmployeeName
	if req.EmployeeUuid != row.EmployeeUuid {
		employee, err := l.svcCtx.Model.HrmEmployee.FindOneByUuid(l.ctx, nil, req.EmployeeUuid)
		if err != nil {
			if err == sqlx.ErrNotFound {
				return nil, status.ErrorMessage(errcodes.TaxOpeningEmployeeNotFoundCode, "关联员工不存在")
			}
			return nil, err
		}
		employeeName = employee.Name
	}
	if req.EmployeeUuid != row.EmployeeUuid || req.SubjectUuid != row.SubjectUuid || req.TaxYear != row.TaxYear {
		dup, err := l.svcCtx.Model.HrmTaxOpening.FindOneByCondition(l.ctx, nil,
			openingUniqueCondition(req.EmployeeUuid, req.SubjectUuid, req.TaxYear)...)
		if err != nil && err != hrm_tax_opening.ErrNotFound {
			return nil, err
		}
		if dup != nil && dup.Uuid != row.Uuid {
			return nil, status.ErrorMessage(errcodes.TaxOpeningDuplicateCode, "该员工该主体该年度已存在累计期初")
		}
	}

	subjectName := row.SubjectName
	if req.SubjectUuid != row.SubjectUuid {
		subjectName = ""
		if req.SubjectUuid != "" {
			subject, err := l.svcCtx.Model.HrmSalarySubject.FindOneByUuid(l.ctx, nil, req.SubjectUuid)
			if err != nil {
				if err == sqlx.ErrNotFound {
					return nil, status.ErrorMessage(errcodes.TaxOpeningInvalidCode, "薪酬主体不存在")
				}
				return nil, err
			}
			subjectName = subject.SubjectName
		}
	}

	if err = l.svcCtx.Model.HrmTaxOpening.UpdateFieldsByCondition(l.ctx, nil, map[string]any{
		string(hrm_tax_opening.EmployeeUuid):          req.EmployeeUuid,
		string(hrm_tax_opening.EmployeeName):          employeeName,
		string(hrm_tax_opening.SubjectUuid):           req.SubjectUuid,
		string(hrm_tax_opening.SubjectName):           subjectName,
		string(hrm_tax_opening.TaxYear):               req.TaxYear,
		string(hrm_tax_opening.UntilMonth):            req.UntilMonth,
		string(hrm_tax_opening.IncomeTotal):           req.IncomeTotal,
		string(hrm_tax_opening.SpecialDeductionTotal): req.SpecialDeductionTotal,
		string(hrm_tax_opening.SpecialExtraTotal):     req.SpecialExtraTotal,
		string(hrm_tax_opening.ExemptTotal):           req.ExemptTotal,
		string(hrm_tax_opening.OtherDeductionTotal):   req.OtherDeductionTotal,
		string(hrm_tax_opening.TaxableIncomeTotal):    req.TaxableIncomeTotal,
		string(hrm_tax_opening.TaxPaidTotal):          req.TaxPaidTotal,
		string(hrm_tax_opening.Remark):                req.Remark,
	}, openingUuidCondition(req.Uuid)...); err != nil {
		return nil, err
	}

	return &types.EditResponse{}, nil
}
