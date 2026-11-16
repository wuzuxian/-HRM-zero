package salary_payslip

import (
	"context"
	"net/http"
	"strings"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_payslip"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_payslip"
)

type Ledger struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewLedger(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Ledger {
	return &Ledger{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// Ledger 薪资统计台账（demo-inventory 8.4 /hrm/salary/statistics）：
// 统计卡（员工数/应发合计/应扣合计/实发合计/人工成本）+ 按批次×员工明细行
// （异常=批次 warnings 透传；已发薪/已发条=批次 pay_status/payslip_status 只读透传）。
func (l *Ledger) Ledger(req *types.LedgerRequest) (resp *types.LedgerResponse, err error) {
	chain := condition.NewChain().
		Equal(hrm_salary_payslip.BatchUuid, req.BatchUuid, condition.WithSkip(req.BatchUuid == "")).
		Like(hrm_salary_payslip.EmployeeName, "%"+req.EmployeeName+"%", condition.WithSkip(req.EmployeeName == "")).
		Like(hrm_salary_payslip.OrgName, "%"+req.OrgName+"%", condition.WithSkip(req.OrgName == "")).
		OrderByDesc(hrm_salary_payslip.BatchYear).
		OrderByDesc(hrm_salary_payslip.BatchMonth).
		OrderByDesc(hrm_salary_payslip.Id)

	rows, _, err := l.svcCtx.Model.HrmSalaryPayslip.PageByCondition(l.ctx, nil,
		chain.Page(int(req.Current), int(req.Size)).Build()...)
	if err != nil {
		return nil, err
	}

	// 统计卡与异常过滤需要全量行（统计卡口径=筛选命中的全部工资条；分页只作用于明细列表）。
	all, err := l.svcCtx.Model.HrmSalaryPayslip.FindByCondition(l.ctx, nil, chain.Build()...)
	if err != nil {
		return nil, err
	}

	// 批次上下文（异常/已发薪/已发条透传）
	batchCtx, err := loadBatchCtx(l.ctx, l.svcCtx, rows)
	if err != nil {
		return nil, err
	}

	records := make([]types.LedgerRow, 0, len(rows))
	employeeCount := int64(0)
	earning, deduction, net, labor := 0.0, 0.0, 0.0, 0.0
	for _, row := range all {
		bc := batchCtx[row.BatchUuid]
		abnormal := ""
		if bc != nil && bc.warnings != "" {
			abnormal = bc.warnings
		}
		// 仅看异常：批次告警透传为行级异常（无告警的行过滤掉）
		if req.OnlyAbnormal && abnormal == "" {
			continue
		}
		employeeCount++
		earning += row.EarningTotal
		deduction += row.DeductionTotal
		net += row.NetTotal
		labor += row.LaborCostTotal
	}
	// 明细行也走同一过滤（分页在过滤后的全集上切）
	filtered := make([]*hrm_salary_payslip.HrmSalaryPayslip, 0, len(all))
	for _, row := range all {
		bc := batchCtx[row.BatchUuid]
		if req.OnlyAbnormal && (bc == nil || bc.warnings == "") {
			continue
		}
		filtered = append(filtered, row)
	}
	start := int((req.Current - 1) * req.Size)
	if start < 0 || start > len(filtered) {
		start = 0
	}
	end := start + int(req.Size)
	if end > len(filtered) || int(req.Size) == 0 {
		if int(req.Size) == 0 {
			end = len(filtered)
		} else {
			end = len(filtered)
		}
	}
	if start > end {
		start = end
	}
	for _, row := range filtered[min(start, len(filtered)):min(end, len(filtered))] {
		bc := batchCtx[row.BatchUuid]
		var paid, payslipped int64
		abnormal := ""
		if bc != nil {
			paid, payslipped = bc.paid, bc.payslipped
			abnormal = bc.warnings
		}
		records = append(records, types.LedgerRow{
			BatchUuid:      row.BatchUuid,
			BatchNo:        row.BatchNo,
			BatchName:      row.BatchName,
			YearMonth:      yearMonthOf(row.BatchYear, row.BatchMonth),
			EmployeeUuid:   row.EmployeeUuid,
			EmployeeNo:     row.EmployeeNo,
			EmployeeName:   row.EmployeeName,
			OrgName:        row.OrgName,
			EarningTotal:   row.EarningTotal,
			DeductionTotal: row.DeductionTotal,
			NetTotal:       row.NetTotal,
			LaborCostTotal: row.LaborCostTotal,
			Abnormal:       abnormal,
			Paid:           paid,
			Payslipped:     payslipped,
		})
	}

	return &types.LedgerResponse{
		EmployeeCount:  employeeCount,
		EarningTotal:   round2(earning),
		DeductionTotal: round2(deduction),
		NetTotal:       round2(net),
		LaborCostTotal: round2(labor),
		Records:        records,
	}, nil
}

// batchCtx 单批次的透传字段快照。
type batchCtxRow struct {
	warnings   string
	paid       int64
	payslipped int64
}

// loadBatchCtx 一次性加载涉及的批次行（异常文本 + pay_status/payslip_status）。
func loadBatchCtx(ctx context.Context, svcCtx *svc.ServiceContext, rows []*hrm_salary_payslip.HrmSalaryPayslip) (map[string]*batchCtxRow, error) {
	out := make(map[string]*batchCtxRow, 8)
	for _, row := range rows {
		if _, ok := out[row.BatchUuid]; ok {
			continue
		}
		batch, err := svcCtx.Model.HrmSalaryBatch.FindOneByUuid(ctx, nil, row.BatchUuid)
		if err != nil {
			continue // 批次被删的孤儿工资条：透传字段留空
		}
		bc := &batchCtxRow{paid: 0, payslipped: 0}
		if batch.PayStatus == "1" {
			bc.paid = 1
		}
		if batch.PayslipStatus == "1" {
			bc.payslipped = 1
		}
		if batch.Warnings.Valid {
			bc.warnings = strings.TrimSpace(batch.Warnings.String)
		}
		out[row.BatchUuid] = bc
	}
	return out, nil
}
