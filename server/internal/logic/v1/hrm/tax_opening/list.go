package tax_opening

import (
	"context"
	"net/http"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_tax_opening"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/tax_opening"
)

type List struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewList(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *List {
	return &List{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

func (l *List) List(req *types.ListRequest) (resp *types.ListResponse, err error) {
	rows, total, err := l.svcCtx.Model.HrmTaxOpening.PageByCondition(l.ctx, nil,
		condition.NewChain().
			Page(req.Current, req.Size).
			Like(hrm_tax_opening.EmployeeName, "%"+req.EmployeeName+"%", condition.WithSkip(req.EmployeeName == "")).
			Equal(hrm_tax_opening.SubjectUuid, req.SubjectUuid, condition.WithSkip(req.SubjectUuid == "")).
			Equal(hrm_tax_opening.TaxYear, req.TaxYear, condition.WithSkip(req.TaxYear == "")).
			OrderByDesc(hrm_tax_opening.Id).
			Build()...)
	if err != nil {
		return nil, err
	}

	// 身份证列从员工表带出（demo-inventory 列：员工、身份证…）
	idCardByUuid := make(map[string]string)
	if len(rows) > 0 {
		employees, err := l.svcCtx.Model.HrmEmployee.FindByCondition(l.ctx, nil)
		if err != nil {
			return nil, err
		}
		for _, emp := range employees {
			idCardByUuid[emp.Uuid] = emp.IdCard
		}
	}

	records := make([]types.HrmTaxOpening, 0, len(rows))
	for _, row := range rows {
		records = append(records, types.HrmTaxOpening{
			Uuid:                  row.Uuid,
			EmployeeUuid:          row.EmployeeUuid,
			EmployeeName:          row.EmployeeName,
			IdCard:                idCardByUuid[row.EmployeeUuid],
			SubjectUuid:           row.SubjectUuid,
			SubjectName:           row.SubjectName,
			TaxYear:               row.TaxYear,
			UntilMonth:            row.UntilMonth,
			IncomeTotal:           row.IncomeTotal,
			SpecialDeductionTotal: row.SpecialDeductionTotal,
			SpecialExtraTotal:     row.SpecialExtraTotal,
			ExemptTotal:           row.ExemptTotal,
			OtherDeductionTotal:   row.OtherDeductionTotal,
			TaxableIncomeTotal:    row.TaxableIncomeTotal,
			TaxPaidTotal:          row.TaxPaidTotal,
			Remark:                row.Remark,
			CreateTime:            row.CreateTime.Format(time.DateTime),
			UpdateTime:            row.UpdateTime.Format(time.DateTime),
		})
	}

	return &types.ListResponse{PageResponse: types.PageResponse{
		Current: req.Current,
		Size:    req.Size,
		Total:   total,
	}, Records: records}, nil
}
