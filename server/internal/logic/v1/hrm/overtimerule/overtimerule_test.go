package overtimerule

import (
	"context"
	"testing"

	"github.com/jzero-io/jzero/core/stores/condition"

	coreSvc "github.com/jzero-io/jzero-admin/core-engine/svc"
	orglogic "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/org"
	attendancegrouplogic "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/attendancegroup"
	attendancegrouptypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/attendancegroup"
	orgtypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/org"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_overtime_rule"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/overtimerule"
)

func newSvcCtx(t *testing.T) *svc.ServiceContext {
	t.Helper()
	env := svctest.New(t)
	return &svc.ServiceContext{
		ServiceContext: &coreSvc.ServiceContext{SqlxConn: env.SqlConn},
		Model:          env.Model,
	}
}

func policy(approval, punch int64, multiplier float64) types.DayOvertimePolicy {
	return types.DayOvertimePolicy{ApprovalRequired: approval, PunchRequired: punch, CompMultiplier: multiplier}
}

func addReq(name string) *types.AddRequest {
	return &types.AddRequest{
		RuleName:       name,
		Workday:        policy(1, 1, 1),
		Restday:        policy(1, 0, 1.5),
		Holiday:        policy(0, 0, 3),
		MinUnitMinutes: 30,
		Status:         "1",
	}
}

// TestAddValidation 覆盖 Add 入参校验: 名称/倍数边界/最小单位/月上限/考勤组绑定.
func TestAddValidation(t *testing.T) {
	ctx := context.Background()
	svcCtx := newSvcCtx(t)

	cases := []struct {
		name string
		mut  func(r *types.AddRequest)
	}{
		{"empty name", func(r *types.AddRequest) { r.RuleName = "" }},
		{"zero multiplier", func(r *types.AddRequest) { r.Restday.CompMultiplier = 0 }},
		{"negative multiplier", func(r *types.AddRequest) { r.Holiday.CompMultiplier = -1 }},
		{"invalid min unit", func(r *types.AddRequest) { r.MinUnitMinutes = 20 }},
		{"negative monthly limit", func(r *types.AddRequest) { r.MonthlyLimitHours = -1 }},
		{"bad status", func(r *types.AddRequest) { r.Status = "2" }},
		{"unknown group binding", func(r *types.AddRequest) { r.GroupUuids = []string{"no-such-group"} }},
	}
	for _, tc := range cases {
		req := addReq("规则-" + tc.name)
		tc.mut(req)
		if _, err := NewAdd(ctx, svcCtx, nil).Add(req); err == nil {
			t.Errorf("%s: expected error, got nil", tc.name)
		}
	}

	// valid request inserts
	if _, err := NewAdd(ctx, svcCtx, nil).Add(addReq("默认规则")); err != nil {
		t.Fatalf("valid add: %v", err)
	}
	rows, err := svcCtx.Model.HrmOvertimeRule.FindByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_overtime_rule.RuleName, "默认规则").Build()...)
	if err != nil {
		t.Fatalf("find inserted: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	// default min unit = 30 when omitted
	req := addReq("默认最小单位")
	req.MinUnitMinutes = 0
	req.Status = ""
	if _, err := NewAdd(ctx, svcCtx, nil).Add(req); err != nil {
		t.Fatalf("default add: %v", err)
	}
	row, err := svcCtx.Model.HrmOvertimeRule.FindOneByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_overtime_rule.RuleName, "默认最小单位").Build()...)
	if err != nil {
		t.Fatalf("find default row: %v", err)
	}
	if row.MinUnitMinutes != 30 {
		t.Errorf("min unit default = %d, want 30", row.MinUnitMinutes)
	}
	if row.WorkdayApprovalRequired != 1 || row.RestdayPunchRequired != 0 || row.HolidayCompMultiplier != 3 {
		t.Errorf("policy fields not persisted: %+v", row)
	}
}

// TestEditDeleteToggle CRUD 轮转: edit 覆盖字段/toggle 启停/delete 批删/编辑不存在报错.
func TestEditDeleteToggle(t *testing.T) {
	ctx := context.Background()
	svcCtx := newSvcCtx(t)

	if _, err := NewAdd(ctx, svcCtx, nil).Add(addReq("规则A")); err != nil {
		t.Fatalf("add: %v", err)
	}
	row, err := svcCtx.Model.HrmOvertimeRule.FindOneByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_overtime_rule.RuleName, "规则A").Build()...)
	if err != nil {
		t.Fatalf("find: %v", err)
	}

	// edit not-found
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{Uuid: "missing", RuleName: "x"}); err == nil {
		t.Error("edit missing: expected error")
	}

	editReq := &types.EditRequest{
		Uuid:             row.Uuid,
		RuleName:         "规则A改",
		Workday:          policy(0, 1, 1.25),
		Restday:          policy(1, 0, 1.5),
		Holiday:          policy(0, 0, 3),
		MinUnitMinutes:   60,
		MonthlyLimitHours: 36,
		Status:           "0",
		Remark:           "改",
	}
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(editReq); err != nil {
		t.Fatalf("edit: %v", err)
	}
	updated, err := svcCtx.Model.HrmOvertimeRule.FindOneByUuid(ctx, nil, row.Uuid)
	if err != nil {
		t.Fatalf("re-find: %v", err)
	}
	if updated.RuleName != "规则A改" || updated.MinUnitMinutes != 60 || updated.MonthlyLimitHours != 36 ||
		updated.Status != "0" || updated.WorkdayApprovalRequired != 0 || updated.WorkdayCompMultiplier != 1.25 {
		t.Errorf("edit not applied: %+v", updated)
	}

	// toggle back on
	if _, err := NewToggleStatus(ctx, svcCtx, nil).ToggleStatus(&types.ToggleStatusRequest{Uuid: row.Uuid, Status: "1"}); err != nil {
		t.Fatalf("toggle: %v", err)
	}
	if _, err := NewToggleStatus(ctx, svcCtx, nil).ToggleStatus(&types.ToggleStatusRequest{Uuid: row.Uuid, Status: "9"}); err == nil {
		t.Error("toggle bad status: expected error")
	}
	toggled, _ := svcCtx.Model.HrmOvertimeRule.FindOneByUuid(ctx, nil, row.Uuid)
	if toggled.Status != "1" {
		t.Errorf("toggle status = %s, want 1", toggled.Status)
	}

	// list filter
	listResp, err := NewList(ctx, svcCtx, nil).List(&types.ListRequest{PageRequest: types.PageRequest{Current: 1, Size: 10}, RuleName: "规则A"})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if listResp.Total != 1 || len(listResp.Records) != 1 || listResp.Records[0].RuleName != "规则A改" {
		t.Errorf("list result unexpected: %+v", listResp)
	}
	if listResp.Records[0].Workday.CompMultiplier != 1.25 || listResp.Records[0].MinUnitMinutes != 60 {
		t.Errorf("record mapping wrong: %+v", listResp.Records[0])
	}

	// delete
	if _, err := NewDelete(ctx, svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{row.Uuid}}); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := svcCtx.Model.HrmOvertimeRule.FindOneByUuid(ctx, nil, row.Uuid); err != hrm_overtime_rule.ErrNotFound {
		t.Errorf("deleted row still found: %v", err)
	}
}

// seedOvertimeEnvs 建一个组织（供组织默认规则回退用）+ 一个考勤组（供绑定校验/解析用），返回 (orgUuid, groupUuid).
func seedOvertimeGroup(t *testing.T, svcCtx *svc.ServiceContext) (string, string) {
	t.Helper()
	ctx := context.Background()
	if _, err := orglogic.NewAdd(ctx, svcCtx, nil).Add(&orgtypes.AddRequest{OrgName: "加班组织", OrgType: "2", OrderNum: 1, Status: "1"}); err != nil {
		t.Fatalf("add org: %v", err)
	}
	orgs, err := svcCtx.Model.HrmOrg.FindByCondition(ctx, nil, condition.NewChain().Build()...)
	if err != nil {
		t.Fatalf("find orgs: %v", err)
	}
	if len(orgs) != 1 {
		t.Fatalf("expected 1 org, got %d", len(orgs))
	}
	if _, err := attendancegrouplogic.NewAdd(ctx, svcCtx, nil).Add(&attendancegrouptypes.AddRequest{
		GroupName: "技术部考勤组",
		GroupType: "1",
		WorkDays:  []int64{1, 2, 3, 4, 5},
		StartTime: "09:00",
		EndTime:   "18:00",
		Status:    "1",
	}); err != nil {
		t.Fatalf("add attendance group: %v", err)
	}
	groups, err := svcCtx.Model.HrmAttendanceGroup.FindByCondition(ctx, nil,
		condition.NewChain().Build()...)
	if err != nil {
		t.Fatalf("find groups: %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	return orgs[0].Uuid, groups[0].Uuid
}

// TestResolveOvertimeRuleWorkedExamples ResolveOvertimeRule 消费契约 worked examples:
// 考勤组绑定规则优先 > 组织默认规则 > 全局默认规则 > 未找到 found=false; 停用规则不参与; 按 dayType 取当天策略.
func TestResolveOvertimeRuleWorkedExamples(t *testing.T) {
	ctx := context.Background()
	svcCtx := newSvcCtx(t)
	orgUuid, groupUuid := seedOvertimeGroup(t, svcCtx)

	// 全局兜底规则 (无考勤组绑定, 无 org): 调休倍数 1.5
	globalReq := addReq("全局兜底")
	globalReq.Restday.CompMultiplier = 1.5
	if _, err := NewAdd(ctx, svcCtx, nil).Add(globalReq); err != nil {
		t.Fatalf("add global: %v", err)
	}

	// 组织默认规则 (org_uuid=orgUuid, 未绑考勤组): 工作日免审批, 倍数 1.25
	orgReq := addReq("组织默认")
	orgReq.OrgUuid = orgUuid
	orgReq.Workday = policy(0, 1, 1.25)
	if _, err := NewAdd(ctx, svcCtx, nil).Add(orgReq); err != nil {
		t.Fatalf("add org rule: %v", err)
	}

	// 考勤组绑定规则: 休息日调休×2, 免打卡
	groupReq := addReq("考勤组规则")
	groupReq.GroupUuids = []string{groupUuid}
	groupReq.Restday = policy(1, 0, 2)
	if _, err := NewAdd(ctx, svcCtx, nil).Add(groupReq); err != nil {
		t.Fatalf("add group rule: %v", err)
	}

	// ① 命中考勤组绑定规则 (restday → 倍数 2)
	resp, err := ResolveOvertimeRule(ctx, svcCtx, groupUuid, "", DayTypeRestday)
	if err != nil {
		t.Fatalf("resolve group: %v", err)
	}
	if !resp.Found || resp.Source != RuleSourceGroup || resp.SourceName != "考勤组规则" {
		t.Errorf("group resolve: %+v", resp)
	}
	if resp.CompMultiplier != 2 || resp.PunchRequired {
		t.Errorf("group restday policy: %+v", resp)
	}

	// ② 未绑定考勤组 → 组织默认 (workday → 免审批, 倍数 1.25)
	resp, err = ResolveOvertimeRule(ctx, svcCtx, "", orgUuid, DayTypeWorkday)
	if err != nil {
		t.Fatalf("resolve org: %v", err)
	}
	if !resp.Found || resp.Source != RuleSourceOrg || resp.SourceName != "组织默认" {
		t.Errorf("org resolve: %+v", resp)
	}
	if resp.ApprovalRequired || resp.CompMultiplier != 1.25 {
		t.Errorf("org workday policy: %+v", resp)
	}

	// ③ 无组织 → 全局兜底 (holiday → 免审批 倍数 3)
	resp, err = ResolveOvertimeRule(ctx, svcCtx, "", "", DayTypeHoliday)
	if err != nil {
		t.Fatalf("resolve global: %v", err)
	}
	if !resp.Found || resp.Source != RuleSourceGlobal || resp.CompMultiplier != 3 || resp.ApprovalRequired {
		t.Errorf("global resolve: %+v", resp)
	}

	// ④ 停用规则不参与解析: 停用考勤组规则后回落全局
	if _, err := NewToggleStatus(ctx, svcCtx, nil).ToggleStatus(&types.ToggleStatusRequest{Uuid: groupUuidToRuleUuid(t, ctx, svcCtx, "考勤组规则"), Status: "0"}); err != nil {
		t.Fatalf("disable group rule: %v", err)
	}
	resp, err = ResolveOvertimeRule(ctx, svcCtx, groupUuid, "", DayTypeRestday)
	if err != nil {
		t.Fatalf("resolve after disable: %v", err)
	}
	if resp.SourceName != "全局兜底" || resp.CompMultiplier != 1.5 {
		t.Errorf("fallback after disable: %+v", resp)
	}
}

func groupUuidToRuleUuid(t *testing.T, ctx context.Context, svcCtx *svc.ServiceContext, ruleName string) string {
	t.Helper()
	row, err := svcCtx.Model.HrmOvertimeRule.FindOneByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_overtime_rule.RuleName, ruleName).Build()...)
	if err != nil {
		t.Fatalf("find rule %s: %v", ruleName, err)
	}
	return row.Uuid
}

// TestResolveDayTypeInvalid 非法当天类型报错.
func TestResolveDayTypeInvalid(t *testing.T) {
	ctx := context.Background()
	svcCtx := newSvcCtx(t)
	_, err := ResolveOvertimeRule(ctx, svcCtx, "", "", "weekday")
	if err == nil {
		t.Fatal("invalid dayType: expected error")
	}
}
