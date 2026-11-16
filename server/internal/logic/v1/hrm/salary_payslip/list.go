package salary_payslip

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_payslip"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_payslip"
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

// List 工资条管理列表（筛选 批次/员工姓名/年月/已读状态；年月倒序）。
func (l *List) List(req *types.ListRequest) (resp *types.ListResponse, err error) {
	chain := condition.NewChain().
		Equal(hrm_salary_payslip.BatchUuid, req.BatchUuid, condition.WithSkip(req.BatchUuid == "")).
		Like(hrm_salary_payslip.EmployeeName, "%"+req.EmployeeName+"%", condition.WithSkip(req.EmployeeName == "")).
		OrderByDesc(hrm_salary_payslip.BatchYear).
		OrderByDesc(hrm_salary_payslip.BatchMonth).
		OrderByDesc(hrm_salary_payslip.Id).
		Page(int(req.Current), int(req.Size))

	// 年月筛选（yearMonth 'YYYY-MM'）
	if req.YearMonth != "" {
		year, month, ok := splitYearMonth(req.YearMonth)
		if ok {
			chain = chain.Equal(hrm_salary_payslip.BatchYear, year).Equal(hrm_salary_payslip.BatchMonth, month)
		}
	}
	// 已读筛选：'0'=未读 '1'=已读（''=全部）
	switch req.ReadStatus {
	case "0", "1":
		chain = chain.Equal(hrm_salary_payslip.ReadStatus, req.ReadStatus)
	}

	rows, total, err := l.svcCtx.Model.HrmSalaryPayslip.PageByCondition(l.ctx, nil, chain.Build()...)
	if err != nil {
		return nil, err
	}
	records := make([]types.HrmSalaryPayslip, 0, len(rows))
	for _, row := range rows {
		records = append(records, payslipRowToRecord(row))
	}
	return &types.ListResponse{PageResponse: types.PageResponse{
		Current: req.Current,
		Size:    req.Size,
		Total:   total,
	}, Records: records}, nil
}
