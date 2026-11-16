package perf_period

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	coreSvc "github.com/jzero-io/jzero-admin/core-engine/svc"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_org"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_scheme"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/perf_period"
)

// DB-backed logic tests for 绩效批次 (HRM-42), run via svctest:
// sqlite by default; HRM_TEST_DB=mysql runs the same assertions against a
// disposable docker MySQL with the real migrations.
//
// Coverage (issue #44 acceptance criteria):
//   - PERF 编号（PERF-年-流水，复用 hrm_bill_sequence，bill_type_code='PERF'，日期=年份）
//   - 圈人逻辑（适用范围多维条件组合，部门含下级）
//   - 生命周期状态机（草稿→进行中→待确认→待校准→结果已发布→已发布，forward-only）
//   - 发布守卫（发布前须有参与人）+ 幂等（圈人全量覆盖，重复圈定结果一致）

var testCtx = context.Background()

func newCtx(t *testing.T) (*svc.ServiceContext, *svctest.Env) {
	t.Helper()
	env := svctest.New(t)
	return &svc.ServiceContext{
		ServiceContext: &coreSvc.ServiceContext{SqlxConn: env.SqlConn},
		Model:          env.Model,
	}, env
}

// mustAddScheme seeds one 启用 绩效方案 and returns its uuid.
func mustAddScheme(t *testing.T, env *svctest.Env) string {
	t.Helper()
	uuid := fmt.Sprintf("scheme-%d", envDial(env))
	err := hrm_perf_scheme.NewHrmPerfSchemeModel(env.SqlConn).InsertV2(testCtx, nil, &hrm_perf_scheme.HrmPerfScheme{
		Uuid:           uuid,
		Name:           "绩效方案",
		TemplateUuid:   "tpl-x",
		PeriodType:     "month",
		AssessmentMode: "kpi",
		ScoreMode:      "percent",
		PayLinkMode:    "coefficient",
		Status:         "1",
	})
	if err != nil {
		t.Fatalf("seed scheme: %v", err)
	}
	return uuid
}

var envCounter int

func envDial(env *svctest.Env) int {
	envCounter++
	return envCounter
}

// seedOrg inserts one 组织 node and returns its uuid.
func seedOrg(t *testing.T, env *svctest.Env, parentUuid, name, orgType string) string {
	t.Helper()
	uuid := fmt.Sprintf("org-%s", name)
	err := hrm_org.NewHrmOrgModel(env.SqlConn).InsertV2(testCtx, nil, &hrm_org.HrmOrg{
		Uuid:       uuid,
		ParentUuid: parentUuid,
		OrgName:    name,
		OrgType:    orgType,
		OrderNum:   1,
		Status:     "1",
	})
	if err != nil {
		t.Fatalf("seed org %s: %v", name, err)
	}
	return uuid
}

// seedEmployee inserts one 员工档案 row (irrelevant archive fields left empty)
// and returns its uuid. personStatus: 1=正式 2=试用期 5=离职 6=退休.
func seedEmployee(t *testing.T, env *svctest.Env, name, companyUuid, orgUuid, personStatus, position, post string) string {
	t.Helper()
	uuid := fmt.Sprintf("emp-%s", name)
	err := hrm_employee.NewHrmEmployeeModel(env.SqlConn).InsertV2(testCtx, nil, &hrm_employee.HrmEmployee{
		Uuid:         uuid,
		EmployeeNo:   fmt.Sprintf("E%03d", len(name)),
		Name:         name,
		CompanyUuid:  companyUuid,
		OrgUuid:      orgUuid,
		Position:     position,
		Post:         post,
		PersonStatus: personStatus,
	})
	if err != nil {
		t.Fatalf("seed employee %s: %v", name, err)
	}
	return uuid
}

func participantUuids(t *testing.T, env *svctest.Env, periodUuid string) []string {
	t.Helper()
	rows, err := GetPeriodParticipants(testCtx, &svc.ServiceContext{
		ServiceContext: &coreSvc.ServiceContext{SqlxConn: env.SqlConn},
		Model:          env.Model,
	}, periodUuid)
	if err != nil {
		t.Fatalf("GetPeriodParticipants: %v", err)
	}
	uuids := make([]string, 0, len(rows))
	for _, row := range rows {
		uuids = append(uuids, row.EmployeeUuid)
	}
	return uuids
}

func asSet(list []string) map[string]struct{} {
	set := make(map[string]struct{}, len(list))
	for _, v := range list {
		set[v] = struct{}{}
	}
	return set
}

// ---- PERF 编号 ----

func TestFormatPeriodNo(t *testing.T) {
	no, err := FormatPeriodNo(2026, 1)
	if err != nil || no != "PERF-2026-0001" {
		t.Fatalf("want PERF-2026-0001, got %q err=%v", no, err)
	}
	no, err = FormatPeriodNo(2026, 9999)
	if err != nil || no != "PERF-2026-9999" {
		t.Fatalf("want PERF-2026-9999, got %q err=%v", no, err)
	}
	if _, err := FormatPeriodNo(2026, 10000); !errors.Is(err, ErrPeriodNoSeqOutOfRange) {
		t.Fatalf("want seq out of range, got %v", err)
	}
}

func TestGeneratePeriodNoPerYear(t *testing.T) {
	_, env := newCtx(t)
	schemeUuid := mustAddScheme(t, env)

	// 同一年度内流水递增（日期=年份）
	u1, err := AddPeriod(testCtx, env.SqlConn, "2026月度批次", "month", schemeUuid,
		`{"companyUuids":[],"orgUuids":[],"positions":[],"posts":[],"personStatuses":["1"]}`,
		"2026-01-01", "2026-01-31", "", "")
	if err != nil {
		t.Fatalf("add period 1: %v", err)
	}
	u2, err := AddPeriod(testCtx, env.SqlConn, "2026第二月批次", "month", schemeUuid,
		`{"companyUuids":[],"orgUuids":[],"positions":[],"posts":[],"personStatuses":["1"]}`,
		"2026-02-01", "2026-02-28", "", "")
	if err != nil {
		t.Fatalf("add period 2: %v", err)
	}

	p1, err := FindPeriodByUuid(testCtx, env.SqlConn, u1)
	if err != nil {
		t.Fatalf("find period 1: %v", err)
	}
	p2, err := FindPeriodByUuid(testCtx, env.SqlConn, u2)
	if err != nil {
		t.Fatalf("find period 2: %v", err)
	}
	if !strings.HasPrefix(p1.PeriodNo, "PERF-") {
		t.Fatalf("period no must start with PERF-: %q", p1.PeriodNo)
	}
	if !strings.HasPrefix(p2.PeriodNo, "PERF-") {
		t.Fatalf("period no must start with PERF-: %q", p2.PeriodNo)
	}
	seq1 := strings.Split(p1.PeriodNo, "-")
	seq2 := strings.Split(p2.PeriodNo, "-")
	if len(seq1) != 3 || len(seq2) != 3 {
		t.Fatalf("period no format must be PERF-年-流水: %q / %q", p1.PeriodNo, p2.PeriodNo)
	}
	if seq1[1] != "2026" || seq2[1] != "2026" {
		t.Fatalf("year segment must be 周期开始年份: %q / %q", p1.PeriodNo, p2.PeriodNo)
	}
	if seq2[2] <= seq1[2] {
		t.Fatalf("seq must increase within the year: %q then %q", p1.PeriodNo, p2.PeriodNo)
	}
}

// ---- CRUD + 列表 enrich ----

func TestPeriodCrudAndList(t *testing.T) {
	_, env := newCtx(t)
	schemeUuid := mustAddScheme(t, env)

	uuid, err := AddPeriod(testCtx, env.SqlConn, "2026年1月月度批次", "month", schemeUuid,
		`{"companyUuids":[],"orgUuids":[],"positions":["后端工程师"],"posts":[],"personStatuses":["1","2"]}`,
		"2026-01-01", "2026-01-31", "2026-02-05", "备注")
	if err != nil {
		t.Fatalf("add: %v", err)
	}

	row, err := FindPeriodByUuid(testCtx, env.SqlConn, uuid)
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if row.Status != PeriodStatusDraft {
		t.Fatalf("new period must be 草稿(0), got %q", row.Status)
	}
	if row.Name != "2026年1月月度批次" || row.PeriodType != "month" || row.SchemeUuid != schemeUuid {
		t.Fatalf("unexpected row: %+v", row)
	}

	// Edit（草稿可编辑）
	if err := EditPeriod(testCtx, env.SqlConn, uuid, "2026年1月月度批次(改)", "month", schemeUuid,
		`{"companyUuids":[],"orgUuids":[],"positions":[],"posts":[],"personStatuses":["1"]}`,
		"2026-01-01", "2026-01-31", "", ""); err != nil {
		t.Fatalf("edit: %v", err)
	}
	row, _ = FindPeriodByUuid(testCtx, env.SqlConn, uuid)
	if row.Name != "2026年1月月度批次(改)" {
		t.Fatalf("edit not applied: %+v", row)
	}

	// Edit not found（合法枚举 + 不存在的 uuid）
	err = EditPeriod(testCtx, env.SqlConn, "missing", "x", "month", schemeUuid, "{}", "2026-01-01", "2026-01-31", "", "")
	if !errors.Is(err, ErrPeriodNotFound) {
		t.Fatalf("want ErrPeriodNotFound, got %v", err)
	}

	// List 带 schemeName + participantCount 快照
	l := NewList(testCtx, &svc.ServiceContext{
		ServiceContext: &coreSvc.ServiceContext{SqlxConn: env.SqlConn},
		Model:          env.Model,
	}, nil)
	var req types.ListRequest
	req.PageRequest.Current = 1
	req.PageRequest.Size = 10
	resp, err := l.List(&req)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if resp.Total != 1 || len(resp.Records) != 1 {
		t.Fatalf("want 1 record, got total=%d", resp.Total)
	}
	rec := resp.Records[0]
	if rec.SchemeName != "绩效方案" {
		t.Fatalf("want schemeName snapshot, got %q", rec.SchemeName)
	}

	// Delete（草稿可删）
	if err := DeletePeriods(testCtx, env.SqlConn, []string{uuid}); err != nil {
		t.Fatalf("delete: %v", err)
	}
	_, err = FindPeriodByUuid(testCtx, env.SqlConn, uuid)
	if !errors.Is(err, ErrPeriodNotFound) {
		t.Fatalf("want ErrPeriodNotFound after delete, got %v", err)
	}
}

// ---- 圈人逻辑（多维条件组合，部门含下级）----

// Org tree:
//
//	C(公司) ─ D1(部门) ─ D2(下级部门)
//	        └ D3(部门)
func seedScopeFixture(t *testing.T, env *svctest.Env) (company, d1, d2, d3 string) {
	t.Helper()
	company = seedOrg(t, env, "", "总公司", "1")
	d1 = seedOrg(t, env, company, "研发一部", "2")
	d2 = seedOrg(t, env, d1, "研发一组", "2")
	d3 = seedOrg(t, env, company, "市场部", "2")

	seedEmployee(t, env, "张三", company, d1, "1", "后端工程师", "专员")   // 命中 D1
	seedEmployee(t, env, "李四", company, d2, "1", "后端工程师", "专员")   // 命中 D1 的下级 D2
	seedEmployee(t, env, "王五", company, d3, "1", "市场专员", "专员")     // 不在 D1 子树
	seedEmployee(t, env, "赵六", company, d1, "5", "后端工程师", "专员")   // 离职，非在册
	seedEmployee(t, env, "钱七", company, d1, "1", "前端工程师", "专员")   // 职位不匹配
	seedEmployee(t, env, "孙八", company, d1, "2", "前端工程师", "经理")   // 试用期+职务经理
	return
}

func TestScopeParticipantsDepartmentWithDescendants(t *testing.T) {
	_, env := newCtx(t)
	company, d1, _, d3 := seedScopeFixture(t, env)
	schemeUuid := mustAddScheme(t, env)

	// 部门 D1（含下级）+ 人员状态正式
	scope := fmt.Sprintf(`{"companyUuids":[],"orgUuids":["%s"],"positions":[],"posts":[],"personStatuses":["1"]}`, d1)
	uuid, err := AddPeriod(testCtx, env.SqlConn, "研发一部批次", "month", schemeUuid, scope, "2026-01-01", "2026-01-31", "", "")
	if err != nil {
		t.Fatalf("add: %v", err)
	}

	got := asSet(participantUuids(t, env, uuid))
	// D1 子树在册 + 状态正式 = 张三/李四/钱七（钱七职位不同但 scope 未按职位过滤）
	want := asSet([]string{"emp-张三", "emp-李四", "emp-钱七"})
	if len(got) != 3 {
		t.Fatalf("want {张三,李四,钱七} (部门含下级 D2), got %v", got)
	}
	for k := range want {
		if _, ok := got[k]; !ok {
			t.Fatalf("missing participant %s, got %v", k, got)
		}
	}
	if _, ok := got["emp-王五"]; ok {
		t.Fatalf("王五 (不在 D1 子树) must not participate: %v", got)
	}
	if _, ok := got["emp-赵六"]; ok {
		t.Fatalf("赵六 (离职) must not participate: %v", got)
	}
	_ = company
	_ = d3
}

func TestScopeParticipantsCompanyWide(t *testing.T) {
	_, env := newCtx(t)
	company, _, _, _ := seedScopeFixture(t, env)
	schemeUuid := mustAddScheme(t, env)

	// 公司级圈定 → 全公司在册员工
	scope := fmt.Sprintf(`{"companyUuids":["%s"],"orgUuids":[],"positions":[],"posts":[],"personStatuses":[]}`, company)
	uuid, err := AddPeriod(testCtx, env.SqlConn, "全公司批次", "month", schemeUuid, scope, "2026-01-01", "2026-01-31", "", "")
	if err != nil {
		t.Fatalf("add: %v", err)
	}

	got := asSet(participantUuids(t, env, uuid))
	// 在册口径: 赵六(离职5) 不参与；其余 5 人都在册
	if len(got) != 5 {
		t.Fatalf("want 5 在册 participants, got %d: %v", len(got), got)
	}
	if _, ok := got["emp-赵六"]; ok {
		t.Fatalf("离职员工 must not participate: %v", got)
	}
}

func TestScopeParticipantsMultiDimCombo(t *testing.T) {
	_, env := newCtx(t)
	company, d1, _, _ := seedScopeFixture(t, env)
	schemeUuid := mustAddScheme(t, env)

	// 多维组合: 部门含下级 + 职位=后端工程师 → 张三+李四
	scope := fmt.Sprintf(`{"companyUuids":[],"orgUuids":["%s"],"positions":["后端工程师"],"posts":[],"personStatuses":[]}`, d1)
	uuid, err := AddPeriod(testCtx, env.SqlConn, "后端批次", "month", schemeUuid, scope, "2026-01-01", "2026-01-31", "", "")
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	got := asSet(participantUuids(t, env, uuid))
	if len(got) != 2 {
		t.Fatalf("want {张三,李四} (多维: D1子树+后端工程师), got %v", got)
	}

	// 多维组合: 全公司 + 职务=经理 → 孙八
	scope2 := fmt.Sprintf(`{"companyUuids":["%s"],"orgUuids":[],"positions":[],"posts":["经理"],"personStatuses":[]}`, company)
	uuid2, err := AddPeriod(testCtx, env.SqlConn, "经理批次", "month", schemeUuid, scope2, "2026-01-01", "2026-01-31", "", "")
	if err != nil {
		t.Fatalf("add 2: %v", err)
	}
	got2 := asSet(participantUuids(t, env, uuid2))
	if len(got2) != 1 {
		t.Fatalf("want {孙八} (职务=经理), got %v", got2)
	}
	if _, ok := got2["emp-孙八"]; !ok {
		t.Fatalf("want 孙八, got %v", got2)
	}
}

func TestRefreshParticipantsIdempotent(t *testing.T) {
	_, env := newCtx(t)
	company, _, _, _ := seedScopeFixture(t, env)
	schemeUuid := mustAddScheme(t, env)

	scope := fmt.Sprintf(`{"companyUuids":["%s"],"orgUuids":[],"positions":[],"posts":[],"personStatuses":[]}`, company)
	uuid, err := AddPeriod(testCtx, env.SqlConn, "幂等批次", "month", schemeUuid, scope, "2026-01-01", "2026-01-31", "", "")
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	first := participantUuids(t, env, uuid)

	// 重复圈定：全量覆盖，结果一致（幂等）
	if _, err := RefreshParticipants(testCtx, env.SqlConn, uuid); err != nil {
		t.Fatalf("refresh 1: %v", err)
	}
	second := participantUuids(t, env, uuid)
	if len(first) != len(second) {
		t.Fatalf("refresh must be idempotent: %d vs %d", len(first), len(second))
	}
	for _, u := range first {
		if _, ok := asSet(second)[u]; !ok {
			t.Fatalf("participant %s lost after refresh", u)
		}
	}
}

// ---- 生命周期状态机 + 发布守卫 ----

func TestLifecycleStateMachine(t *testing.T) {
	_, env := newCtx(t)
	company, _, _, _ := seedScopeFixture(t, env)
	schemeUuid := mustAddScheme(t, env)

	scope := fmt.Sprintf(`{"companyUuids":["%s"],"orgUuids":[],"positions":[],"posts":[],"personStatuses":[]}`, company)
	uuid, err := AddPeriod(testCtx, env.SqlConn, "状态机批次", "month", schemeUuid, scope, "2026-01-01", "2026-01-31", "", "")
	if err != nil {
		t.Fatalf("add: %v", err)
	}

	assertStatus := func(want string) {
		t.Helper()
		row, err := FindPeriodByUuid(testCtx, env.SqlConn, uuid)
		if err != nil {
			t.Fatalf("find: %v", err)
		}
		if row.Status != want {
			t.Fatalf("status want %s got %s", want, row.Status)
		}
	}

	assertStatus(PeriodStatusDraft)

	// 发布：草稿 → 进行中（有参与人，守卫通过）
	if err := PublishPeriod(testCtx, env.SqlConn, uuid); err != nil {
		t.Fatalf("publish: %v", err)
	}
	assertStatus(PeriodStatusOngoing)

	// 幂等守卫：再次发布 = 非法流转
	if err := PublishPeriod(testCtx, env.SqlConn, uuid); !errors.Is(err, ErrPeriodTransitionInvalid) {
		t.Fatalf("re-publish must be invalid transition, got %v", err)
	}

	// 草稿期专属动作在发布后被拒
	if _, err := RefreshParticipants(testCtx, env.SqlConn, uuid); !errors.Is(err, ErrPeriodNotEditable) {
		t.Fatalf("refresh after publish must be rejected, got %v", err)
	}

	// 逐级推进 1→2→3→4→5
	for _, want := range []string{PeriodStatusConfirming, PeriodStatusCalibrating,
		PeriodStatusResultPublished, PeriodStatusPublished} {
		if err := AdvancePeriod(testCtx, env.SqlConn, uuid); err != nil {
			t.Fatalf("advance to %s: %v", want, err)
		}
		assertStatus(want)
	}

	// 终态不可再推进
	if err := AdvancePeriod(testCtx, env.SqlConn, uuid); !errors.Is(err, ErrPeriodTransitionInvalid) {
		t.Fatalf("advance past 终态 must be invalid, got %v", err)
	}
}

func TestPublishGuardRequiresParticipants(t *testing.T) {
	_, env := newCtx(t)
	schemeUuid := mustAddScheme(t, env)

	// 空范围也可以是合法批次（全员），这里用不可能命中的多维条件制造 0 参与人
	scope := `{"companyUuids":[],"orgUuids":[],"positions":["不存在职位XYZ"],"posts":[],"personStatuses":[]}`
	uuid, err := AddPeriod(testCtx, env.SqlConn, "零参与人批次", "month", schemeUuid, scope, "2026-01-01", "2026-01-31", "", "")
	if err != nil {
		t.Fatalf("add: %v", err)
	}

	if err := PublishPeriod(testCtx, env.SqlConn, uuid); !errors.Is(err, ErrPeriodNoParticipants) {
		t.Fatalf("publish with 0 participants must be guarded, got %v", err)
	}

	// 发布守卫之后状态仍为草稿
	row, err := FindPeriodByUuid(testCtx, env.SqlConn, uuid)
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if row.Status != PeriodStatusDraft {
		t.Fatalf("failed publish must not mutate status, got %s", row.Status)
	}
}

func TestEditDeleteOnlyInDraft(t *testing.T) {
	_, env := newCtx(t)
	company, _, _, _ := seedScopeFixture(t, env)
	schemeUuid := mustAddScheme(t, env)

	scope := fmt.Sprintf(`{"companyUuids":["%s"],"orgUuids":[],"positions":[],"posts":[],"personStatuses":[]}`, company)
	uuid, err := AddPeriod(testCtx, env.SqlConn, "草稿期批次", "month", schemeUuid, scope, "2026-01-01", "2026-01-31", "", "")
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := PublishPeriod(testCtx, env.SqlConn, uuid); err != nil {
		t.Fatalf("publish: %v", err)
	}

	// 进行中后编辑/删除被拒
	err = EditPeriod(testCtx, env.SqlConn, uuid, "改名", "month", schemeUuid, scope, "2026-01-01", "2026-01-31", "", "")
	if !errors.Is(err, ErrPeriodNotEditable) {
		t.Fatalf("edit after publish must be rejected, got %v", err)
	}
	err = DeletePeriods(testCtx, env.SqlConn, []string{uuid})
	if !errors.Is(err, ErrPeriodNotEditable) {
		t.Fatalf("delete after publish must be rejected, got %v", err)
	}
}

// ---- 校验 ----

func TestPeriodValidation(t *testing.T) {
	_, env := newCtx(t)
	schemeUuid := mustAddScheme(t, env)

	cases := []struct {
		label string
		name  string
		pType string
		scope string
		start string
		end   string
	}{
		{"名称为空", "", "month", "{}", "2026-01-01", "2026-01-31"},
		{"批次类型非法", "x", "year", "{}", "2026-01-01", "2026-01-31"},
		{"方案缺失", "x", "month", "", "2026-01-01", "2026-01-31"},
		{"周期开始格式非法", "x", "month", "{}", "20260101", "2026-01-31"},
		{"结束早于开始", "x", "month", "{}", "2026-02-01", "2026-01-31"},
		{"scope 非法 JSON", "x", "month", "not-json", "2026-01-01", "2026-01-31"},
		{"人员状态非法", "x", "month", `{"personStatuses":["9"]}`, "2026-01-01", "2026-01-31"},
	}
	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			_, err := AddPeriod(testCtx, env.SqlConn, tc.name, tc.pType, schemeUuid, tc.scope, tc.start, tc.end, "", "")
			if err == nil {
				t.Fatalf("%s: want validation error", tc.label)
			}
			if !errors.Is(err, ErrPeriodInvalid) && !errors.Is(err, ErrPeriodNameRequired) &&
				!errors.Is(err, ErrPeriodSchemeRequired) && !errors.Is(err, ErrPeriodScopeInvalid) {
				t.Fatalf("%s: unexpected error type: %v", tc.label, err)
			}
		})
	}

	// 引用不存在的绩效方案
	_, err := AddPeriod(testCtx, env.SqlConn, "方案缺失批次", "month", "no-such-scheme", "{}", "2026-01-01", "2026-01-31", "", "")
	if !errors.Is(err, ErrPeriodSchemeNotFound) {
		t.Fatalf("want ErrPeriodSchemeNotFound, got %v", err)
	}
}
