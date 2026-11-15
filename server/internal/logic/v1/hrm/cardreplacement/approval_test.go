package cardreplacement

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"

	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/punch"
	configModel "github.com/jzero-io/jzero-admin/server/internal/model/hrm_attendance_config"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_punch_record"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/cardreplacement"
)

// fixtures: workday 2026-11-16 是周一（与 punch.settle_test 同一 worked-example 日期）。
const workday = "2026-11-16"

// TestAddSnapshotsOriginalClocks 新增落库：HR206- 编号、未提交、员工快照、
// 原上班打卡/原下班打卡快照列（提交时 hrm_punch_record 当日行）——
// 独立期望值: req 字面量（09:10 / 18:05）.
func TestAddSnapshotsOriginalClocks(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")

	orgUuid := seedOrgTree(t, svcCtx)
	seedEmployee(t, svcCtx, orgUuid)

	// 造当日打卡行: 原上班 08:55 / 原下班 18:05
	row := &hrm_punch_record.HrmPunchRecord{
		Uuid:          uuidNew(),
		PunchDate:     workday,
		EmployeeUuid:  testEmpUuid,
		EmployeeNo:    "E0001",
		EmployeeName:  "张三",
		Source:        punch.SourceApp,
	}
	if _, err := svcCtx.Model.HrmPunchRecord.Insert(context.Background(), nil, row); err != nil {
		t.Fatalf("seed punch row: %v", err)
	}
	inAt := mustParseTime(t, "2026-11-16 08:55:00")
	outAt := mustParseTime(t, "2026-11-16 18:05:00")
	if err := svcCtx.Model.HrmPunchRecord.UpdateFieldsByCondition(context.Background(), nil, map[string]any{
		"clock_in_time": inAt, "clock_in_status": punch.StatusNormal,
		"clock_out_time": outAt, "clock_out_status": punch.StatusNormal,
	}, condByPunchUuid(row.Uuid)...); err != nil {
		t.Fatalf("seed punch times: %v", err)
	}

	billUuid, req := addBill(t, svcCtx, ctx, nil)

	rec := findBill(t, svcCtx, billUuid)
	if rec == nil {
		t.Fatal("bill missing")
	}
	if rec.EmployeeNo != "E0001" || rec.EmployeeName != "张三" {
		t.Fatalf("employee snapshot = %s/%s, want E0001/张三", rec.EmployeeNo, rec.EmployeeName)
	}
	if rec.OrigClockInTime != "2026-11-16 08:55:00" {
		t.Fatalf("origClockInTime = %q, want 2026-11-16 08:55:00 (原上班打卡快照)", rec.OrigClockInTime)
	}
	if rec.OrigClockOutTime != "2026-11-16 18:05:00" {
		t.Fatalf("origClockOutTime = %q, want 2026-11-16 18:05:00 (原下班打卡快照)", rec.OrigClockOutTime)
	}
	if rec.PunchDate != req.PunchDate || rec.CardType != req.CardType || rec.PunchTime != req.PunchTime {
		t.Fatalf("bill fields mismatch: %+v vs req %+v", rec, req)
	}
	if rec.CreatorName != "张三HR" || rec.CreatorUuid == "" {
		t.Fatalf("creator snapshot wrong: uuid=%q name=%q", rec.CreatorUuid, rec.CreatorName)
	}
}

// TestAddSnapshotEmptyWhenNoPunchRow 无打卡行 → 原打卡快照为空（缺卡语义）.
func TestAddSnapshotEmptyWhenNoPunchRow(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")

	orgUuid := seedOrgTree(t, svcCtx)
	seedEmployee(t, svcCtx, orgUuid)

	billUuid, _ := addBill(t, svcCtx, ctx, nil)
	rec := findBill(t, svcCtx, billUuid)
	if rec.OrigClockInTime != "" || rec.OrigClockOutTime != "" {
		t.Fatalf("orig clocks = %q/%q, want empty (no punch row)", rec.OrigClockInTime, rec.OrigClockOutTime)
	}
}

// TestAddValidatesInputs Add 校验: 类型非法 / 日期与时间不属同一天 → 10571/10573.
func TestAddValidatesInputs(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")

	orgUuid := seedOrgTree(t, svcCtx)
	seedEmployee(t, svcCtx, orgUuid)

	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		EmployeeUuid: testEmpUuid, PunchDate: workday,
		CardType: "lunch", PunchTime: workday + " 09:00:00",
	}); crErrCode(err) != 10571 {
		t.Fatalf("bad cardType err code = %v, want 10571", crErrCode(err))
	}
	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		EmployeeUuid: testEmpUuid, PunchDate: workday,
		CardType: CardTypeClockIn, PunchTime: "2026-11-17 09:00:00",
	}); crErrCode(err) != 10573 {
		t.Fatalf("cross-day punchTime err code = %v, want 10573", crErrCode(err))
	}
}

// TestApproveWritesPunchAndResettlesDay 审批通过事务内：补卡时间写入当日打卡行
// clock_in 列 + re-run SettleDay 迟到→正常（issues 验收核心场景）。
// 独立期望值: 组织默认规则 09:00 上班 → 09:10 迟到 10 分钟；补卡 08:50 后应 normal 0 分钟.
func TestApproveWritesPunchAndResettlesDay(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")

	orgUuid := seedOrgTree(t, svcCtx)
	seedEmployee(t, svcCtx, orgUuid)
	saveOrgConfig(t, svcCtx, orgUuid, "09:00", "18:00", []int64{1, 2, 3, 4, 5})

	// 造迟到场景: 上班卡 09:10（迟到 10 分钟）、下班卡 18:05，先日结固化快照
	seedPunchRow(t, svcCtx, testEmpUuid, workday, "2026-11-16 09:10:00", "2026-11-16 18:05:00")
	if res := mustSettle(t, svcCtx, testEmpUuid, workday); res.Status != punch.SettleLate {
		t.Fatalf("pre-approval settle = %q, want late", res.Status)
	}

	billUuid, _ := addBill(t, svcCtx, ctx, func(r *types.AddRequest) {
		r.CardType = CardTypeClockIn
		r.PunchTime = "2026-11-16 08:50:00"
	})
	submitAndApprove(t, env, svcCtx, billUuid)

	// 打卡行 clock_in 已被补卡时间覆盖（原 09:10 → 08:50）
	prow := findPunchRowOf(t, svcCtx, testEmpUuid, workday)
	if prow == nil || !prow.ClockInTime.Valid {
		t.Fatal("clock_in row/time missing after approval")
	}
	if got := prow.ClockInTime.Time.Format("2006-01-02 15:04:05"); got != "2026-11-16 08:50:00" {
		t.Fatalf("clock_in_time = %q, want 2026-11-16 08:50:00 (补卡写入)", got)
	}
	if prow.Source != punch.SourceCardReplacement {
		t.Fatalf("source = %q, want card_replacement", prow.Source)
	}

	// 日结快照已被重算: 迟到 → 正常, late_minutes 0
	if prow.SettleStatus != punch.SettleNormal {
		t.Fatalf("settle_status = %q, want normal (迟到→正常)", prow.SettleStatus)
	}
	if prow.SettleLateMinutes != 0 {
		t.Fatalf("settle_late_minutes = %d, want 0", prow.SettleLateMinutes)
	}

	// 单据状态 审批通过
	if rec := findBill(t, svcCtx, billUuid); rec.Status != bill.StatusApproved {
		t.Fatalf("status after approval = %q, want approved", rec.Status)
	}
}

// TestApproveClockOutCreatesPunchRow 无打卡行 + 下班卡补卡 → 建行写 clock_out + 日结重算.
func TestApproveClockOutCreatesPunchRow(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")

	orgUuid := seedOrgTree(t, svcCtx)
	seedEmployee(t, svcCtx, orgUuid)
	saveOrgConfig(t, svcCtx, orgUuid, "09:00", "18:00", []int64{1, 2, 3, 4, 5})

	// 当日已有迟到上班卡、无下班卡 → 缺卡
	seedPunchRow(t, svcCtx, testEmpUuid, workday, "2026-11-16 09:05:00", "")
	if res := mustSettle(t, svcCtx, testEmpUuid, workday); res.Status != punch.SettleMissing {
		t.Fatalf("pre-approval settle = %q, want missing", res.Status)
	}

	billUuid, _ := addBill(t, svcCtx, ctx, func(r *types.AddRequest) {
		r.CardType = CardTypeClockOut
		r.PunchTime = "2026-11-16 18:10:00"
	})
	submitAndApprove(t, env, svcCtx, billUuid)

	prow := findPunchRowOf(t, svcCtx, testEmpUuid, workday)
	if prow == nil || !prow.ClockOutTime.Valid {
		t.Fatal("clock_out missing after approval (应写入已有行)")
	}
	if got := prow.ClockOutTime.Time.Format("2006-01-02 15:04:05"); got != "2026-11-16 18:10:00" {
		t.Fatalf("clock_out_time = %q, want 2026-11-16 18:10:00", got)
	}
	// 09:05 上班（弹性内? 无弹性 → 迟到 5 分钟）、18:10 下班 → late（迟到优先展示），非 missing
	if prow.SettleStatus != punch.SettleLate {
		t.Fatalf("settle_status = %q, want late (缺卡→重算后迟到)", prow.SettleStatus)
	}
}

// TestApproveCreatesMissingRow 当日完全无打卡行 → 补卡建行（含快照列）+ 日结.
func TestApproveCreatesMissingRow(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")

	orgUuid := seedOrgTree(t, svcCtx)
	seedEmployee(t, svcCtx, orgUuid)
	saveOrgConfig(t, svcCtx, orgUuid, "09:00", "18:00", []int64{1, 2, 3, 4, 5})

	billUuid, _ := addBill(t, svcCtx, ctx, func(r *types.AddRequest) {
		r.CardType = CardTypeClockIn
		r.PunchTime = "2026-11-16 08:55:00"
	})
	submitAndApprove(t, env, svcCtx, billUuid)

	prow := findPunchRowOf(t, svcCtx, testEmpUuid, workday)
	if prow == nil {
		t.Fatal("punch row not created for card replacement")
	}
	if prow.EmployeeNo != "E0001" || prow.EmployeeName != "张三" {
		t.Fatalf("punch row snapshot = %s/%s, want E0001/张三", prow.EmployeeNo, prow.EmployeeName)
	}
	if prow.Source != punch.SourceCardReplacement {
		t.Fatalf("source = %q, want card_replacement", prow.Source)
	}
	if prow.SettleStatus != punch.SettleMissing {
		// 仅上班卡、缺下班卡 → missing（日结语义）
		t.Fatalf("settle_status = %q, want missing (只有上班卡)", prow.SettleStatus)
	}
}

// TestApproveIdempotent 审批通过是终态：重复 ApproveBill 被状态机拒绝，打卡行不二次变更.
func TestApproveIdempotent(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")

	orgUuid := seedOrgTree(t, svcCtx)
	seedEmployee(t, svcCtx, orgUuid)
	saveOrgConfig(t, svcCtx, orgUuid, "09:00", "18:00", []int64{1, 2, 3, 4, 5})

	billUuid, _ := addBill(t, svcCtx, ctx, nil)
	submitAndApprove(t, env, svcCtx, billUuid)

	// 重复审批 → ErrInvalidTransition
	err := bill.ApproveBill(approvalCtx(svcCtx), env.SqlConn, BillTypeCode, billUuid,
		bill.Operator{UserUuid: "op-1", Name: "管理员"}, "再次同意")
	if err == nil {
		t.Fatal("second approve must fail (审批通过 is terminal)")
	}

	prow := findPunchRowOf(t, svcCtx, testEmpUuid, workday)
	if got := prow.ClockInTime.Time.Format("2006-01-02 15:04:05"); got != "2026-11-16 09:10:00" {
		t.Fatalf("clock_in_time = %q, want unchanged 2026-11-16 09:10:00", got)
	}
}

// TestLifecycleGuardsAndFilters 生命周期守卫 + mine-by-default + 筛选.
func TestLifecycleGuardsAndFilters(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")

	orgUuid := seedOrgTree(t, svcCtx)
	seedEmployee(t, svcCtx, orgUuid)

	billUuid, _ := addBill(t, svcCtx, ctx, nil)

	// 审批中不可编辑/删除
	if err := bill.SubmitBill(approvalCtx(svcCtx), env.SqlConn, BillTypeCode, billUuid,
		bill.Operator{UserUuid: "op-1", Name: "管理员"}, ""); err != nil {
		t.Fatalf("SubmitBill: %v", err)
	}
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid: billUuid, EmployeeUuid: testEmpUuid, PunchDate: workday,
		CardType: CardTypeClockIn, PunchTime: workday + " 09:00:00",
	}); crErrCode(err) != 10572 {
		t.Fatalf("edit while approving err code = %v, want 10572", crErrCode(err))
	}
	if _, err := NewDelete(ctx, svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{billUuid}}); crErrCode(err) != 10572 {
		t.Fatalf("delete while approving err code = %v, want 10572", crErrCode(err))
	}

	// 撤销 → 已撤销可编辑/删除
	if err := bill.RevokeBill(approvalCtx(svcCtx), env.SqlConn, BillTypeCode, billUuid,
		bill.Operator{UserUuid: "op-1", Name: "管理员"}, ""); err != nil {
		t.Fatalf("RevokeBill: %v", err)
	}
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid: billUuid, EmployeeUuid: testEmpUuid, PunchDate: workday,
		CardType: CardTypeClockOut, PunchTime: workday + " 18:00:00", Reason: "忘记打卡下班",
	}); err != nil {
		t.Fatalf("edit after revoke: %v", err)
	}

	// 筛选: 补卡类型
	filtered := listBills(t, svcCtx, ctx, &types.ListRequest{CardType: CardTypeClockOut, Mine: 1})
	if len(filtered.Records) != 1 {
		t.Fatalf("clockOut filter records = %d, want 1", len(filtered.Records))
	}
	other := listBills(t, svcCtx, ctx, &types.ListRequest{CardType: CardTypeClockIn, Mine: 1})
	if len(other.Records) != 0 {
		t.Fatalf("clockIn filter records = %d, want 0", len(other.Records))
	}

	// 筛选: 工号
	byNo := listBills(t, svcCtx, ctx, &types.ListRequest{EmployeeNo: "E0001", Mine: 1})
	if len(byNo.Records) != 1 {
		t.Fatalf("employeeNo filter records = %d, want 1", len(byNo.Records))
	}

	// 我的单据: 其他用户看不到 user-1 的单据
	mine2 := listBills(t, svcCtx, testCtxFor("user-2", "李四"), &types.ListRequest{})
	if len(mine2.Records) != 0 {
		t.Fatalf("other user sees %d bills, want 0 (mine-by-default)", len(mine2.Records))
	}
	// 管理员可见
	admin := listBills(t, svcCtx, superCtx("admin", "管理员"), &types.ListRequest{})
	if len(admin.Records) != 1 {
		t.Fatalf("admin sees %d bills, want 1", len(admin.Records))
	}
}

// ---- fixture helpers（组织默认规则/打卡行造数；口径同 punch.settle_test）----

func saveOrgConfig(t *testing.T, svcCtx *svc.ServiceContext, orgUuid, startTime, endTime string, workDays []int64) {
	t.Helper()
	// 直接写 hrm_attendance_config（与 attendanceconfig.Save 端点同语义），
	// 避免跨包引入 attendanceconfig logic 的类型歧义。
	days := ""
	for i, d := range workDays {
		if i > 0 {
			days += ","
		}
		days += itoa64(d)
	}
	if _, err := svcCtx.Model.HrmAttendanceConfig.Insert(context.Background(), nil, &configModel.HrmAttendanceConfig{
		Uuid: uuidNew(), OrgUuid: orgUuid,
		StartTime: startTime, EndTime: endTime, WorkDays: days,
	}); err != nil {
		t.Fatalf("save config: %v", err)
	}
}

func itoa64(v int64) string {
	return strconv.FormatInt(v, 10)
}

func seedPunchRow(t *testing.T, svcCtx *svc.ServiceContext, employeeUuid, date, clockIn, clockOut string) {
	t.Helper()
	row, err := svcCtx.Model.HrmPunchRecord.FindOneByCondition(context.Background(), nil,
		condByPunch(employeeUuid, date)...)
	if err != nil && err != hrm_punch_record.ErrNotFound {
		t.Fatalf("find punch row: %v", err)
	}
	if row == nil {
		row = &hrm_punch_record.HrmPunchRecord{
			Uuid:         uuidNew(),
			PunchDate:    date,
			EmployeeUuid: employeeUuid,
			EmployeeNo:   "E0001",
			EmployeeName: "张三",
			Source:       punch.SourceApp,
		}
		if _, err := svcCtx.Model.HrmPunchRecord.Insert(context.Background(), nil, row); err != nil {
			t.Fatalf("insert punch row: %v", err)
		}
	}
	fields := map[string]any{}
	if clockIn != "" {
		fields["clock_in_time"] = mustParseTime(t, clockIn)
		fields["clock_in_status"] = punch.StatusNormal
	}
	if clockOut != "" {
		fields["clock_out_time"] = mustParseTime(t, clockOut)
		fields["clock_out_status"] = punch.StatusNormal
	}
	if err := svcCtx.Model.HrmPunchRecord.UpdateFieldsByCondition(context.Background(), nil, fields,
		condByPunchUuid(row.Uuid)...); err != nil {
		t.Fatalf("update punch row: %v", err)
	}
}

func mustSettle(t *testing.T, svcCtx *svc.ServiceContext, employeeUuid, date string) punch.DaySettleResult {
	t.Helper()
	res, err := punch.SettleDay(context.Background(), svcCtx, employeeUuid, date)
	if err != nil {
		t.Fatalf("SettleDay(%s, %s): %v", employeeUuid, date, err)
	}
	return *res
}

func mustParseTime(t *testing.T, s string) time.Time {
	t.Helper()
	v, err := time.ParseInLocation("2006-01-02 15:04:05", s, time.Local)
	if err != nil {
		t.Fatalf("parse time %q: %v", s, err)
	}
	return v
}

func condByPunch(employeeUuid, date string) []condition.Condition {
	return condition.NewChain().
		Equal(hrm_punch_record.PunchDate, date).
		Equal(hrm_punch_record.EmployeeUuid, employeeUuid).
		Build()
}

func condByPunchUuid(uuid string) []condition.Condition {
	return condition.NewChain().Equal(hrm_punch_record.Uuid, uuid).Build()
}
