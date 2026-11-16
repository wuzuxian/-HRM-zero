package monthreport

// HRM-23 考勤月报聚合域（demo-inventory 4.4 考勤月报）。
//
// 口径（与 HRM-22 考勤统计一致，Go 内存聚合照抄，规避 sqlite/mysql/pgx 方言聚合差异）：
//   - 应出勤天 = 月内工作日数（三级回退生效规则 attendancegroup.ResolveEffectiveRule 的
//     WorkDays，1=周一..7=周日；无规则兜底周一至周五）。
//   - 实出勤天 = settle_status ∈ {normal, late, early}（迟到/早退也算出勤，同时计异常）。
//   - 迟到/早退/缺卡次数 = settle_status 计数。
//   - 旷工天 = missing 且上/下班卡均无打卡（全天未出勤；单侧缺卡只计缺卡次数）——同 HRM-22。
//   - 工作日/休息日/节假日加班(h) = hrm_overtime_bill status=approved 单据按 day_type
//     （1=工作日 2=休息日 3=法定节假日）分桶的折算基准时长
//     （confirm_hours>0 取核定，否则申请时长 apply_hours）。
//   - 调休入账(h) = status=approved 单据 accrue_hours 合计（HRM-21 审批通过折算落定）。
//
// 锁定语义：
//   - 锁定月报 = 当月实时聚合结果写入 hrm_attendance_month_report 作为快照（row_hash
//     为行内容 hash）。
//   - 锁定月列表/导出/GetMonthReportRows 均取快照行。
//   - 锁定月「重新计算」= 重跑聚合与快照 diff，不一致行标 changed=1（只看锁定后变更）；
//     快照数值不覆盖（保锁定口径），changed 标记提示管理员复核后解锁。
//   - 未锁定月「重新计算」= 重跑聚合覆盖月报表行（缓存刷新），changed 清零。
//   - 解锁 = 删除该月快照行 + 清变更标记（月报恢复实时聚合）。
//     （月报表为可重建的聚合快照，删除即回实时口径；流水表 hrm_attendance_month_lock 留痕。）
//
// 消费契约（HRM-37 试算）:
//
//	GetMonthReportRows(ctx, svcCtx, yearMonth string, orgUuids []string) ([]MonthReportRow, error)
//	  - yearMonth: "yyyy-MM"（必填）。
//	  - orgUuids: 可为空；非空时仅返回部门（含下级语义由调用方展开，本函数按传入集合过滤）
//	    命中员工（employee.OrgUuid ∈ orgUuids）的月报行，按工号升序。
//	  - 锁定月返回锁定快照行；未锁定月实时聚合（unlocked=true 标记）。
//	  - 行字段见 MonthReportRow（含加班三桶 + 调休入账 + 异常合计；无 Changed 标记语义）。

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/attendancegroup"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/punch"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_attendance_month_report"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_overtime_bill"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_org"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_punch_record"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
)

// YearMonthLayout 月份格式。
const YearMonthLayout = "2006-01"

// DayTypeWorkday/Restday/Holiday 加班单当天类型（hrm_overtime_bill.day_type）。
const (
	DayTypeWorkday = "1"
	DayTypeRestday = "2"
	DayTypeHoliday = "3"
)

// BillStatusApproved 加班单已审批状态（hrm_bill 底座状态码）。
const BillStatusApproved = "approved"

// MonthReportRow 聚合后的单员工月报行（快照行/实时行共用；HRM-37 消费契约载体）。
type MonthReportRow struct {
	YearMonth    string  `json:"yearMonth"`
	EmployeeUuid string  `json:"employeeUuid"`
	EmployeeNo   string  `json:"employeeNo"`
	EmployeeName string  `json:"employeeName"`
	OrgUuid      string  `json:"orgUuid"`
	OrgName      string  `json:"orgName"`
	AttGroupName string  `json:"attGroupName"`
	ExpectedDays int64   `json:"expectedDays"`
	AttendDays   int64   `json:"attendDays"`
	LateCount    int64   `json:"lateCount"`
	EarlyCount   int64   `json:"earlyCount"`
	MissingCount int64   `json:"missingCount"`
	AbsentDays   int64   `json:"absentDays"`
	WorkdayOtH   float64 `json:"workdayOtH"`
	RestdayOtH   float64 `json:"restdayOtH"`
	HolidayOtH   float64 `json:"holidayOtH"`
	CompAccrueH  float64 `json:"compAccrueH"`
}

// monthRange 年月 → 起止日期（yyyy-MM-dd）。
func monthRange(yearMonth string) (string, string, error) {
	start, err := time.ParseInLocation(YearMonthLayout, yearMonth, time.Local)
	if err != nil {
		return "", "", fmt.Errorf("月份必须为 yyyy-MM: %s", yearMonth)
	}
	end := start.AddDate(0, 1, -1)
	return start.Format(punch.DateLayout), end.Format(punch.DateLayout), nil
}

// currentYearMonth 当月 yyyy-MM。
func currentYearMonth() string {
	return time.Now().Format(YearMonthLayout)
}

// validateYearMonth 月份格式校验（空→当月）。
func validateYearMonth(yearMonth string) (string, error) {
	if yearMonth == "" {
		return currentYearMonth(), nil
	}
	if _, err := time.ParseInLocation(YearMonthLayout, yearMonth, time.Local); err != nil {
		return "", fmt.Errorf("月份必须为 yyyy-MM: %s", yearMonth)
	}
	return yearMonth, nil
}

// isLocked 该月份是否已锁定（存在锁定快照行即视为锁定）。
func isLocked(ctx context.Context, svcCtx *svc.ServiceContext, yearMonth string) (bool, error) {
	row, err := svcCtx.Model.HrmAttendanceMonthReport.FindOneByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_attendance_month_report.YearMonth, yearMonth).Build()...)
	if err != nil {
		return false, nil // ErrNotFound 或查询错误 → 未锁定（快照查询阶段再如实上抛）
	}
	return row != nil, nil
}

// listEmployees 月报员工集合（全体有部门员工；employeeUuids 非空时限定）。
func listEmployees(ctx context.Context, svcCtx *svc.ServiceContext, employeeUuids []string) ([]*hrm_employee.HrmEmployee, error) {
	chain := condition.NewChain().NotEqual(hrm_employee.OrgUuid, "")
	if len(employeeUuids) > 0 {
		chain = chain.In(hrm_employee.Uuid, employeeUuids)
	}
	return svcCtx.Model.HrmEmployee.FindByCondition(ctx, nil, chain.Build()...)
}

// listMonthEmployees 月报员工集合：orgUuids（已展开含下级）/employeeUuids 双重过滤。
func listMonthEmployees(ctx context.Context, svcCtx *svc.ServiceContext, orgUuids, employeeUuids []string) ([]*hrm_employee.HrmEmployee, error) {
	chain := condition.NewChain().NotEqual(hrm_employee.OrgUuid, "")
	if len(orgUuids) > 0 {
		chain = chain.In(hrm_employee.OrgUuid, orgUuids)
	}
	if len(employeeUuids) > 0 {
		chain = chain.In(hrm_employee.Uuid, employeeUuids)
	}
	return svcCtx.Model.HrmEmployee.FindByCondition(ctx, nil, chain.Build()...)
}

// personMonthAgg 单员工月内聚合（照抄 HRM-22 内存聚合）。
type personMonthAgg struct {
	statusCounts map[string]int64
	absentDays   int64
}

func newPersonMonthAgg() *personMonthAgg {
	return &personMonthAgg{statusCounts: map[string]int64{}}
}

// aggregateRows 实时聚合一整个月的月报行（按员工工号升序）。
func aggregateRows(ctx context.Context, svcCtx *svc.ServiceContext, yearMonth string,
	employees []*hrm_employee.HrmEmployee) ([]MonthReportRow, error) {
	if len(employees) == 0 {
		return []MonthReportRow{}, nil
	}
	startDate, endDate, err := monthRange(yearMonth)
	if err != nil {
		return nil, err
	}

	// 打卡行（settle_* 快照列）。
	employeeUuids := make([]string, 0, len(employees))
	for _, e := range employees {
		employeeUuids = append(employeeUuids, e.Uuid)
	}
	punchRows, err := svcCtx.Model.HrmPunchRecord.FindByCondition(ctx, nil,
		condition.NewChain().
			GreaterEqualThan(hrm_punch_record.PunchDate, startDate).
			LessEqualThan(hrm_punch_record.PunchDate, endDate).
			In(hrm_punch_record.EmployeeUuid, employeeUuids).Build()...)
	if err != nil {
		return nil, err
	}
	aggs := map[string]*personMonthAgg{}
	for _, row := range punchRows {
		agg, ok := aggs[row.EmployeeUuid]
		if !ok {
			agg = newPersonMonthAgg()
			aggs[row.EmployeeUuid] = agg
		}
		agg.statusCounts[row.SettleStatus]++
		if row.SettleStatus == punch.SettleMissing && !row.ClockInTime.Valid && !row.ClockOutTime.Valid {
			agg.absentDays++
		}
	}

	// 已审批加班单（按员工分桶）。
	overtimeRows, err := svcCtx.Model.HrmOvertimeBill.FindByCondition(ctx, nil,
		condition.NewChain().
			Equal(hrm_overtime_bill.Status, BillStatusApproved).
			GreaterEqualThan(hrm_overtime_bill.OvertimeDate, startDate).
			LessEqualThan(hrm_overtime_bill.OvertimeDate, endDate).
			In(hrm_overtime_bill.EmployeeUuid, employeeUuids).Build()...)
	if err != nil {
		return nil, err
	}
	otAggs := map[string]*overtimeAgg{}
	for _, row := range overtimeRows {
		agg, ok := otAggs[row.EmployeeUuid]
		if !ok {
			agg = newOvertimeAgg()
			otAggs[row.EmployeeUuid] = agg
		}
		agg.add(row.DayType, row.ConfirmHours, row.ApplyHours, row.AccrueHours)
	}

	rows := make([]MonthReportRow, 0, len(employees))
	for _, employee := range employees {
		rows = append(rows, buildMonthRow(ctx, svcCtx, yearMonth, employee, aggs[employee.Uuid], otAggs[employee.Uuid]))
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].EmployeeNo < rows[j].EmployeeNo })
	return rows, nil
}

// overtimeAgg 单员工加班单聚合（三桶 + 调休入账）。
type overtimeAgg struct {
	workdayH float64
	restdayH float64
	holidayH float64
	accrueH  float64
}

func newOvertimeAgg() *overtimeAgg { return &overtimeAgg{} }

func (a *overtimeAgg) add(dayType string, confirmHours, applyHours, accrueHours float64) {
	// 折算基准：核定时长>0 取核定，否则申请时长（HRM-21 折算口径）。
	base := applyHours
	if confirmHours > 0 {
		base = confirmHours
	}
	switch dayType {
	case DayTypeWorkday:
		a.workdayH += base
	case DayTypeRestday:
		a.restdayH += base
	case DayTypeHoliday:
		a.holidayH += base
	}
	a.accrueH += accrueHours
}

// buildMonthRow 单员工月报行（应出勤天按生效规则 WorkDays 逐日判定，照抄 HRM-22）。
func buildMonthRow(ctx context.Context, svcCtx *svc.ServiceContext, yearMonth string,
	employee *hrm_employee.HrmEmployee, agg *personMonthAgg, otAgg *overtimeAgg) MonthReportRow {
	row := MonthReportRow{
		YearMonth:    yearMonth,
		EmployeeUuid: employee.Uuid,
		EmployeeNo:   employee.EmployeeNo,
		EmployeeName: employee.Name,
		OrgUuid:      employee.OrgUuid,
		OrgName:      orgName(ctx, svcCtx, employee.OrgUuid),
		AttGroupName: attGroupName(ctx, svcCtx, employee),
	}
	row.ExpectedDays = expectedDaysInMonth(ctx, svcCtx, employee, yearMonth)
	if agg != nil {
		row.AttendDays = agg.statusCounts[punch.SettleNormal] + agg.statusCounts[punch.SettleLate] + agg.statusCounts[punch.SettleEarly]
		row.LateCount = agg.statusCounts[punch.SettleLate]
		row.EarlyCount = agg.statusCounts[punch.SettleEarly]
		row.MissingCount = agg.statusCounts[punch.SettleMissing]
		row.AbsentDays = agg.absentDays
	}
	if otAgg != nil {
		row.WorkdayOtH = round2(otAgg.workdayH)
		row.RestdayOtH = round2(otAgg.restdayH)
		row.HolidayOtH = round2(otAgg.holidayH)
		row.CompAccrueH = round2(otAgg.accrueH)
	}
	return row
}

// expectedDaysInMonth 月内应出勤天（WorkDays 口径；无规则兜底周一至周五）。
func expectedDaysInMonth(ctx context.Context, svcCtx *svc.ServiceContext, employee *hrm_employee.HrmEmployee, yearMonth string) int64 {
	workDays := map[int64]bool{1: true, 2: true, 3: true, 4: true, 5: true}
	rule, err := attendancegroup.ResolveEffectiveRule(ctx, svcCtx, employee)
	if err != nil {
		logx.WithContext(ctx).Errorf("resolve rule for %s: %v", employee.Uuid, err)
	} else if rule != nil && rule.Found {
		workDays = map[int64]bool{}
		for _, d := range rule.WorkDays {
			workDays[d] = true
		}
	}
	count := int64(0)
	for date := firstDayOfMonth(yearMonth); ; {
		if workDays[WeekdayOfDay(date)] {
			count++
		}
		if date == lastDayOfMonth(yearMonth) {
			break
		}
		next, err := addOneDay(date)
		if err != nil {
			break
		}
		date = next
	}
	return count
}

// firstDayOfMonth / lastDayOfMonth yyyy-MM → yyyy-MM-dd。
func firstDayOfMonth(yearMonth string) string {
	return yearMonth + "-01"
}

func lastDayOfMonth(yearMonth string) string {
	t, err := time.ParseInLocation(YearMonthLayout, yearMonth, time.Local)
	if err != nil {
		return yearMonth + "-31"
	}
	return t.AddDate(0, 1, -1).Format(punch.DateLayout)
}

// WeekdayOfDay yyyy-MM-dd → 考勤日（1=周一..7=周日；非法日期返回 0，视为非工作日）。
// 与 punch.settle_domain.weekdayOf 同口径（该函数未导出，月报域自持一份）。
func WeekdayOfDay(date string) int64 {
	t, err := time.ParseInLocation(punch.DateLayout, date, time.Local)
	if err != nil {
		return 0
	}
	wd := int64(t.Weekday()) // Go: 周日=0
	if wd == 0 {
		return 7
	}
	return wd
}

// addOneDay yyyy-MM-dd + 1 天。
func addOneDay(date string) (string, error) {
	t, err := time.ParseInLocation(punch.DateLayout, date, time.Local)
	if err != nil {
		return "", err
	}
	return t.AddDate(0, 0, 1).Format(punch.DateLayout), nil
}

// orgName 部门名（空/查无返回空串）。
func orgName(ctx context.Context, svcCtx *svc.ServiceContext, orgUuid string) string {
	if orgUuid == "" {
		return ""
	}
	org, err := svcCtx.Model.HrmOrg.FindOneByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_org.Uuid, orgUuid).Build()...)
	if err != nil || org == nil {
		return ""
	}
	return org.OrgName
}

// attGroupName 考勤组名（生效规则来源名；组织默认规则/无规则返回空）。
func attGroupName(ctx context.Context, svcCtx *svc.ServiceContext, employee *hrm_employee.HrmEmployee) string {
	rule, err := attendancegroup.ResolveEffectiveRule(ctx, svcCtx, employee)
	if err != nil || rule == nil || !rule.Found {
		return ""
	}
	if rule.Source == attendancegroup.RuleSourceGroup {
		return rule.SourceName
	}
	return ""
}

func round2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}

// rowHashOf 月报行内容 hash（锁定快照 diff 依据；不包含 changed/uuid 等元数据）。
func rowHashOf(row MonthReportRow) string {
	h := sha256.New()
	fmt.Fprintf(h, "%s|%s|%s|%s|%s|%s|%s|%d|%d|%d|%d|%d|%d|%.2f|%.2f|%.2f|%.2f",
		row.YearMonth, row.EmployeeUuid, row.EmployeeNo, row.EmployeeName,
		row.OrgUuid, row.OrgName, row.AttGroupName,
		row.ExpectedDays, row.AttendDays, row.LateCount, row.EarlyCount, row.MissingCount, row.AbsentDays,
		row.WorkdayOtH, row.RestdayOtH, row.HolidayOtH, row.CompAccrueH)
	return hex.EncodeToString(h.Sum(nil))
}

// GetMonthReportRows HRM-37 试算消费契约：指定月份的按人月报行。
//   - 锁定月 → 锁定快照行（hmm 不带 Changed 语义）；
//   - 未锁定月 → 实时聚合（unlocked=true 由调用方按 isMonthLocked 自行判断）。
//
// orgUuids 非空时仅返回 employee.OrgUuid ∈ orgUuids 的行（含下级语义由调用方展开）。
// 行按工号升序。
func GetMonthReportRows(ctx context.Context, svcCtx *svc.ServiceContext, yearMonth string, orgUuids []string) ([]MonthReportRow, bool, error) {
	if _, err := validateYearMonth(yearMonth); err != nil {
		return nil, false, err
	}
	locked, err := isLocked(ctx, svcCtx, yearMonth)
	if err != nil {
		return nil, false, err
	}
	if locked {
		rows, err := snapshotRows(ctx, svcCtx, yearMonth, orgUuids)
		return rows, true, err
	}
	employees, err := listMonthEmployees(ctx, svcCtx, orgUuids, nil)
	if err != nil {
		return nil, false, err
	}
	rows, err := aggregateRows(ctx, svcCtx, yearMonth, employees)
	return rows, false, err
}

// snapshotRows 月报表快照行 → MonthReportRow（按工号升序；orgUuids 过滤）。
func snapshotRows(ctx context.Context, svcCtx *svc.ServiceContext, yearMonth string, orgUuids []string) ([]MonthReportRow, error) {
	chain := condition.NewChain().Equal(hrm_attendance_month_report.YearMonth, yearMonth)
	if len(orgUuids) > 0 {
		chain = chain.In(hrm_attendance_month_report.OrgUuid, orgUuids)
	}
	snapshots, err := svcCtx.Model.HrmAttendanceMonthReport.FindByCondition(ctx, nil,
		chain.OrderByAsc(hrm_attendance_month_report.EmployeeNo).Build()...)
	if err != nil {
		return nil, err
	}
	rows := make([]MonthReportRow, 0, len(snapshots))
	for _, s := range snapshots {
		rows = append(rows, MonthReportRow{
			YearMonth:    s.YearMonth,
			EmployeeUuid: s.EmployeeUuid,
			EmployeeNo:   s.EmployeeNo,
			EmployeeName: s.EmployeeName,
			OrgUuid:      s.OrgUuid,
			OrgName:      s.OrgName,
			AttGroupName: s.AttGroupName,
			ExpectedDays: s.ExpectedDays,
			AttendDays:   s.AttendDays,
			LateCount:    s.LateCount,
			EarlyCount:   s.EarlyCount,
			MissingCount: s.MissingCount,
			AbsentDays:   s.AbsentDays,
			WorkdayOtH:   s.WorkdayOvertimeHours,
			RestdayOtH:   s.RestdayOvertimeHours,
			HolidayOtH:   s.HolidayOvertimeHours,
			CompAccrueH:  s.CompAccrueHours,
		})
	}
	return rows, nil
}

// snapshotRow 月报表快照行 + changed 标记。
type snapshotRow struct {
	row     MonthReportRow
	changed bool
	hash    string
}

// snapshotRowsForList 快照行读取（列表/导出用；带 changed/uuid 元数据）。
func snapshotRowsForList(ctx context.Context, svcCtx *svc.ServiceContext, yearMonth string,
	orgUuids, employeeUuids []string) ([]snapshotRow, error) {
	chain := condition.NewChain().Equal(hrm_attendance_month_report.YearMonth, yearMonth)
	if len(orgUuids) > 0 {
		chain = chain.In(hrm_attendance_month_report.OrgUuid, orgUuids)
	}
	if len(employeeUuids) > 0 {
		chain = chain.In(hrm_attendance_month_report.EmployeeUuid, employeeUuids)
	}
	snapshots, err := svcCtx.Model.HrmAttendanceMonthReport.FindByCondition(ctx, nil,
		chain.OrderByAsc(hrm_attendance_month_report.EmployeeNo).Build()...)
	if err != nil {
		return nil, err
	}
	rows := make([]snapshotRow, 0, len(snapshots))
	for _, s := range snapshots {
		rows = append(rows, snapshotRow{
			row: MonthReportRow{
				YearMonth:    s.YearMonth,
				EmployeeUuid: s.EmployeeUuid,
				EmployeeNo:   s.EmployeeNo,
				EmployeeName: s.EmployeeName,
				OrgUuid:      s.OrgUuid,
				OrgName:      s.OrgName,
				AttGroupName: s.AttGroupName,
				ExpectedDays: s.ExpectedDays,
				AttendDays:   s.AttendDays,
				LateCount:    s.LateCount,
				EarlyCount:   s.EarlyCount,
				MissingCount: s.MissingCount,
				AbsentDays:   s.AbsentDays,
				WorkdayOtH:   s.WorkdayOvertimeHours,
				RestdayOtH:   s.RestdayOvertimeHours,
				HolidayOtH:   s.HolidayOvertimeHours,
				CompAccrueH:  s.CompAccrueHours,
			},
			changed: s.Changed == 1,
			hash:    s.RowHash,
		})
	}
	return rows, nil
}

// DefaultLockDay 默认锁定日（次月 5 号锁定当月月报）——Dashboard buildLockHint 的
// 常量口径（HRM-17 notes：落地配置表后替换此常量即可）。
const DefaultLockDay = 5

// lockHint 锁定日提示：「N月考勤将于次月MM-05锁定」（HRM-17 buildLockHint 同款文案）。
func lockHint(yearMonth string) string {
	t, err := time.ParseInLocation(YearMonthLayout, yearMonth, time.Local)
	if err != nil {
		return ""
	}
	next := t.AddDate(0, 1, 0)
	return fmt.Sprintf("%d月考勤将于次月%s-%02d锁定", t.Month(), next.Format("01"), DefaultLockDay)
}

// employeeUuidsByName 姓名（模糊）→ 员工 uuid 集合。
func employeeUuidsByName(ctx context.Context, svcCtx *svc.ServiceContext, name string) ([]string, error) {
	rows, err := svcCtx.Model.HrmEmployee.FindByCondition(ctx, nil,
		condition.NewChain().Like(hrm_employee.Name, "%"+name+"%").Build()...)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.Uuid)
	}
	return out, nil
}

// subOrgUuids 部门含下级（orgUuid 自身 + 全部后代，与 attendancegroup/leavebalance 同款展开）。
func subOrgUuids(ctx context.Context, svcCtx *svc.ServiceContext, orgUuid string) ([]string, error) {
	all, err := svcCtx.Model.HrmOrg.FindByCondition(ctx, nil,
		condition.NewChain().OrderByAsc(hrm_org.Id).Build()...)
	if err != nil {
		return nil, err
	}
	childrenOf := make(map[string][]string, len(all))
	for _, row := range all {
		childrenOf[row.ParentUuid] = append(childrenOf[row.ParentUuid], row.Uuid)
	}
	out := []string{orgUuid}
	stack := []string{orgUuid}
	for len(stack) > 0 {
		cur := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		for _, child := range childrenOf[cur] {
			out = append(out, child)
			stack = append(stack, child)
		}
	}
	return out, nil
}

// pageBounds 内存分页边界（Current/Size 防御性钳制，与 punch 同款）。
func pageBounds(current, size, total int64) (int64, int64) {
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 10
	}
	start := (current - 1) * size
	if start > total {
		start = total
	}
	end := start + size
	if end > total {
		end = total
	}
	return start, end
}

// IsMonthLocked 月份锁定状态查询（HRM-37 等消费方判断快照/实时口径用）。
func IsMonthLocked(ctx context.Context, svcCtx *svc.ServiceContext, yearMonth string) (bool, error) {
	return isLocked(ctx, svcCtx, yearMonth)
}
