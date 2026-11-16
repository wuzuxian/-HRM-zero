package punch

// HRM-22 考勤统计聚合口径 logic 测试（svctest sqlite）：
//   构造多员工多日打卡行 → 聚合断言各统计卡与按人明细；
//   仅看异常过滤；异常摘要文本；应出勤天按工作日口径（规则 WorkDays）。

import (
	"context"
	"testing"

	"github.com/jzero-io/jzero/core/stores/condition"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_punch_record"
	employeelogic "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/employee"
	employeetypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/employee"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/punch"

	"github.com/jzero-io/jzero-admin/server/internal/svc"
)

// seedStatsEmployee 新增员工到指定组织（走组织默认规则）。
func seedStatsEmployee(t *testing.T, svcCtx *svc.ServiceContext, name, employeeNo, orgUuid string) string {
	t.Helper()
	req := &employeetypes.AddRequest{Name: name, EmployeeNo: employeeNo, OrgUuid: orgUuid, PersonStatus: "1"}
	if _, err := employeelogic.NewAdd(context.Background(), svcCtx, nil).Add(req); err != nil {
		t.Fatalf("add employee %s: %v", name, err)
	}
	row, err := svcCtx.Model.HrmEmployee.FindOneByCondition(context.Background(), nil,
		condition.NewChain().Equal(hrm_employee.EmployeeNo, employeeNo).Build()...)
	if err != nil {
		t.Fatalf("find employee %s: %v", employeeNo, err)
	}
	return row.Uuid
}

// seedSettledRow 直接落一行已日结的打卡记录（测试造数：settle_* 列显式指定，
// 模拟 SettleDay 落库结果，不依赖规则重算）。
func seedSettledRow(t *testing.T, svcCtx *svc.ServiceContext, employeeUuid, employeeNo, name, date, settleStatus string, lateMinutes, earlyMinutes int64, inTime, outTime string) {
	t.Helper()
	row := &hrm_punch_record.HrmPunchRecord{
		Uuid:               newUuid(),
		PunchDate:          date,
		EmployeeUuid:       employeeUuid,
		EmployeeNo:         employeeNo,
		EmployeeName:       name,
		Source:             SourceApp,
		SettleStatus:       settleStatus,
		SettleLateMinutes:  lateMinutes,
		SettleEarlyMinutes: earlyMinutes,
		SettleWorkHours:    8.0,
	}
	if inTime != "" {
		row.ClockInTime = mustTime(inTime)
		row.ClockInStatus = StatusNormal
	}
	if outTime != "" {
		row.ClockOutTime = mustTime(outTime)
		row.ClockOutStatus = StatusNormal
	}
	if _, err := svcCtx.Model.HrmPunchRecord.Insert(context.Background(), nil, row); err != nil {
		t.Fatalf("seed settled row %s/%s: %v", employeeNo, date, err)
	}
}

// TestStatisticsCardsAndPerPerson 统计卡 + 按人明细聚合口径。
// 2026-11-16(周一)/11-17(周二)/11-18(周三)/11-21(周六,休)：
//   张三: 11-16 late / 11-17 late / 11-18 normal / 11-21 rest → 应出勤 3, 出勤 3, 迟到 2
//   李四: 11-16 missing(无卡) / 11-17 normal / 11-18 missing(有上班卡,缺下班卡) → 应出勤 3, 出勤 1, 缺勤 1, 缺卡 2
func TestStatisticsCardsAndPerPerson(t *testing.T) {
	svcCtx := newSvcCtx(t)
	ctx := context.Background()
	s := seedOrgAndEmployee(t, svcCtx)
	saveOrgConfig(t, svcCtx, s.orgTech, "09:00", "18:00", []int64{1, 2, 3, 4, 5})
	zhang := s.empZhang
	li := seedStatsEmployee(t, svcCtx, "李四", "E0002", s.orgTech)

	seedSettledRow(t, svcCtx, zhang, "E0001", "张三", "2026-11-16", SettleLate, 30, 0, "2026-11-16 09:30:00", "2026-11-16 18:00:00")
	seedSettledRow(t, svcCtx, zhang, "E0001", "张三", "2026-11-17", SettleLate, 15, 0, "2026-11-17 09:15:00", "2026-11-17 18:00:00")
	seedSettledRow(t, svcCtx, zhang, "E0001", "张三", "2026-11-18", SettleNormal, 0, 0, "2026-11-18 09:00:00", "2026-11-18 18:00:00")
	seedSettledRow(t, svcCtx, zhang, "E0001", "张三", "2026-11-21", SettleRest, 0, 0, "", "")

	seedSettledRow(t, svcCtx, li, "E0002", "李四", "2026-11-16", SettleMissing, 0, 0, "", "")
	seedSettledRow(t, svcCtx, li, "E0002", "李四", "2026-11-17", SettleNormal, 0, 0, "2026-11-17 09:00:00", "2026-11-17 18:00:00")
	seedSettledRow(t, svcCtx, li, "E0002", "李四", "2026-11-18", SettleMissing, 0, 0, "2026-11-18 09:00:00", "")

	resp, err := NewStatistics(ctx, svcCtx, nil).Statistics(&types.StatisticsRequest{
		PageRequest: types.PageRequest{Current: 1, Size: 10},
		StartDate:   "2026-11-16", EndDate: "2026-11-21",
	})
	if err != nil {
		t.Fatalf("statistics: %v", err)
	}

	// 统计卡：记录条数 7；迟到 2；早退 0；缺卡 2；统计人数 2；异常人数 2。
	cards := resp.Cards
	if cards.RecordCount != 7 || cards.LateCount != 2 || cards.EarlyCount != 0 || cards.MissingCount != 2 {
		t.Fatalf("cards = %+v, want record=7 late=2 early=0 missing=2", cards)
	}
	if cards.StatEmployees != 2 || cards.AbnormalEmployees != 2 {
		t.Fatalf("cards employees = %d/%d, want 2/2", cards.StatEmployees, cards.AbnormalEmployees)
	}

	// 按人明细（按工号升序: 张三 E0001 / 李四 E0002）。
	if len(resp.Records) != 2 {
		t.Fatalf("records = %d, want 2", len(resp.Records))
	}
	z := resp.Records[0]
	if z.EmployeeNo != "E0001" || z.EmployeeName != "张三" {
		t.Fatalf("first row = %s/%s", z.EmployeeNo, z.EmployeeName)
	}
	if z.ExpectedDays != 5 { // 周一至周五口径: 11-16(一)/17(二)/18(三)/19(四)/20(五)，21(六)不算
		t.Fatalf("zhang expectedDays = %d, want 5", z.ExpectedDays)
	}
	if z.AttendDays != 3 || z.AbsentDays != 0 || z.LateCount != 2 || z.EarlyCount != 0 || z.MissingCount != 0 {
		t.Fatalf("zhang person = %+v", z)
	}
	if z.AbnormalTotal != 2 || z.AbnormalSummary != "迟到2次" {
		t.Fatalf("zhang abnormal = %d/%q, want 2/迟到2次", z.AbnormalTotal, z.AbnormalSummary)
	}

	l4 := resp.Records[1]
	if l4.ExpectedDays != 5 || l4.AttendDays != 1 || l4.AbsentDays != 1 || l4.MissingCount != 2 || l4.LateCount != 0 {
		t.Fatalf("li person = %+v", l4)
	}
	if l4.AbnormalTotal != 2 || l4.AbnormalSummary != "缺卡2次" {
		t.Fatalf("li abnormal = %d/%q, want 2/缺卡2次", l4.AbnormalTotal, l4.AbnormalSummary)
	}
	if l4.TotalWorkHours != 24.0 {
		t.Fatalf("li workHours = %v, want 24.0", l4.TotalWorkHours)
	}
}

// TestStatisticsAbnormalSummaryMixed 混合异常摘要「迟到3次,缺卡1次」语义 + 早退。
func TestStatisticsAbnormalSummaryMixed(t *testing.T) {
	summary := abnormalSummary(3, 0, 1)
	if summary != "迟到3次,缺卡1次" {
		t.Fatalf("summary = %q", summary)
	}
	if abnormalSummary(0, 2, 0) != "早退2次" {
		t.Fatalf("summary early = %q", abnormalSummary(0, 2, 0))
	}
	if abnormalSummary(0, 0, 0) != "" {
		t.Fatalf("summary empty = %q", abnormalSummary(0, 0, 0))
	}
}

// TestStatisticsOnlyAbnormalFilter 仅看异常过滤：无异常员工被剔除。
func TestStatisticsOnlyAbnormalFilter(t *testing.T) {
	svcCtx := newSvcCtx(t)
	ctx := context.Background()
	s := seedOrgAndEmployee(t, svcCtx)
	saveOrgConfig(t, svcCtx, s.orgTech, "09:00", "18:00", []int64{1, 2, 3, 4, 5})
	zhang := s.empZhang
	wang := seedStatsEmployee(t, svcCtx, "王五", "E0003", s.orgTech)

	seedSettledRow(t, svcCtx, zhang, "E0001", "张三", "2026-11-16", SettleLate, 10, 0, "2026-11-16 09:10:00", "2026-11-16 18:00:00")
	seedSettledRow(t, svcCtx, wang, "E0003", "王五", "2026-11-16", SettleNormal, 0, 0, "2026-11-16 09:00:00", "2026-11-16 18:00:00")

	resp, err := NewStatistics(ctx, svcCtx, nil).Statistics(&types.StatisticsRequest{
		PageRequest: types.PageRequest{Current: 1, Size: 10},
		StartDate:   "2026-11-16", EndDate: "2026-11-16",
		OnlyAbnormal: "1",
	})
	if err != nil {
		t.Fatalf("statistics onlyAbnormal: %v", err)
	}
	if len(resp.Records) != 1 || resp.Records[0].EmployeeNo != "E0001" {
		t.Fatalf("onlyAbnormal records = %+v, want only E0001", resp.Records)
	}

	// 不过滤时两人都在。
	resp, err = NewStatistics(ctx, svcCtx, nil).Statistics(&types.StatisticsRequest{
		PageRequest: types.PageRequest{Current: 1, Size: 10},
		StartDate:   "2026-11-16", EndDate: "2026-11-16",
	})
	if err != nil {
		t.Fatalf("statistics all: %v", err)
	}
	if len(resp.Records) != 2 {
		t.Fatalf("all records = %d, want 2", len(resp.Records))
	}
}

// TestStatisticsNameFilterAndRange 姓名筛选 + 日期区间聚合边界。
func TestStatisticsNameFilterAndRange(t *testing.T) {
	svcCtx := newSvcCtx(t)
	ctx := context.Background()
	s := seedOrgAndEmployee(t, svcCtx)
	saveOrgConfig(t, svcCtx, s.orgTech, "09:00", "18:00", []int64{1, 2, 3, 4, 5})
	zhang := s.empZhang
	li := seedStatsEmployee(t, svcCtx, "李四", "E0002", s.orgTech)

	seedSettledRow(t, svcCtx, zhang, "E0001", "张三", "2026-11-16", SettleNormal, 0, 0, "2026-11-16 09:00:00", "2026-11-16 18:00:00")
	seedSettledRow(t, svcCtx, li, "E0002", "李四", "2026-11-20", SettleEarly, 0, 45, "2026-11-20 09:00:00", "2026-11-20 17:15:00")

	// 姓名筛选 → 只有张三。
	resp, err := NewStatistics(ctx, svcCtx, nil).Statistics(&types.StatisticsRequest{
		PageRequest: types.PageRequest{Current: 1, Size: 10},
		StartDate:   "2026-11-01", EndDate: "2026-11-30", Name: "张",
	})
	if err != nil {
		t.Fatalf("statistics name: %v", err)
	}
	if len(resp.Records) != 1 || resp.Records[0].EmployeeName != "张三" || resp.Cards.RecordCount != 1 {
		t.Fatalf("name filter = %+v cards %+v", resp.Records, resp.Cards)
	}

	// 区间收窄到 [11-17, 11-30] → 只剩李四早退。
	resp, err = NewStatistics(ctx, svcCtx, nil).Statistics(&types.StatisticsRequest{
		PageRequest: types.PageRequest{Current: 1, Size: 10},
		StartDate:   "2026-11-17", EndDate: "2026-11-30",
	})
	if err != nil {
		t.Fatalf("statistics range: %v", err)
	}
	if resp.Cards.EarlyCount != 1 || resp.Cards.RecordCount != 1 {
		t.Fatalf("range cards = %+v, want early=1 record=1", resp.Cards)
	}

	// 非法日期 → 10630。
	_, err = NewStatistics(ctx, svcCtx, nil).Statistics(&types.StatisticsRequest{
		PageRequest: types.PageRequest{Current: 1, Size: 10}, StartDate: "2026-13-99",
	})
	if err == nil {
		t.Fatalf("bad date should error 10630")
	}
}

// TestStatisticsDetail 查看明细：某员工区间逐日 settle 详情（复用 toRecord 语义）。
func TestStatisticsDetail(t *testing.T) {
	svcCtx := newSvcCtx(t)
	ctx := context.Background()
	s := seedOrgAndEmployee(t, svcCtx)
	zhang := s.empZhang

	seedSettledRow(t, svcCtx, zhang, "E0001", "张三", "2026-11-16", SettleLate, 10, 0, "2026-11-16 09:10:00", "2026-11-16 18:00:00")
	seedSettledRow(t, svcCtx, zhang, "E0001", "张三", "2026-11-17", SettleNormal, 0, 0, "2026-11-17 09:00:00", "2026-11-17 18:00:00")
	seedSettledRow(t, svcCtx, "OTHER", "E0009", "别人", "2026-11-16", SettleNormal, 0, 0, "2026-11-16 09:00:00", "2026-11-16 18:00:00")

	resp, err := NewStatisticsDetail(ctx, svcCtx, nil).StatisticsDetail(&types.StatisticsDetailRequest{
		EmployeeUuid: zhang, StartDate: "2026-11-16", EndDate: "2026-11-17",
	})
	if err != nil {
		t.Fatalf("detail: %v", err)
	}
	if len(resp.Records) != 2 {
		t.Fatalf("detail records = %d, want 2", len(resp.Records))
	}
	if resp.Records[0].PunchDate != "2026-11-16" || resp.Records[0].SettleStatus != SettleLate {
		t.Fatalf("detail[0] = %+v", resp.Records[0])
	}
	if resp.Records[0].ClockInStatus != StatusNormal {
		t.Fatalf("detail[0] clockInStatus = %q, want normal (settle missing 补语义只对空状态)", resp.Records[0].ClockInStatus)
	}

	// 员工必填 → 10630
	_, err = NewStatisticsDetail(ctx, svcCtx, nil).StatisticsDetail(&types.StatisticsDetailRequest{})
	if err == nil {
		t.Fatalf("empty employeeUuid should error 10630")
	}
}

// TestStatisticsExpectedDaysByRuleWorkDays 应出勤天按生效规则 WorkDays 口径：
// 考勤组 work_days=1,2,3,4,5,6（周一至周六）时，区间含周六（11-21）应出勤 4 天。
func TestStatisticsExpectedDaysByRuleWorkDays(t *testing.T) {
	svcCtx := newSvcCtx(t)
	ctx := context.Background()
	s := seedOrgAndEmployee(t, svcCtx)
	zhang := s.empZhang
	addGroup(t, svcCtx, "六天组", "09:00", "18:00", []int64{1, 2, 3, 4, 5, 6}, []string{zhang})

	seedSettledRow(t, svcCtx, zhang, "E0001", "张三", "2026-11-16", SettleNormal, 0, 0, "2026-11-16 09:00:00", "2026-11-16 18:00:00")

	resp, err := NewStatistics(ctx, svcCtx, nil).Statistics(&types.StatisticsRequest{
		PageRequest: types.PageRequest{Current: 1, Size: 10},
		StartDate:   "2026-11-16", EndDate: "2026-11-21",
	})
	if err != nil {
		t.Fatalf("statistics: %v", err)
	}
	if len(resp.Records) != 1 {
		t.Fatalf("records = %d, want 1", len(resp.Records))
	}
	if resp.Records[0].ExpectedDays != 6 { // 11-16..21：16/17/18/19/20/21 全为工作日（6 天工作制）
		t.Fatalf("expectedDays = %d, want 6 (6-day work week)", resp.Records[0].ExpectedDays)
	}
}
