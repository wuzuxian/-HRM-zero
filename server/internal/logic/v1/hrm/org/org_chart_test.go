package org

import (
	"context"
	"errors"
	"testing"

	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/org"
)

// findChartNode walks the chart tree and returns the first node with the given name.
func findChartNode(nodes []types.HrmOrgChartNode, name string) *types.HrmOrgChartNode {
	for i := range nodes {
		if nodes[i].OrgName == name {
			return &nodes[i]
		}
		if found := findChartNode(nodes[i].Children, name); found != nil {
			return found
		}
	}
	return nil
}

// mustOrgChart runs the OrgChart logic with the given 本级人数 source.
func mustOrgChart(
	t *testing.T,
	svcCtx *svc.ServiceContext,
	counts func(ctx context.Context, orgUuids []string) (map[string]int64, error),
) []types.HrmOrgChartNode {
	t.Helper()
	lc := NewOrgChart(context.Background(), svcCtx, nil)
	if counts != nil {
		lc.memberCounts = counts
	}
	resp, err := lc.OrgChart()
	if err != nil {
		t.Fatalf("org chart: %v", err)
	}
	return resp.Records
}

// TestOrgChartCounts locks the 人数汇总口径: 本级人数 comes from the member-count
// source (在册员工数 per org), 汇总人数 = 本级人数 + 全部下级组织的汇总人数,
// accumulated bottom-up over the whole subtree.
func TestOrgChartCounts(t *testing.T) {
	svcCtx, _ := newCtx(t)
	ctx := context.Background()

	if _, err := NewAdd(ctx, svcCtx, nil).Add(addRoot("架构公司", "1", 1)); err != nil {
		t.Fatalf("add root: %v", err)
	}
	rootUuid := mustList(t, svcCtx, types.ListRequest{})[0].Uuid
	if _, err := NewAddChild(ctx, svcCtx, nil).AddChild(&types.AddChildRequest{
		ParentUuid: rootUuid, OrgName: "研发部", OrgType: "2", Leader: "张三", Status: "1", OrderNum: 1,
	}); err != nil {
		t.Fatalf("add 研发部: %v", err)
	}
	deptUuid := mustList(t, svcCtx, types.ListRequest{})[0].Children[0].Uuid
	if _, err := NewAddChild(ctx, svcCtx, nil).AddChild(&types.AddChildRequest{
		ParentUuid: deptUuid, OrgName: "前端组", OrgType: "2", Leader: "李四", Status: "1", OrderNum: 1,
	}); err != nil {
		t.Fatalf("add 前端组: %v", err)
	}
	if _, err := NewAddChild(ctx, svcCtx, nil).AddChild(&types.AddChildRequest{
		ParentUuid: rootUuid, OrgName: "市场部", OrgType: "2", Status: "1", OrderNum: 2,
	}); err != nil {
		t.Fatalf("add 市场部: %v", err)
	}

	// worked example (spec of the 口径), independent of the implementation:
	// 本级: 公司=1 研发部=5 前端组=3 市场部=2
	// 汇总: 前端组=3 研发部=5+3=8 市场部=2 公司=1+8+2=11
	direct := map[string]int64{}
	byName := map[string]string{"架构公司": rootUuid, "研发部": deptUuid}
	tree := mustList(t, svcCtx, types.ListRequest{})
	byName["前端组"] = findByName(tree, "研发部").Children[0].Uuid
	byName["市场部"] = findByName(tree, "市场部").Uuid
	for name, n := range map[string]int64{"架构公司": 1, "研发部": 5, "前端组": 3, "市场部": 2} {
		direct[byName[name]] = n
	}

	records := mustOrgChart(t, svcCtx, func(_ context.Context, orgUuids []string) (map[string]int64, error) {
		out := make(map[string]int64, len(orgUuids))
		for _, u := range orgUuids {
			out[u] = direct[u]
		}
		return out, nil
	})

	root := findChartNode(records, "架构公司")
	if root == nil {
		t.Fatalf("missing root: %+v", records)
	}
	dept := findChartNode(records, "研发部")
	leaf := findChartNode(records, "前端组")
	mkt := findChartNode(records, "市场部")

	// 负责人/类型 carried through so the chart can render them
	if root.Leader != "" || dept.Leader != "张三" || leaf.Leader != "李四" {
		t.Fatalf("leader not carried: root=%q dept=%q leaf=%q", root.Leader, dept.Leader, leaf.Leader)
	}
	if root.OrgType != "1" || dept.OrgType != "2" {
		t.Fatalf("org type not carried: %+v", records)
	}

	if root.MemberCount != 1 || root.TotalMemberCount != 11 {
		t.Fatalf("公司 counts wrong: got member=%d total=%d, want 1/11", root.MemberCount, root.TotalMemberCount)
	}
	if dept.MemberCount != 5 || dept.TotalMemberCount != 8 {
		t.Fatalf("研发部 counts wrong: got member=%d total=%d, want 5/8", dept.MemberCount, dept.TotalMemberCount)
	}
	if leaf.MemberCount != 3 || leaf.TotalMemberCount != 3 {
		t.Fatalf("前端组 counts wrong: got member=%d total=%d, want 3/3", leaf.MemberCount, leaf.TotalMemberCount)
	}
	if mkt.MemberCount != 2 || mkt.TotalMemberCount != 2 {
		t.Fatalf("市场部 counts wrong: got member=%d total=%d, want 2/2", mkt.MemberCount, mkt.TotalMemberCount)
	}
}

// TestOrgChartSourceError surfaces source failures instead of rendering a wrong chart.
func TestOrgChartSourceError(t *testing.T) {
	svcCtx, _ := newCtx(t)
	if _, err := NewAdd(context.Background(), svcCtx, nil).Add(addRoot("报错公司", "1", 1)); err != nil {
		t.Fatalf("add root: %v", err)
	}
	lc := NewOrgChart(context.Background(), svcCtx, nil)
	lc.memberCounts = func(context.Context, []string) (map[string]int64, error) {
		return nil, errors.New("boom")
	}
	if _, err := lc.OrgChart(); err == nil {
		t.Fatalf("source error should propagate")
	}
}

// TestOrgChartDefaultCountsNoEmployeeTable exercises the default member-count
// source against the real migrations: until 员工档案 (hrm_employee) lands, the
// table is missing and the count query degrades to zeros instead of failing
// the chart.
func TestOrgChartDefaultCountsNoEmployeeTable(t *testing.T) {
	svcCtx, env := newCtx(t)
	if _, err := NewAdd(context.Background(), svcCtx, nil).Add(addRoot("空表公司", "1", 1)); err != nil {
		t.Fatalf("add root: %v", err)
	}

	// the raw count query on a database without hrm_employee degrades to zeros
	counts, err := countEmployeesByOrg(context.Background(), env.SqlConn, []string{"any-org"})
	if err != nil {
		t.Fatalf("countEmployeesByOrg should tolerate missing hrm_employee: %v", err)
	}
	if len(counts) != 0 {
		t.Fatalf("expected empty counts, got %+v", counts)
	}

	records := mustOrgChart(t, svcCtx, nil) // default source
	root := findChartNode(records, "空表公司")
	if root == nil {
		t.Fatalf("missing root: %+v", records)
	}
	if root.MemberCount != 0 || root.TotalMemberCount != 0 {
		t.Fatalf("default source should yield zeros without hrm_employee, got %+v", root)
	}
}

// TestOrgChartEmpty verifies an empty 组织 table yields an empty chart.
func TestOrgChartEmpty(t *testing.T) {
	svcCtx, _ := newCtx(t)
	records := mustOrgChart(t, svcCtx, nil)
	if len(records) != 0 {
		t.Fatalf("expected empty chart, got %+v", records)
	}
}
