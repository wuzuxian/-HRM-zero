package tax_opening

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_tax_opening"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/tax_opening"
)

type Add struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewAdd(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Add {
	return &Add{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

func (l *Add) Add(req *types.AddRequest) (resp *types.AddResponse, err error) {
	if err = validateOpeningCommon(req.EmployeeUuid, req.TaxYear, req.UntilMonth,
		req.IncomeTotal, req.SpecialDeductionTotal, req.SpecialExtraTotal, req.ExemptTotal,
		req.OtherDeductionTotal, req.TaxableIncomeTotal, req.TaxPaidTotal); err != nil {
		return nil, err
	}

	employee, err := l.svcCtx.Model.HrmEmployee.FindOneByUuid(l.ctx, nil, req.EmployeeUuid)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.TaxOpeningEmployeeNotFoundCode, "关联员工不存在")
		}
		return nil, err
	}

	// 主体快照（HRM-33 薪酬主体；不校验主体状态，停用主体允许存量期初只读展示）
	subjectName := ""
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

	// 唯一守卫: 员工 + 主体 + 年度
	dup, err := l.svcCtx.Model.HrmTaxOpening.FindOneByCondition(l.ctx, nil,
		openingUniqueCondition(req.EmployeeUuid, req.SubjectUuid, req.TaxYear)...)
	if err != nil && err != hrm_tax_opening.ErrNotFound {
		return nil, err
	}
	if dup != nil {
		return nil, status.ErrorMessage(errcodes.TaxOpeningDuplicateCode, "该员工该主体该年度已存在累计期初")
	}

	if _, err = l.svcCtx.Model.HrmTaxOpening.Insert(l.ctx, nil, &hrm_tax_opening.HrmTaxOpening{
		Uuid:                  uuid.New().String(),
		EmployeeUuid:          req.EmployeeUuid,
		EmployeeName:          employee.Name,
		SubjectUuid:           req.SubjectUuid,
		SubjectName:           subjectName,
		TaxYear:               req.TaxYear,
		UntilMonth:            req.UntilMonth,
		IncomeTotal:           req.IncomeTotal,
		SpecialDeductionTotal: req.SpecialDeductionTotal,
		SpecialExtraTotal:     req.SpecialExtraTotal,
		ExemptTotal:           req.ExemptTotal,
		OtherDeductionTotal:   req.OtherDeductionTotal,
		TaxableIncomeTotal:    req.TaxableIncomeTotal,
		TaxPaidTotal:          req.TaxPaidTotal,
		Remark:                req.Remark,
	}); err != nil {
		return nil, err
	}

	return &types.AddResponse{}, nil
}
