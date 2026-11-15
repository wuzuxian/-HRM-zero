package org

import (
	"bytes"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"

	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/org"
)

func newCtx(t *testing.T) (*svc.ServiceContext, *svctest.Env) {
	t.Helper()
	env := svctest.New(t)
	return &svc.ServiceContext{Model: env.Model}, env
}

// findByName walks a tree and returns the first node with the given name.
func findByName(nodes []types.HrmOrg, name string) *types.HrmOrg {
	for i := range nodes {
		if nodes[i].OrgName == name {
			return &nodes[i]
		}
		if found := findByName(nodes[i].Children, name); found != nil {
			return found
		}
	}
	return nil
}

// TestOrgTreeCrud walks the primary user story: 新增组织 → 新增下级 → 编辑 →
// 删除/批量删除, verifying the tree-shaped list in between.
func TestOrgTreeCrud(t *testing.T) {
	svcCtx, env := newCtx(t)
	ctx := context.Background()

	// 新增组织: root company
	if _, err := NewAdd(ctx, svcCtx, nil).Add(addRoot("总公司", "1", 1)); err != nil {
		t.Fatalf("add root company: %v", err)
	}

	// tree initially has exactly one root and no children
	tree := mustList(t, svcCtx, types.ListRequest{})
	if len(tree) != 1 || tree[0].OrgName != "总公司" || tree[0].Children != nil {
		t.Fatalf("unexpected tree after root add: %+v", tree)
	}
	rootUuid := tree[0].Uuid

	// 新增下级: department under company
	if _, err := NewAddChild(ctx, svcCtx, nil).AddChild(&types.AddChildRequest{
		ParentUuid: rootUuid,
		OrgName:    "研发部",
		OrgType:    "2",
		Status:     "1",
		OrderNum:   1,
	}); err != nil {
		t.Fatalf("add child dept: %v", err)
	}

	// Add with explicit ParentUuid behaves like add-child (grandchild)
	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		ParentUuid: findByName(mustList(t, svcCtx, types.ListRequest{}), "研发部").Uuid,
		OrgName:    "前端组",
		OrgType:    "2",
		Status:     "1",
	}); err != nil {
		t.Fatalf("add grandchild via add: %v", err)
	}

	tree = mustList(t, svcCtx, types.ListRequest{})
	dept := findByName(tree, "研发部")
	if dept == nil || len(dept.Children) != 1 || dept.Children[0].OrgName != "前端组" {
		t.Fatalf("expected 研发部>前端组 nesting, got: %+v", tree)
	}

	// 编辑: rename leaf, change type/status/leader/order
	leaf := dept.Children[0]
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid:     leaf.Uuid,
		OrgName:  "前端中心",
		OrgType:  "2",
		Leader:   "张三",
		OrderNum: 9,
		Status:   "1",
	}); err != nil {
		t.Fatalf("edit leaf: %v", err)
	}
	tree = mustList(t, svcCtx, types.ListRequest{})
	edited := findByName(tree, "前端中心")
	if edited == nil || edited.Leader != "张三" || edited.OrderNum != 9 {
		t.Fatalf("edit not applied: %+v", tree)
	}
	if findByName(tree, "前端组") != nil {
		t.Fatalf("old name still present after edit: %+v", tree)
	}

	// 删除: node with children must be rejected
	if _, err := NewDelete(ctx, svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{rootUuid}}); err == nil {
		t.Fatalf("deleting node with children should fail")
	}

	// 删除 leaf succeeds
	if _, err := NewDelete(ctx, svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{edited.Uuid}}); err != nil {
		t.Fatalf("delete leaf: %v", err)
	}

	// 批量删除 remaining
	_ = env
	tree = mustList(t, svcCtx, types.ListRequest{})
	dept = findByName(tree, "研发部")
	var uuids []string
	for _, child := range dept.Children {
		uuids = append(uuids, child.Uuid)
	}
	uuids = append(uuids, dept.Uuid)
	if _, err := NewDelete(ctx, svcCtx, nil).Delete(&types.DeleteRequest{Uuids: uuids}); err != nil {
		t.Fatalf("batch delete: %v", err)
	}
	if tree := mustList(t, svcCtx, types.ListRequest{}); len(tree) != 1 || len(tree[0].Children) != 0 {
		t.Fatalf("expected only root left, got: %+v", tree)
	}
}

// TestOrgSiblingOrder verifies 显示顺序 ordering of siblings in the tree.
func TestOrgSiblingOrder(t *testing.T) {
	svcCtx, _ := newCtx(t)
	ctx := context.Background()

	if _, err := NewAdd(ctx, svcCtx, nil).Add(addRoot("顺序公司", "1", 1)); err != nil {
		t.Fatalf("add root: %v", err)
	}
	rootUuid := mustList(t, svcCtx, types.ListRequest{})[0].Uuid

	for _, c := range []struct {
		name  string
		order int64
	}{{"丙部门", 30}, {"甲部门", 10}, {"乙部门", 20}} {
		if _, err := NewAddChild(ctx, svcCtx, nil).AddChild(&types.AddChildRequest{
			ParentUuid: rootUuid, OrgName: c.name, OrgType: "2", Status: "1", OrderNum: c.order,
		}); err != nil {
			t.Fatalf("add child %s: %v", c.name, err)
		}
	}

	tree := mustList(t, svcCtx, types.ListRequest{})
	children := tree[0].Children
	if len(children) != 3 ||
		children[0].OrgName != "甲部门" || children[1].OrgName != "乙部门" || children[2].OrgName != "丙部门" {
		t.Fatalf("siblings not ordered by order_num: %+v", children)
	}
}

// TestOrgTreeFilter verifies the 组织名称/组织类型/组织状态 filters keep the
// matched node together with its ancestor chain.
func TestOrgTreeFilter(t *testing.T) {
	svcCtx, _ := newCtx(t)
	ctx := context.Background()

	if _, err := NewAdd(ctx, svcCtx, nil).Add(addRoot("过滤公司", "1", 1)); err != nil {
		t.Fatalf("add root: %v", err)
	}
	rootUuid := mustList(t, svcCtx, types.ListRequest{})[0].Uuid
	if _, err := NewAddChild(ctx, svcCtx, nil).AddChild(&types.AddChildRequest{
		ParentUuid: rootUuid, OrgName: "市场部", OrgType: "2", Status: "1",
	}); err != nil {
		t.Fatalf("add child: %v", err)
	}

	// by name: 市场部 keeps its parent 公司
	tree := mustList(t, svcCtx, types.ListRequest{OrgName: "市场"})
	if len(tree) != 1 || tree[0].OrgName != "过滤公司" ||
		len(tree[0].Children) != 1 || tree[0].Children[0].OrgName != "市场部" {
		t.Fatalf("name filter should keep ancestor chain: %+v", tree)
	}

	// by type: only the company matches (departments excluded)
	tree = mustList(t, svcCtx, types.ListRequest{OrgType: "1"})
	if len(tree) != 1 || len(tree[0].Children) != 0 {
		t.Fatalf("type filter should return only the company: %+v", tree)
	}
}

// TestOrgImportExport round-trips the excelize import/export: build a workbook
// in memory, import it, verify the tree (including parent resolution by
// 上级组织 name), then export and re-read the rows.
func TestOrgImportExport(t *testing.T) {
	svcCtx, _ := newCtx(t)
	ctx := context.Background()

	// pre-existing org referenced by the import file
	if _, err := NewAdd(ctx, svcCtx, nil).Add(addRoot("导入公司", "1", 1)); err != nil {
		t.Fatalf("add root: %v", err)
	}

	xlsx := buildImportWorkbook(t, [][]string{
		{"人事部", "部门", "李四", "10", "开启", "导入公司"},
		{"招聘组", "部门", "王五", "20", "关闭", "人事部"},
	})
	content := workbookToBase64(t, xlsx)

	resp, err := NewImport(ctx, svcCtx, nil).Import(&types.ImportRequest{
		FileName: "orgs.xlsx",
		Content:  content,
	})
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if resp.SuccessCount != 2 {
		t.Fatalf("expected 2 imported, got %d", resp.SuccessCount)
	}

	tree := mustList(t, svcCtx, types.ListRequest{})
	hr := findByName(tree, "人事部")
	if hr == nil || hr.ParentUuid != tree[0].Uuid || len(hr.Children) != 1 {
		t.Fatalf("imported tree wrong: %+v", tree)
	}
	if hr.Children[0].OrgName != "招聘组" || hr.Children[0].Status != "0" {
		t.Fatalf("imported child fields wrong: %+v", hr.Children[0])
	}

	// export decodes back into the same rows
	exported, err := NewExport(ctx, svcCtx, nil).Export()
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	rows := readWorkbookRows(t, exported.Content)
	// header + root + 2 imported
	if len(rows) != 4 {
		t.Fatalf("expected 4 exported rows (header+3), got %d: %v", len(rows), rows)
	}
	if rows[0][0] != "组织名称" {
		t.Fatalf("export header wrong: %v", rows[0])
	}
	var sawClosedChild bool
	for _, r := range rows[1:] {
		if r[0] == "招聘组" && r[1] == "部门" && r[5] == "人事部" {
			sawClosedChild = true
		}
	}
	if !sawClosedChild {
		t.Fatalf("exported rows missing 招聘组 with parent 人事部: %v", rows)
	}
}

// TestOrgImportRejectsUnknownParent keeps import honest: a row referencing a
// non-existent 上级组织 must fail instead of silently orphaning nodes.
func TestOrgImportRejectsUnknownParent(t *testing.T) {
	svcCtx, _ := newCtx(t)
	ctx := context.Background()

	xlsx := buildImportWorkbook(t, [][]string{
		{"孤儿部门", "部门", "", "1", "开启", "不存在的组织"},
	})
	_, err := NewImport(ctx, svcCtx, nil).Import(&types.ImportRequest{
		FileName: "orgs.xlsx",
		Content:  workbookToBase64(t, xlsx),
	})
	if err == nil {
		t.Fatalf("import with unknown parent should fail")
	}
}

// TestOrgAddRejectsUnknownParent mirrors the same guard on 新增组织.
func TestOrgAddRejectsUnknownParent(t *testing.T) {
	svcCtx, _ := newCtx(t)
	ctx := context.Background()

	req := addRoot("无父母司", "1", 1)
	req.ParentUuid = uuid.New().String()
	if _, err := NewAdd(ctx, svcCtx, nil).Add(req); err == nil {
		t.Fatalf("add with unknown parent should fail")
	}
}

func addRoot(name, orgType string, orderNum int64) *types.AddRequest {
	return &types.AddRequest{OrgName: name, OrgType: orgType, OrderNum: orderNum, Status: "1"}
}

func mustList(t *testing.T, svcCtx *svc.ServiceContext, req types.ListRequest) []types.HrmOrg {
	t.Helper()
	resp, err := NewList(context.Background(), svcCtx, nil).List(&req)
	if err != nil {
		t.Fatalf("list tree: %v", err)
	}
	return resp.Records
}

func buildImportWorkbook(t *testing.T, rows [][]string) *excelize.File {
	t.Helper()
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	for i, row := range rows {
		cell, err := excelize.CoordinatesToCellName(1, i+2) // data starts at row 2 (header at 1)
		if err != nil {
			t.Fatalf("coordinates: %v", err)
		}
		if err = f.SetSheetRow(sheet, cell, &row); err != nil {
			t.Fatalf("set row: %v", err)
		}
	}
	// header row must match the import template
	header := []string{"组织名称", "组织类型", "负责人", "显示顺序", "组织状态", "上级组织"}
	if err := f.SetSheetRow(sheet, "A1", &header); err != nil {
		t.Fatalf("set header: %v", err)
	}
	return f
}

func workbookToBase64(t *testing.T, f *excelize.File) string {
	t.Helper()
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		t.Fatalf("write workbook: %v", err)
	}
	return encodeBase64(buf.Bytes())
}

func readWorkbookRows(t *testing.T, base64Content string) [][]string {
	t.Helper()
	f, err := excelize.OpenReader(decodeBase64(t, base64Content))
	if err != nil {
		t.Fatalf("open exported workbook: %v", err)
	}
	rows, err := f.GetRows(f.GetSheetName(0))
	if err != nil {
		t.Fatalf("read exported rows: %v", err)
	}
	return rows
}
