// Package salary_payslip 工资条域共享逻辑（HRM-39，GLOSSARY「工资条」）。
package salary_payslip

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_batch"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_batch_item"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_payslip"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_payslip"
)

// round2 保留 2 位小数（金额精度口径，与 salary_batch/taxcalc/perf 同款）。
func round2(v float64) float64 {
	return float64(int64(v*100+0.5)) / 100
}

// uuidOf 新 v4 uuid（工资条行主键快照）。
func uuidOf() string {
	return uuid.New().String()
}

// nowStr 首次已读时间（read_time datetime 列）。
func nowStr() string {
	return time.Now().Format(time.DateTime)
}

// splitYearMonth 解析 'YYYY-MM' → (year, month, ok)。
func splitYearMonth(s string) (int64, int64, bool) {
	if len(s) != 7 || s[4] != '-' {
		return 0, 0, false
	}
	var y, m int
	if _, err := fmt.Sscanf(s, "%d-%d", &y, &m); err != nil || y < 2000 || y > 2999 || m < 1 || m > 12 {
		return 0, 0, false
	}
	return int64(y), int64(m), true
}

// yearMonthOf 年月格式化 'YYYY-MM'。
func yearMonthOf(year, month int64) string {
	return fmt.Sprintf("%04d-%02d", year, month)
}

// min 取小值（台账分页切片用）。
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// statusNotFound / statusInvalid / statusNotConfirmed / statusNoResult
// 业务错误码便捷包装（10695-10698，见 errcodes/hrm_salary_payslip.go）。
func statusNotFound(msg string) error {
	return status.ErrorMessage(errcodes.SalaryPayslipNotFoundCode, msg)
}

func statusInvalid(msg string) error {
	return status.ErrorMessage(errcodes.SalaryPayslipInvalidCode, msg)
}

func statusNotConfirmed(msg string) error {
	return status.ErrorMessage(errcodes.SalaryPayslipBatchNotConfirmedCode, msg)
}

func statusNoResult(msg string) error {
	return status.ErrorMessage(errcodes.SalaryPayslipBatchNoResultCode, msg)
}

// findBatchByUuid 按 uuid 取批次行（不存在 → SalaryPayslipNotFoundCode）。
func findBatchByUuid(ctx context.Context, m model.Model, batchUuid string) (*hrm_salary_batch.HrmSalaryBatch, error) {
	if batchUuid == "" {
		return nil, statusInvalid("batch_uuid 不能为空")
	}
	row, err := m.HrmSalaryBatch.FindOneByUuid(ctx, nil, batchUuid)
	if err != nil {
		if err == hrm_salary_batch.ErrNotFound {
			return nil, statusNotFound("薪资批次不存在")
		}
		return nil, err
	}
	return row, nil
}

// payslipItem 工资条明细 JSON 元素。
type payslipItem struct {
	ItemCode     string  `json:"itemCode"`
	ItemName     string  `json:"itemName"`
	ItemClass    string  `json:"itemClass"`
	Amount       float64 `json:"amount"`
	Source       string  `json:"source"`
	SourceDetail string  `json:"sourceDetail"`
}

// generateForBatch 在事务内按批次试算明细行聚合生成每员工工资条
// （UNIQUE(batch_uuid,employee_uuid)：先 DeleteByCondition 再逐行 Insert，
// 重发布幂等覆盖；明细按薪资项目 JSON 快照，calc_order ASC）。
// 返回生成的工资条张数。
func generateForBatch(ctx context.Context, session sqlx.Session, m model.Model, batch *hrm_salary_batch.HrmSalaryBatch) (int64, error) {
	items, err := m.HrmSalaryBatchItem.FindByCondition(ctx, session,
		condition.NewChain().
			Equal(hrm_salary_batch_item.BatchUuid, batch.Uuid).
			OrderByAsc(hrm_salary_batch_item.CalcOrder).
			OrderByAsc(hrm_salary_batch_item.Id).
			Build()...)
	if err != nil {
		return 0, err
	}
	if len(items) == 0 {
		return 0, statusNoResult("批次尚无试算结果，无法生成工资条")
	}

	yearMonth := fmt.Sprintf("%04d-%02d", batch.BatchYear, batch.BatchMonth)

	// 按员工聚合
	type agg struct {
		employeeUuid, employeeNo, employeeName, orgName string
		earning, deduction, laborCost                   float64
		items                                           []payslipItem
	}
	order := make([]string, 0, 8)
	byEmp := make(map[string]*agg, 8)
	for _, it := range items {
		a, ok := byEmp[it.EmployeeUuid]
		if !ok {
			a = &agg{
				employeeUuid: it.EmployeeUuid, employeeNo: it.EmployeeNo,
				employeeName: it.EmployeeName, orgName: it.OrgName,
			}
			byEmp[it.EmployeeUuid] = a
			order = append(order, it.EmployeeUuid)
		}
		switch it.ItemClass {
		case "earning":
			a.earning += it.Amount
		case "deduction":
			a.deduction += it.Amount
		case "cost":
			a.laborCost += it.Amount
		}
		a.items = append(a.items, payslipItem{
			ItemCode: it.ItemCode, ItemName: it.ItemName, ItemClass: it.ItemClass,
			Amount: it.Amount, Source: it.Source, SourceDetail: it.SourceDetail,
		})
	}

	// 重发布幂等覆盖：清掉旧工资条（已读留痕随覆盖重置——重发布意味着金额重算，旧已读作废）
	if err := m.HrmSalaryPayslip.DeleteByCondition(ctx, session,
		condition.NewChain().Equal(hrm_salary_payslip.BatchUuid, batch.Uuid).Build()...); err != nil {
		return 0, err
	}

	count := int64(0)
	for _, empUuid := range order {
		a := byEmp[empUuid]
		itemsJson, err := json.Marshal(a.items)
		if err != nil {
			return 0, err
		}
		if err := m.HrmSalaryPayslip.InsertV2(ctx, session, &hrm_salary_payslip.HrmSalaryPayslip{
			Uuid:           uuidOf(),
			BatchUuid:      batch.Uuid,
			BatchNo:        batch.BatchNo,
			BatchName:      batch.BatchName,
			BatchYear:      batch.BatchYear,
			BatchMonth:     batch.BatchMonth,
			EmployeeUuid:   a.employeeUuid,
			EmployeeNo:     a.employeeNo,
			EmployeeName:   a.employeeName,
			OrgName:        a.orgName,
			Title:          fmt.Sprintf("%s %s 工资条", batch.BatchName, yearMonth),
			EarningTotal:   round2(a.earning),
			DeductionTotal: round2(a.deduction),
			NetTotal:       round2(a.earning - a.deduction),
			LaborCostTotal: round2(a.laborCost),
			ItemsJson:      string(itemsJson),
			Published:      1,
			ReadStatus:     0,
			ReadTime:       sql.NullTime{},
		}); err != nil {
			return 0, err
		}
		count++
	}

	// 批次工资条进度写回（HRM-37 契约：payslip_status 0=未生成 1=已生成）
	if err := m.HrmSalaryBatch.UpdateFieldsByCondition(ctx, session,
		map[string]any{"payslip_status": "1"},
		condition.NewChain().Equal(hrm_salary_batch.Id, batch.Id).Build()...); err != nil {
		return 0, err
	}
	return count, nil
}

// MarkRead 员工本人查看工资条记 已读+已读时间（幂等不覆盖首次已读时间）。
func MarkRead(ctx context.Context, conn sqlx.SqlConn, payslipUuid string) error {
	if payslipUuid == "" {
		return statusInvalid("uuid 不能为空")
	}
	m := hrm_salary_payslip.NewHrmSalaryPayslipModel(conn)
	row, err := m.FindOneByUuid(ctx, nil, payslipUuid)
	if err != nil {
		if err == hrm_salary_payslip.ErrNotFound {
			return statusNotFound("工资条不存在")
		}
		return err
	}
	if row.ReadStatus == 1 {
		return nil // 幂等：首次已读时间不覆盖
	}
	return m.UpdateFieldsByCondition(ctx, nil, map[string]any{
		"read_status": 1,
		"read_time":   nowStr(),
	}, condition.NewChain().Equal(hrm_salary_payslip.Id, row.Id).Build()...)
}

// payslipRowToRecord 工资条行 → API record（read_time NULL=未读空串；items_json 解析）。
func payslipRowToRecord(row *hrm_salary_payslip.HrmSalaryPayslip) types.HrmSalaryPayslip {
	readTime := ""
	if row.ReadTime.Valid {
		readTime = row.ReadTime.Time.Format("2006-01-02 15:04:05")
	}
	items := make([]types.HrmSalaryPayslipItem, 0, 8)
	_ = json.Unmarshal([]byte(row.ItemsJson), &items)
	return types.HrmSalaryPayslip{
		Uuid:           row.Uuid,
		BatchUuid:      row.BatchUuid,
		BatchNo:        row.BatchNo,
		BatchName:      row.BatchName,
		YearMonth:      fmt.Sprintf("%04d-%02d", row.BatchYear, row.BatchMonth),
		Title:          row.Title,
		EmployeeUuid:   row.EmployeeUuid,
		EmployeeNo:     row.EmployeeNo,
		EmployeeName:   row.EmployeeName,
		OrgName:        row.OrgName,
		EarningTotal:   row.EarningTotal,
		DeductionTotal: row.DeductionTotal,
		NetTotal:       row.NetTotal,
		Items:          items,
		Published:      row.Published,
		ReadStatus:     row.ReadStatus,
		ReadTime:       readTime,
		CreateTime:     row.CreateTime.Format("2006-01-02 15:04:05"),
		UpdateTime:     row.UpdateTime.Format("2006-01-02 15:04:05"),
	}
}
