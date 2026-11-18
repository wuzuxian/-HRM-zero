package attendancegroup

import (
	"context"
	"testing"

	"github.com/jzero-io/jzero/core/stores/condition"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	coreSvc "github.com/jzero-io/jzero-admin/core-engine/svc"
	"github.com/jzero-io/jzero-admin/server/internal/config"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	orglogic "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/org"
	employeelogic "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/employee"
	orgtypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/org"
	employeetypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/employee"
	configlogic "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/attendanceconfig"
	configtypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/attendanceconfig"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/attendancegroup"
)

func newSvcCtx(t *testing.T) *svc.ServiceContext {
	t.Helper()
	env := svctest.New(t)
	// 员工档案 logic 使用 SqlxConn 做事务; attendancegroup/config 逻辑只依赖 Model。
	return &svc.ServiceContext{
		ServiceContext: &coreSvc.ServiceContext{SqlxConn: env.SqlConn},
		ConfigCenter:   config.NewStaticConfigCenter(config.Config{Hrm: config.HrmConf{InitialUserPassword: "test-pwd-123"}}),
		Model:          env.Model,
	}
}

// seedEnv 建组织树 公司A(orgA) → 技术部(orgTech) → 后端组(orgBackend)（orgSales 平级），
// 并返回新增组织的 uuid 查询器（通过 org list tree 反查）。
type seeded struct {
	orgA, orgTech, orgBackend, orgSales string
	empZhang, empLi, empWang, empZhao   string
}

func seedEnv(t *testing.T, svcCtx *svc.ServiceContext) seeded {
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
		uuid, err := findOrgUuid(ctx, svcCtx, name)
		if err != nil {
			t.Fatalf("find org %s: %v", name, err)
		}
		return uuid
	}

	s := seeded{}
	s.orgA = addOrg("", "公司A")
	s.orgTech = addOrg(s.orgA, "技术部")
	s.orgBackend = addOrg(s.orgTech, "后端组")
	s.orgSales = addOrg(s.orgA, "销售部")

	addEmp := func(orgUuid, no, name string) string {
		req := &employeetypes.AddRequest{Name: name, EmployeeNo: no, OrgUuid: orgUuid, PersonStatus: "1"}
		if _, err := employeelogic.NewAdd(ctx, svcCtx, nil).Add(req); err != nil {
			t.Fatalf("add employee %s: %v", name, err)
		}
		row, err := svcCtx.Model.HrmEmployee.FindOneByCondition(ctx, nil, condition.NewChain().Equal(hrm_employee.EmployeeNo, no).Build()...)
		if err != nil {
			t.Fatalf("find employee %s: %v", no, err)
		}
		return row.Uuid
	}

	s.empZhang = addEmp(s.orgBackend, "E0001", "张三") // 后端组
	s.empLi = addEmp(s.orgBackend, "E0002", "李四")    // 后端组
	s.empWang = addEmp(s.orgSales, "E0003", "王五")     // 销售部
	s.empZhao = addEmp(s.orgTech, "E0004", "赵六")      // 技术部（不在任何考勤组）
	return s
}

func findOrgUuid(ctx context.Context, svcCtx *svc.ServiceContext, name string) (string, error) {
	rows, err := svcCtx.Model.HrmOrg.FindByCondition(ctx, nil)
	if err != nil {
		return "", err
	}
	for _, row := range rows {
		if row.OrgName == name {
			return row.Uuid, nil
		}
	}
	return "", nil
}

// addGroup 新增考勤组（固定班制 周一至周五 09:00-18:00）。
func addGroup(t *testing.T, ctx context.Context, svcCtx *svc.ServiceContext, name string, orgUuids, memberUuids []string) {
	t.Helper()
	req := &types.AddRequest{
		GroupName: name,
		GroupType: groupTypeFixed,
		WorkDays:  []int64{1, 2, 3, 4, 5},
		StartTime: "09:00",
		EndTime:   "18:00",
		OrgUuids:  orgUuids,
		MemberUuids: memberUuids,
		Managers:  []string{"王经理"},
		Status:    "1",
	}
	if _, err := NewAdd(ctx, svcCtx, nil).Add(req); err != nil {
		t.Fatalf("add attendance group %s: %v", name, err)
	}
}

// saveConfig 保存组织默认规则（经由 attendanceconfig 包的 Save 端点）。
func saveConfig(t *testing.T, ctx context.Context, svcCtx *svc.ServiceContext, orgUuid, startTime, endTime string, workDays []int64) {
	t.Helper()
	if _, err := configlogic.NewSave(ctx, svcCtx, nil).Save(&configtypes.SaveRequest{
		OrgUuid: orgUuid, StartTime: startTime, EndTime: endTime, WorkDays: workDays,
	}); err != nil {
		t.Fatalf("save attendance config for %s: %v", orgUuid, err)
	}
}

// ruleOf 执行生效规则查询并断言命中来源。
func ruleOf(t *testing.T, ctx context.Context, svcCtx *svc.ServiceContext, employeeUuid string) *types.GetEffectiveRuleResponse {
	t.Helper()
	resp, err := NewEffectiveRule(ctx, svcCtx, nil).EffectiveRule(&types.GetEffectiveRuleRequest{EmployeeUuid: employeeUuid})
	if err != nil {
		t.Fatalf("effective rule for %s: %v", employeeUuid, err)
	}
	return resp
}


// TestThreeLevelFallbackWorkedExamples locks the 三级回退 rule
// (考勤组 > 本部门规则 > 最近上级组织规则 > 无规则兜底) with worked examples:
//
//	组织树: 公司A → 技术部 → 后端组, 公司A → 销售部
//	考勤组「研发标准班」适用范围 = 部门[后端组] + 指定人员[王五(销售部)]
//	组织默认规则: 后端组 08:30-17:30 / 公司A 10:00-19:00
//
//	张三(后端组)     → 考勤组命中(部门=本部门)  source=group
//	王五(销售部)     → 考勤组命中(指定人员)     source=group
//	赵六(技术部)     → 无考勤组; 技术部无规则 → 上级公司A规则 10:00-19:00  source=org
//	李四(后端组,考勤组停用后) → 本部门规则 08:30-17:30                 source=org
//	独立公司B 员工   → 全空                     found=false
func TestThreeLevelFallbackWorkedExamples(t *testing.T) {
	svcCtx := newSvcCtx(t)
	ctx := context.Background()
	s := seedEnv(t, svcCtx)

	// 三级优先级第 1 级: 考勤组命中（部门含下级 + 指定人员 两条路径）
	addGroup(t, ctx, svcCtx, "研发标准班", []string{s.orgBackend}, []string{s.empWang})

	zhang := ruleOf(t, ctx, svcCtx, s.empZhang)
	if !zhang.Found || zhang.Source != ruleSourceGroup || zhang.SourceName != "研发标准班" {
		t.Fatalf("张三 expect group hit via 适用部门, got: %+v", zhang)
	}
	if zhang.StartTime != "09:00" || zhang.EndTime != "18:00" || len(zhang.WorkDays) != 5 || zhang.GroupType != groupTypeFixed {
		t.Fatalf("张三 expect 09:00-18:00 周一至五 固定班制, got: %+v", zhang)
	}

	wang := ruleOf(t, ctx, svcCtx, s.empWang)
	if !wang.Found || wang.Source != ruleSourceGroup || wang.SourceName != "研发标准班" {
		t.Fatalf("王五 expect group hit via 指定人员, got: %+v", wang)
	}

	// 第 2/3 级准备: 考勤组覆盖后端组, 组织默认规则 = 后端组 08:30-17:30, 公司A 10:00-19:00
	saveConfig(t, ctx, svcCtx, s.orgBackend, "08:30", "17:30", []int64{1, 2, 3, 4, 5})
	saveConfig(t, ctx, svcCtx, s.orgA, "10:00", "19:00", []int64{1, 2, 3, 4, 5})

	// 第 3 级: 赵六(技术部) — 不在考勤组, 技术部无规则, 回退最近上级组织(公司A)
	zhao := ruleOf(t, ctx, svcCtx, s.empZhao)
	if !zhao.Found || zhao.Source != ruleSourceOrg || zhao.StartTime != "10:00" || zhao.EndTime != "19:00" {
		t.Fatalf("赵六 expect fallback to 公司A rule 10:00-19:00, got: %+v", zhao)
	}

	// 第 2 级: 停用考勤组后, 李四(后端组) 命中本部门规则 08:30-17:30
	groups, err := NewList(ctx, svcCtx, nil).List(&types.ListRequest{PageRequest: types.PageRequest{Current: 1, Size: 10}})
	if err != nil {
		t.Fatalf("list groups: %v", err)
	}
	if len(groups.Records) != 1 {
		t.Fatalf("expect 1 group, got %d", len(groups.Records))
	}
	if _, err = NewToggleStatus(ctx, svcCtx, nil).ToggleStatus(&types.ToggleStatusRequest{
		Uuid: groups.Records[0].Uuid, Status: "0",
	}); err != nil {
		t.Fatalf("toggle group status: %v", err)
	}

	li := ruleOf(t, ctx, svcCtx, s.empLi)
	if !li.Found || li.Source != ruleSourceOrg || li.SourceUuid != s.orgBackend {
		t.Fatalf("李四 expect 本部门规则 hit after group disabled, got: %+v", li)
	}
	if li.StartTime != "08:30" || li.EndTime != "17:30" {
		t.Fatalf("李四 expect 08:30-17:30, got: %+v", li)
	}

	// 第 2 级回退到最近上级: 停用后端组规则 → 李四回退到 公司A 10:00-19:00
	rows, err := svcCtx.Model.HrmAttendanceConfig.FindByCondition(ctx, nil)
	if err != nil {
		t.Fatalf("list configs: %v", err)
	}
	for _, row := range rows {
		if row.OrgUuid == s.orgBackend {
			if _, err = configlogic.NewDelete(ctx, svcCtx, nil).Delete(&configtypes.DeleteRequest{Uuids: []string{row.Uuid}}); err != nil {
				t.Fatalf("delete backend config: %v", err)
			}
		}
	}
	li = ruleOf(t, ctx, svcCtx, s.empLi)
	if !li.Found || li.Source != ruleSourceOrg || li.SourceUuid != s.orgA || li.StartTime != "10:00" {
		t.Fatalf("李四 expect fallback to 公司A after 本部门规则 deleted, got: %+v", li)
	}

	// 兜底: 全空 → 无规则. 独立公司B + 员工钱七, 无考勤组/无规则
	bReq := &orgtypes.AddRequest{OrgName: "公司B", OrgType: "1", OrderNum: 2, Status: "1"}
	if _, err := orglogic.NewAdd(ctx, svcCtx, nil).Add(bReq); err != nil {
		t.Fatalf("add company B: %v", err)
	}
	orgB, _ := findOrgUuid(ctx, svcCtx, "公司B")
	if _, err := employeelogic.NewAdd(ctx, svcCtx, nil).Add(&employeetypes.AddRequest{
		Name: "钱七", EmployeeNo: "E0005", OrgUuid: orgB, PersonStatus: "1",
	}); err != nil {
		t.Fatalf("add 钱七: %v", err)
	}
	qian, err := svcCtx.Model.HrmEmployee.FindOneByCondition(ctx, nil, condition.NewChain().Equal(hrm_employee.EmployeeNo, "E0005").Build()...)
	if err != nil {
		t.Fatalf("find 钱七: %v", err)
	}
	resp := ruleOf(t, ctx, svcCtx, qian.Uuid)
	if resp.Found {
		t.Fatalf("钱七 expect no rule (全空兜底), got: %+v", resp)
	}

	// 启用考勤组后 王五 恢复考勤组命中 (启停影响解析)
	if _, err = NewToggleStatus(ctx, svcCtx, nil).ToggleStatus(&types.ToggleStatusRequest{
		Uuid: groups.Records[0].Uuid, Status: "1",
	}); err != nil {
		t.Fatalf("re-enable group: %v", err)
	}
	wang = ruleOf(t, ctx, svcCtx, s.empWang)
	if !wang.Found || wang.Source != ruleSourceGroup {
		t.Fatalf("王五 expect group hit again after re-enable, got: %+v", wang)
	}
}

// TestEffectiveRuleByEmployeeNo 生效规则查询支持按工号查询（页面无 uuid 时）.
func TestEffectiveRuleByEmployeeNo(t *testing.T) {
	svcCtx := newSvcCtx(t)
	ctx := context.Background()
	s := seedEnv(t, svcCtx)

	addGroup(t, ctx, svcCtx, "研发标准班", []string{s.orgTech}, nil)

	resp, err := NewEffectiveRule(ctx, svcCtx, nil).EffectiveRule(&types.GetEffectiveRuleRequest{EmployeeNo: "E0001"})
	if err != nil {
		t.Fatalf("effective rule by employeeNo: %v", err)
	}
	if !resp.Found || resp.Source != ruleSourceGroup {
		t.Fatalf("expect group hit by employeeNo, got: %+v", resp)
	}

	// 两个查询参数都缺 → 参数错误
	if _, err := NewEffectiveRule(ctx, svcCtx, nil).EffectiveRule(&types.GetEffectiveRuleRequest{}); err == nil {
		t.Fatal("effective rule without employee params should fail")
	}
	// 不存在的员工 → 不存在错误
	if _, err := NewEffectiveRule(ctx, svcCtx, nil).EffectiveRule(&types.GetEffectiveRuleRequest{EmployeeNo: "NOPE"}); err == nil {
		t.Fatal("effective rule for missing employee should fail")
	}
}

// TestGroupMemberExpansion 适用范围解析（部门含下级 + 指定人员, 去重）:
// 生效人数 = 考勤组部门(含下级)员工 ∪ 指定人员.
func TestGroupMemberExpansion(t *testing.T) {
	svcCtx := newSvcCtx(t)
	ctx := context.Background()
	s := seedEnv(t, svcCtx)

	// 部门[后端组] + 指定人员[王五,赵六]（赵六在技术部, 走指定人员路径）
	addGroup(t, ctx, svcCtx, "混合范围组", []string{s.orgBackend}, []string{s.empWang, s.empZhao})

	resp, err := NewMembers(ctx, svcCtx, nil).Members(&types.MemberRequest{})
	if err != nil {
		t.Fatalf("members: %v", err)
	}
	if resp.Total != 4 || len(resp.Members) != 4 {
		t.Fatalf("expect 4 members (张三/李四 via 部门含下级 + 王五/赵六 via 指定人员), got total=%d %+v", resp.Total, resp.Members)
	}
	seen := map[string]bool{}
	for _, m := range resp.Members {
		seen[m.Uuid] = true
	}
	for _, want := range []string{s.empZhang, s.empLi, s.empWang, s.empZhao} {
		if !seen[want] {
			t.Fatalf("member %s missing: %+v", want, resp.Members)
		}
	}
}
