package punch

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_punch_record"
	attendanceconfiglogic "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/attendanceconfig"
	attendancegrouplogic "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/attendancegroup"
	employeelogic "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/employee"
	orglogic "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/org"
	configtypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/attendanceconfig"
	grouptypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/attendancegroup"
	employeetypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/employee"
	orgtypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/org"
	coreSvc "github.com/jzero-io/jzero-admin/core-engine/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
)

func newSvcCtx(t *testing.T) *svc.ServiceContext {
	t.Helper()
	env := svctest.New(t)
	// 打卡/日结 logic 只依赖 Model；跨包调用 employee/org 事务需要 SqlxConn。
	return &svc.ServiceContext{
		ServiceContext: &coreSvc.ServiceContext{SqlxConn: env.SqlConn},
		Model:          env.Model,
	}
}

type seeded struct {
	orgA, orgTech string
	empZhang      string
}

// seedOrgAndEmployee 建公司A → 技术部 + 员工张三（无考勤组，走组织默认规则）。
func seedOrgAndEmployee(t *testing.T, svcCtx *svc.ServiceContext) seeded {
	t.Helper()
	ctx := context.Background()

	addOrg := func(parentUuid, name string) string {
		req := &orgtypes.AddRequest{OrgName: name, OrgType: "2", OrderNum: 1, Status: "1"}
		if parentUuid != "" {
			req.ParentUuid = parentUuid
		}
		if _, err := orglogic.NewAdd(ctx, svcCtx, nil).Add(req); err != nil {
			t.Fatalf("add org %s: %v", name, err)
		}
		rows, err := svcCtx.Model.HrmOrg.FindByCondition(ctx, nil)
		if err != nil {
			t.Fatalf("find org %s: %v", name, err)
		}
		for _, row := range rows {
			if row.OrgName == name {
				return row.Uuid
			}
		}
		t.Fatalf("org %s not found", name)
		return ""
	}

	s := seeded{}
	s.orgA = addOrg("", "公司A")
	s.orgTech = addOrg(s.orgA, "技术部")

	req := &employeetypes.AddRequest{Name: "张三", EmployeeNo: "E0001", OrgUuid: s.orgTech, PersonStatus: "1"}
	if _, err := employeelogic.NewAdd(ctx, svcCtx, nil).Add(req); err != nil {
		t.Fatalf("add employee: %v", err)
	}
	row, err := svcCtx.Model.HrmEmployee.FindOneByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_employee.EmployeeNo, "E0001").Build()...)
	if err != nil {
		t.Fatalf("find employee: %v", err)
	}
	s.empZhang = row.Uuid
	return s
}

// saveOrgConfig 保存组织默认规则（周一至周五 09:00-18:00，可覆盖 workDays）。
func saveOrgConfig(t *testing.T, svcCtx *svc.ServiceContext, orgUuid, startTime, endTime string, workDays []int64) {
	t.Helper()
	if _, err := attendanceconfiglogic.NewSave(context.Background(), svcCtx, nil).Save(&configtypes.SaveRequest{
		OrgUuid: orgUuid, StartTime: startTime, EndTime: endTime, WorkDays: workDays,
	}); err != nil {
		t.Fatalf("save config: %v", err)
	}
}

// addGroup 建启用考勤组（固定班制，挂指定人员）。
func addGroup(t *testing.T, svcCtx *svc.ServiceContext, name string, startTime, endTime string, workDays []int64, memberUuids []string) {
	t.Helper()
	addGroupFlexible(t, svcCtx, name, startTime, endTime, workDays, memberUuids, 0)
}

func addGroupFlexible(t *testing.T, svcCtx *svc.ServiceContext, name, startTime, endTime string, workDays []int64, memberUuids []string, flexibleMinutes int64) {
	t.Helper()
	req := &grouptypes.AddRequest{
		GroupName: name, GroupType: "1",
		WorkDays: workDays, StartTime: startTime, EndTime: endTime,
		FlexibleMinutes: flexibleMinutes,
		MemberUuids:     memberUuids, Managers: []string{"王经理"}, Status: "1",
	}
	if _, err := attendancegrouplogic.NewAdd(context.Background(), svcCtx, nil).Add(req); err != nil {
		t.Fatalf("add group %s: %v", name, err)
	}
}

// insertPunchRow 直接落一行打卡记录（测试造数用，不走打卡端点）。
func insertPunchRow(t *testing.T, svcCtx *svc.ServiceContext, date string, clockIn, clockOut time.Time, inStatus, outStatus string, overtime string) {
	t.Helper()
	row := &hrm_punch_record.HrmPunchRecord{
		Uuid:          newUuid(),
		PunchDate:     date,
		EmployeeUuid:  "E-TEST",
		ClockInStatus: inStatus,
		ClockOutStatus: outStatus,
		Overtime:      overtime,
		Source:        SourceApp,
	}
	if !clockIn.IsZero() {
		row.ClockInTime = nullTime(clockIn)
		row.EmployeeUuid = currentTestEmployeeUuid
	}
	if !clockOut.IsZero() {
		row.ClockOutTime = nullTime(clockOut)
	}
	if _, err := svcCtx.Model.HrmPunchRecord.Insert(context.Background(), nil, row); err != nil {
		t.Fatalf("insert punch row: %v", err)
	}
}

// nullTime 测试造数：sql.NullTime 简写。
func nullTime(t time.Time) sql.NullTime {
	return sql.NullTime{Time: t, Valid: true}
}

var currentTestEmployeeUuid string

func mustSettle(t *testing.T, svcCtx *svc.ServiceContext, employeeUuid, date string) DaySettleResult {
	t.Helper()
	res, err := SettleDay(context.Background(), svcCtx, employeeUuid, date)
	if err != nil {
		t.Fatalf("SettleDay(%s, %s): %v", employeeUuid, date, err)
	}
	return *res
}

// TestSettleDayWorkedExamples 日结 worked examples:
// 正常 / 迟到 / 早退 / 缺卡（上/下）/ 跨天班 / 休息日 / 无规则兜底 / 外勤按正常出勤。
func TestSettleDayWorkedExamples(t *testing.T) {
	svcCtx := newSvcCtx(t)
	ctx := context.Background()

	// 2026-11-16 是周一（工作日）、2026-11-15 周日。
	const workday = "2026-11-16"
	const restday = "2026-11-15"

	// 场景 A: 组织默认规则 09:00-18:00（无休息时段），张三在技术部。
	s := seedOrgAndEmployee(t, svcCtx)
	saveOrgConfig(t, svcCtx, s.orgTech, "09:00", "18:00", []int64{1, 2, 3, 4, 5})

	// A1 正常: 08:55 上班 18:05 下班 → normal, 实打工时 9h10m = 9.17
	res := mustSettleWithPunch(t, svcCtx, s.empZhang, workday,
		"2026-11-16 08:55:00", "2026-11-16 18:05:00", StatusNormal, StatusNormal, OvertimeNo)
	if res.Status != SettleNormal {
		t.Fatalf("A1 status = %q, want normal", res.Status)
	}
	if res.LateMinutes != 0 || res.EarlyMinutes != 0 {
		t.Fatalf("A1 late/early = %d/%d, want 0/0", res.LateMinutes, res.EarlyMinutes)
	}
	if res.WorkHours != 9.17 {
		t.Fatalf("A1 workHours = %v, want 9.17", res.WorkHours)
	}

	// A2 迟到: 09:21 上班 18:00 下班 → late(21), 工时 8.65
	res = mustSettleWithPunch(t, svcCtx, s.empZhang, workday,
		"2026-11-16 09:21:00", "2026-11-16 18:00:00", StatusNormal, StatusNormal, OvertimeNo)
	if res.Status != SettleLate {
		t.Fatalf("A2 status = %q, want late", res.Status)
	}
	if res.LateMinutes != 21 {
		t.Fatalf("A2 lateMinutes = %d, want 21", res.LateMinutes)
	}

	// A3 早退: 09:00 上班 17:10 下班 → early(50)
	res = mustSettleWithPunch(t, svcCtx, s.empZhang, workday,
		"2026-11-16 09:00:00", "2026-11-16 17:10:00", StatusNormal, StatusNormal, OvertimeNo)
	if res.Status != SettleEarly {
		t.Fatalf("A3 status = %q, want early", res.Status)
	}
	if res.EarlyMinutes != 50 {
		t.Fatalf("A3 earlyMinutes = %d, want 50", res.EarlyMinutes)
	}

	// A4 迟到且早退并存 → 迟到优先展示为异常状态（status=late, earlyMinutes 仍记录）
	res = mustSettleWithPunch(t, svcCtx, s.empZhang, workday,
		"2026-11-16 09:40:00", "2026-11-16 17:30:00", StatusNormal, StatusNormal, OvertimeNo)
	if res.Status != SettleLate || res.LateMinutes != 40 || res.EarlyMinutes != 30 {
		t.Fatalf("A4 = status %q late %d early %d, want late 40/30", res.Status, res.LateMinutes, res.EarlyMinutes)
	}

	// A5 缺上班卡: 只有下班卡 → missing
	res = mustSettleWithPunch(t, svcCtx, s.empZhang, workday,
		"", "2026-11-16 18:00:00", "", StatusNormal, OvertimeNo)
	if res.Status != SettleMissing {
		t.Fatalf("A5 status = %q, want missing", res.Status)
	}

	// A6 缺下班卡: 只有上班卡 → missing
	res = mustSettleWithPunch(t, svcCtx, s.empZhang, workday,
		"2026-11-16 09:00:00", "", StatusNormal, "", OvertimeNo)
	if res.Status != SettleMissing {
		t.Fatalf("A6 status = %q, want missing", res.Status)
	}

	// A7 休息日: 周日打卡（规则不含周日）→ rest
	res = mustSettleWithPunch(t, svcCtx, s.empZhang, restday,
		"2026-11-15 10:00:00", "2026-11-15 12:00:00", StatusNormal, StatusNormal, OvertimeNo)
	if res.Status != SettleRest {
		t.Fatalf("A7 status = %q, want rest", res.Status)
	}

	// A8 休息日未打卡 → rest（休息日不判缺卡）
	res = mustSettle(t, svcCtx, s.empZhang, restday)
	if res.Status != SettleRest {
		t.Fatalf("A8 status = %q, want rest", res.Status)
	}

	// A9 无打卡记录的工作日 → missing（缺卡）
	res = mustSettle(t, svcCtx, s.empZhang, workday)
	if res.Status != SettleMissing {
		t.Fatalf("A9 status = %q, want missing", res.Status)
	}

	// A10 外勤打卡（外出）→ 按正常出勤
	res = mustSettleWithPunch(t, svcCtx, s.empZhang, workday,
		"2026-11-16 09:00:00", "2026-11-16 18:00:00", StatusField, StatusField, OvertimeNo)
	if res.Status != SettleNormal {
		t.Fatalf("A10 status = %q, want normal (外勤按正常出勤)", res.Status)
	}

	// 场景 B: 无规则兜底 —— 新员工挂在无任何配置的组织（公司A 无 config，组织规则只设在技术部）。
	orgNoRule := ""
	{
		req := &orgtypes.AddRequest{OrgName: "无规则部门", OrgType: "2", OrderNum: 9, Status: "1", ParentUuid: s.orgA}
		if _, err := orglogic.NewAdd(ctx, svcCtx, nil).Add(req); err != nil {
			t.Fatalf("add org 无规则部门: %v", err)
		}
		rows, err := svcCtx.Model.HrmOrg.FindByCondition(ctx, nil)
		if err != nil {
			t.Fatalf("find org: %v", err)
		}
		for _, r := range rows {
			if r.OrgName == "无规则部门" {
				orgNoRule = r.Uuid
			}
		}
	}
	req := &employeetypes.AddRequest{Name: "李四", EmployeeNo: "E0002", OrgUuid: orgNoRule, PersonStatus: "1"}
	if _, err := employeelogic.NewAdd(ctx, svcCtx, nil).Add(req); err != nil {
		t.Fatalf("add employee2: %v", err)
	}
	row, err := svcCtx.Model.HrmEmployee.FindOneByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_employee.EmployeeNo, "E0002").Build()...)
	if err != nil {
		t.Fatalf("find employee2: %v", err)
	}
	res = mustSettle(t, svcCtx, row.Uuid, workday)
	if res.Status != SettleNoRule {
		t.Fatalf("B1 status = %q, want no_rule (无规则兜底)", res.Status)
	}
	if res.Found {
		t.Fatalf("B1 Found = true, want false")
	}

	// 场景 C: 考勤组覆盖组织规则 + 跨天班（22:00-次日06:00）。
	// 新建独立组织避免 A 场景的组织规则干扰（考勤组优先级最高，直接挂张三）。
	addGroup(t, svcCtx, "夜班组", "22:00", "06:00", []int64{1, 2, 3, 4, 5, 6, 7}, []string{s.empZhang})

	// C1 跨天班正常: 周一 21:50 上班 周二 06:10 下班 → normal, 实打 8h20m = 8.33（周一 1 ∈ workDays）
	res = mustSettleWithPunch(t, svcCtx, s.empZhang, workday,
		"2026-11-16 21:50:00", "2026-11-17 06:10:00", StatusNormal, StatusNormal, OvertimeNo)
	if res.Status != SettleNormal {
		t.Fatalf("C1 status = %q, want normal", res.Status)
	}
	if res.WorkHours != 8.33 {
		t.Fatalf("C1 workHours = %v, want 8.33", res.WorkHours)
	}

	// C2 跨天班迟到: 22:15 上班 → late(15)
	res = mustSettleWithPunch(t, svcCtx, s.empZhang, workday,
		"2026-11-16 22:15:00", "2026-11-17 06:00:00", StatusNormal, StatusNormal, OvertimeNo)
	if res.Status != SettleLate || res.LateMinutes != 15 {
		t.Fatalf("C2 = %q late %d, want late 15", res.Status, res.LateMinutes)
	}

	// C3 跨天班早退: 下班 05:20 → early(40)
	res = mustSettleWithPunch(t, svcCtx, s.empZhang, workday,
		"2026-11-16 22:00:00", "2026-11-17 05:20:00", StatusNormal, StatusNormal, OvertimeNo)
	if res.Status != SettleEarly || res.EarlyMinutes != 40 {
		t.Fatalf("C3 = %q early %d, want early 40", res.Status, res.EarlyMinutes)
	}

	// C4 休息日判定跟随规则: 夜班组 7 天全勤 → 周日也 normal
	res = mustSettleWithPunch(t, svcCtx, s.empZhang, restday,
		"2026-11-15 22:00:00", "2026-11-16 06:00:00", StatusNormal, StatusNormal, OvertimeNo)
	if res.Status != SettleNormal {
		t.Fatalf("C4 status = %q, want normal (考勤组 7 天全勤覆盖)", res.Status)
	}

	// 场景 D: 弹性分钟 — 考勤组 flexible_minutes=10, 09:05 上班不判迟到。
	addGroupFlexible(t, svcCtx, "弹性组", "09:00", "18:00", []int64{1, 2, 3, 4, 5}, []string{row.Uuid}, 10)
	res = mustSettleWithPunch(t, svcCtx, row.Uuid, workday,
		"2026-11-16 09:05:00", "2026-11-16 18:00:00", StatusNormal, StatusNormal, OvertimeNo)
	if res.Status != SettleNormal {
		t.Fatalf("D1 status = %q, want normal (弹性10分钟)", res.Status)
	}
	res = mustSettleWithPunch(t, svcCtx, row.Uuid, workday,
		"2026-11-16 09:11:00", "2026-11-16 18:00:00", StatusNormal, StatusNormal, OvertimeNo)
	if res.Status != SettleLate || res.LateMinutes != 1 {
		t.Fatalf("D2 = %q late %d, want late 1 (超出弹性)", res.Status, res.LateMinutes)
	}
}

// 休息时段扣除: 考勤组 rest 12:00-13:00 由 ResolvedRule 不携带，日结只按 上/下班 时间差扣
// 规则中的休息时段（HRM-12 考勤组无 rest 字段，跳过）；此处验证 settle 落库。
func TestSettleDayPersists(t *testing.T) {
	svcCtx := newSvcCtx(t)
	s := seedOrgAndEmployee(t, svcCtx)
	saveOrgConfig(t, svcCtx, s.orgTech, "09:00", "18:00", []int64{1, 2, 3, 4, 5})

	res := mustSettleWithPunch(t, svcCtx, s.empZhang, "2026-11-16",
		"2026-11-16 08:55:00", "2026-11-16 18:05:00", StatusNormal, StatusNormal, OvertimeNo)
	if res.Status != SettleNormal {
		t.Fatalf("status = %q, want normal", res.Status)
	}

	// 落库断言: 通过 model 读回（不绕过接口查内部状态）
	row, err := svcCtx.Model.HrmPunchRecord.FindOneByCondition(context.Background(), nil,
		condition.NewChain().
			Equal(hrm_punch_record.PunchDate, "2026-11-16").
			Equal(hrm_punch_record.EmployeeUuid, s.empZhang).
			Build()...)
	if err != nil {
		t.Fatalf("read back settle row: %v", err)
	}
	if row.SettleStatus != SettleNormal {
		t.Fatalf("persisted settle_status = %q, want normal", row.SettleStatus)
	}
	if row.RuleSource != attendancegrouplogic.RuleSourceOrg {
		t.Fatalf("persisted rule_source = %q, want org", row.RuleSource)
	}
	if row.SettleLateMinutes != 0 || row.SettleEarlyMinutes != 0 {
		t.Fatalf("persisted late/early = %d/%d, want 0/0", row.SettleLateMinutes, row.SettleEarlyMinutes)
	}
	if !row.SettleTime.Valid {
		t.Fatalf("persisted settle_time should be set")
	}
}

// mustSettleWithPunch 造/覆盖当日打卡记录（一人一天一行，重复调用即覆盖）并日结。
func mustSettleWithPunch(t *testing.T, svcCtx *svc.ServiceContext, employeeUuid, date, clockIn, clockOut, inStatus, outStatus, overtime string) DaySettleResult {
	t.Helper()
	row, err := findPunchRow(context.Background(), svcCtx, employeeUuid, date)
	if err != nil {
		t.Fatalf("find punch row: %v", err)
	}
	if row == nil {
		row = &hrm_punch_record.HrmPunchRecord{
			Uuid:         newUuid(),
			PunchDate:    date,
			EmployeeUuid: employeeUuid,
			EmployeeNo:   "E0001",
			EmployeeName: "张三",
			Source:       SourceApp,
		}
		if _, err := svcCtx.Model.HrmPunchRecord.Insert(context.Background(), nil, row); err != nil {
			t.Fatalf("insert punch row: %v", err)
		}
	}
	row.ClockInStatus = inStatus
	row.ClockOutStatus = outStatus
	row.Overtime = overtime
	fields := map[string]any{
		"clock_in_status":  inStatus,
		"clock_out_status": outStatus,
		"overtime":         overtime,
		"clock_in_time":    nil,
		"clock_out_time":   nil,
	}
	if clockIn != "" {
		tm, err := time.ParseInLocation(time.DateTime, clockIn, time.Local)
		if err != nil {
			t.Fatalf("parse clockIn %q: %v", clockIn, err)
		}
		fields["clock_in_time"] = tm
	}
	if clockOut != "" {
		tm, err := time.ParseInLocation(time.DateTime, clockOut, time.Local)
		if err != nil {
			t.Fatalf("parse clockOut %q: %v", clockOut, err)
		}
		fields["clock_out_time"] = tm
	}
	if err := svcCtx.Model.HrmPunchRecord.UpdateFieldsByCondition(context.Background(), nil, fields,
		condition.NewChain().Equal(hrm_punch_record.Uuid, row.Uuid).Build()...); err != nil {
		t.Fatalf("update punch row: %v", err)
	}
	return mustSettle(t, svcCtx, employeeUuid, date)
}
