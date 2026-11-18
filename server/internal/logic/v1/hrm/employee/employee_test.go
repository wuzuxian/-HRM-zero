package employee

import (
	"bytes"
	"context"
	"encoding/base64"
	"testing"

	"github.com/xuri/excelize/v2"

	coreSvc "github.com/jzero-io/jzero-admin/core-engine/svc"
	"github.com/jzero-io/jzero-admin/server/internal/config"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_org"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/employee"
)

func newCtx(t *testing.T) (*svc.ServiceContext, *svctest.Env) {
	t.Helper()
	env := svctest.New(t)
	// 员工档案 logic 使用 SqlxConn 做事务 (从表全删全插); org 逻辑只依赖 Model,
	// 这里比 org 测试多注入 SqlxConn, core-engine 其余字段保持 nil。
	// HRM-50 建档即开号: Add/Import 在事务内接开号接缝, 接缝读 ConfigCenter
	// (初始密码/默认角色码), 测试环境必须注入 StaticConfigCenter。
	return &svc.ServiceContext{
		ServiceContext: &coreSvc.ServiceContext{SqlxConn: env.SqlConn},
		ConfigCenter:   config.NewStaticConfigCenter(config.Config{Hrm: config.HrmConf{InitialUserPassword: "test-pwd-123"}}),
		Model:          env.Model,
	}, env
}

func addEmployee(t *testing.T, svcCtx *svc.ServiceContext, req *types.AddRequest) {
	t.Helper()
	if _, err := NewAdd(context.Background(), svcCtx, nil).Add(req); err != nil {
		t.Fatalf("add employee %s: %v", req.Name, err)
	}
}

func mustList(t *testing.T, svcCtx *svc.ServiceContext, req types.ListRequest) types.ListResponse {
	t.Helper()
	if req.Current == 0 {
		req.Current = 1
	}
	if req.Size == 0 {
		req.Size = 10
	}
	resp, err := NewList(context.Background(), svcCtx, nil).List(&req)
	if err != nil {
		t.Fatalf("list employees: %v", err)
	}
	return *resp
}

func mustDetail(t *testing.T, svcCtx *svc.ServiceContext, uuid string) types.HrmEmployee {
	t.Helper()
	resp, err := NewDetail(context.Background(), svcCtx, nil).Detail(&types.DetailRequest{Uuid: uuid})
	if err != nil {
		t.Fatalf("employee detail: %v", err)
	}
	return resp.Employee
}

func withChildren(name string) *types.AddRequest {
	return &types.AddRequest{
		Name:         name,
		PersonStatus: "1",
		Educations: []types.HrmEmployeeEducation{
			{School: "华中科技大学", Major: "软件工程", Education: "本科", StartDate: "2012-09-01", EndDate: "2016-06-30"},
			{School: "清华大学", Major: "软件工程", Education: "硕士", StartDate: "2016-09-01", EndDate: "2018-06-30"},
		},
		Works: []types.HrmEmployeeWork{
			{Company: "腾讯", Position: "开发工程师", StartDate: "2018-07-01", EndDate: "2020-12-31"},
		},
		Families: []types.HrmEmployeeFamily{
			{Name: "王母", Relation: "母亲", Phone: "13800000001", WorkUnit: "退休"},
		},
	}
}

// TestEmployeeAddAutoNo 覆盖工号自动生成: 空表→00000001, 连续+1, 指定工号已存在报错,
// 非数字历史编号不参与递增。
func TestEmployeeAddAutoNo(t *testing.T) {
	svcCtx, _ := newCtx(t)
	ctx := context.Background()

	addEmployee(t, svcCtx, &types.AddRequest{Name: "张三"})
	addEmployee(t, svcCtx, &types.AddRequest{Name: "李四"})
	list := mustList(t, svcCtx, types.ListRequest{})
	if list.Total != 2 || len(list.Records) != 2 {
		t.Fatalf("total=%d len=%d, want 2/2", list.Total, len(list.Records))
	}
	for _, rec := range list.Records {
		if rec.EmployeeNo != "00000001" && rec.EmployeeNo != "00000002" {
			t.Fatalf("auto employeeNo = %q, want 00000001/00000002", rec.EmployeeNo)
		}
		// HRM-50 建档即开号: 在职新增即产账号, 用户状态=已生成, 登录账号=工号
		if rec.UserStatus != "1" {
			t.Fatalf("userStatus = %q, want 1 (建档即开号)", rec.UserStatus)
		}
		if rec.LoginName != rec.EmployeeNo {
			t.Fatalf("loginName = %q, want %q (开号时用户名=工号)", rec.LoginName, rec.EmployeeNo)
		}
	}

	// 指定已存在工号 → 报错
	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{Name: "王五", EmployeeNo: "00000001"}); err == nil {
		t.Fatal("duplicate employeeNo should fail")
	}

	// 非数字历史编号 (导入旧账) 不参与递增: 指定后下一个自动工号仍按数字最大值+1
	addEmployee(t, svcCtx, &types.AddRequest{Name: "赵六", EmployeeNo: "HR-009"})
	addEmployee(t, svcCtx, &types.AddRequest{Name: "孙七", EmployeeNo: "00000009"})
	list = mustList(t, svcCtx, types.ListRequest{})
	for _, rec := range list.Records {
		if rec.Name == "孙七" && rec.EmployeeNo != "00000009" {
			t.Fatalf("explicit employeeNo lost: %+v", rec)
		}
	}
}

// TestEmployeeChildrenFullReplace 从表全删全插: add 落三表, edit 整体替换, 失败编辑
// 事务回滚不动旧行, delete 级联清空 (spec user story 9)。
func TestEmployeeChildrenFullReplace(t *testing.T) {
	svcCtx, _ := newCtx(t)
	ctx := context.Background()

	addEmployee(t, svcCtx, withChildren("张三"))
	list := mustList(t, svcCtx, types.ListRequest{})
	uuid := list.Records[0].Uuid

	detail := mustDetail(t, svcCtx, uuid)
	if len(detail.Educations) != 2 || len(detail.Works) != 1 || len(detail.Families) != 1 {
		t.Fatalf("children after add: %d/%d/%d, want 2/1/1",
			len(detail.Educations), len(detail.Works), len(detail.Families))
	}

	// edit: 教育经历 2→1, 工作经历 1→0, 家属 1→2 — 全删全插
	editReq := &types.EditRequest{
		Uuid: uuid,
		Name: "张三",
		Educations: []types.HrmEmployeeEducation{
			{School: "北京大学", Major: "计算机", Education: "博士", StartDate: "2018-09-01", EndDate: "2022-06-30"},
		},
		Families: []types.HrmEmployeeFamily{
			{Name: "张父", Relation: "父亲", Phone: "13800000002"},
			{Name: "张母", Relation: "母亲", Phone: "13800000003"},
		},
	}
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(editReq); err != nil {
		t.Fatalf("edit: %v", err)
	}
	detail = mustDetail(t, svcCtx, uuid)
	if len(detail.Educations) != 1 || detail.Educations[0].School != "北京大学" {
		t.Fatalf("educations after full-replace: %+v", detail.Educations)
	}
	if len(detail.Works) != 0 {
		t.Fatalf("works should be emptied, got %+v", detail.Works)
	}
	if len(detail.Families) != 2 {
		t.Fatalf("families after full-replace: %+v", detail.Families)
	}

	// 事务性: 编辑不存在的员工报错且不破坏旧数据
	badEdit := &types.EditRequest{Uuid: "no-such-uuid", Name: "张三", Educations: detail.Educations}
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(badEdit); err == nil {
		t.Fatal("edit nonexistent uuid should fail")
	}
	detail = mustDetail(t, svcCtx, uuid)
	if len(detail.Educations) != 1 || detail.Educations[0].School != "北京大学" {
		t.Fatalf("children corrupted by failed edit: %+v", detail.Educations)
	}

	// delete 级联清空三个从表
	if _, err := NewDelete(ctx, svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{uuid}}); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := NewDetail(ctx, svcCtx, nil).Detail(&types.DetailRequest{Uuid: uuid}); err == nil {
		t.Fatal("deleted employee detail should fail")
	}
	if n, err := svcCtx.Model.HrmEmployeeEducation.CountByCondition(ctx, nil); err != nil || n != 0 {
		t.Fatalf("education rows after cascade delete: n=%d err=%v", n, err)
	}
	if n, err := svcCtx.Model.HrmEmployeeWork.CountByCondition(ctx, nil); err != nil || n != 0 {
		t.Fatalf("work rows after cascade delete: n=%d err=%v", n, err)
	}
	if n, err := svcCtx.Model.HrmEmployeeFamily.CountByCondition(ctx, nil); err != nil || n != 0 {
		t.Fatalf("family rows after cascade delete: n=%d err=%v", n, err)
	}
}

// TestEmployeeListFilters 列表筛选: 工号/姓名/部门/人员状态 + 分页 (spec user story 11).
func TestEmployeeListFilters(t *testing.T) {
	svcCtx, _ := newCtx(t)
	ctx := context.Background()

	// seed orgs directly
	dept := &hrm_org.HrmOrg{Uuid: "org-dept-uuid", OrgName: "研发部", OrgType: "2", Status: "1"}
	other := &hrm_org.HrmOrg{Uuid: "org-other-uuid", OrgName: "市场部", OrgType: "2", Status: "1"}
	for _, org := range []*hrm_org.HrmOrg{dept, other} {
		if err := svcCtx.Model.HrmOrg.InsertV2(ctx, nil, org); err != nil {
			t.Fatalf("seed org: %v", err)
		}
	}

	addEmployee(t, svcCtx, &types.AddRequest{Name: "张三", EmployeeNo: "00000001", OrgUuid: dept.Uuid, PersonStatus: "1"})
	addEmployee(t, svcCtx, &types.AddRequest{Name: "李四", EmployeeNo: "00000002", OrgUuid: other.Uuid, PersonStatus: "5"})
	addEmployee(t, svcCtx, &types.AddRequest{Name: "王五", EmployeeNo: "00000003", OrgUuid: dept.Uuid, PersonStatus: "2"})

	cases := []struct {
		name string
		req  types.ListRequest
		want int64
	}{
		{"无筛选", types.ListRequest{}, 3},
		{"工号模糊", types.ListRequest{EmployeeNo: "0002"}, 1},
		{"姓名模糊", types.ListRequest{Name: "张"}, 1},
		{"部门", types.ListRequest{OrgUuid: dept.Uuid}, 2},
		{"人员状态离职", types.ListRequest{PersonStatus: "5"}, 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			resp := mustList(t, svcCtx, c.req)
			if resp.Total != c.want {
				t.Fatalf("total = %d, want %d", resp.Total, c.want)
			}
		})
	}

	// 分页
	paged := mustList(t, svcCtx, types.ListRequest{PageRequest: types.PageRequest{Current: 2, Size: 2}})
	if len(paged.Records) != 1 {
		t.Fatalf("page 2 size 2 should have 1 row, got %d", len(paged.Records))
	}
}

// buildImportXlsx builds the shared import/export template content.
func buildImportXlsx(t *testing.T, rows [][]string) string {
	t.Helper()
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	if err := f.SetSheetRow(sheet, "A1", &importHeader); err != nil {
		t.Fatalf("set header: %v", err)
	}
	for i, row := range rows {
		cell, err := excelize.CoordinatesToCellName(1, i+2)
		if err != nil {
			t.Fatalf("cell: %v", err)
		}
		if err := f.SetSheetRow(sheet, cell, &row); err != nil {
			t.Fatalf("set row: %v", err)
		}
	}
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		t.Fatalf("write: %v", err)
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

// TestEmployeeImportExportRoundTrip 导入→导出闭环 (spec user story 13):
// 部门/单位按名称解析, 工号空行自动生成, 导出按同一模板回读。
func TestEmployeeImportExportRoundTrip(t *testing.T) {
	svcCtx, _ := newCtx(t)
	ctx := context.Background()

	company := &hrm_org.HrmOrg{Uuid: "org-company", OrgName: "总公司", OrgType: "1", Status: "1"}
	dept := &hrm_org.HrmOrg{Uuid: "org-company-dept", ParentUuid: "org-company", OrgName: "研发部", OrgType: "2", Status: "1"}
	for _, org := range []*hrm_org.HrmOrg{company, dept} {
		if err := svcCtx.Model.HrmOrg.InsertV2(ctx, nil, org); err != nil {
			t.Fatalf("seed org: %v", err)
		}
	}

	content := buildImportXlsx(t, [][]string{
		{"", "张三", "男", "13800000000", "zhang@test.com", "研发部", "总公司", "开发", "工程师", "试用期", "2026-01-01", "", "1990-01-01", "汉族", "群众", "未婚", "湖北", "", "", "武汉市", "王母", "13800000001", "", "", "备注A"},
		{"80000001", "李四", "女", "", "", "研发部", "不存在的部门", "", "", "正式", "", "", "", "", "", "已婚", "", "", "", "", "", "", "", "", ""},
	})
	resp, err := NewImport(ctx, svcCtx, nil).Import(&types.ImportRequest{FileName: "员工.xlsx", Content: content})
	if err == nil {
		t.Fatalf("import with unknown org should fail, got %+v", resp)
	}

	// 修正部门后再导入: 两行都成功, 空工号自动生成
	content = buildImportXlsx(t, [][]string{
		{"", "张三", "男", "13800000000", "zhang@test.com", "研发部", "总公司", "开发", "工程师", "试用期", "2026-01-01", "", "1990-01-01", "汉族", "群众", "未婚", "湖北", "", "", "武汉市", "王母", "13800000001", "", "", "备注A"},
		{"80000001", "李四", "女", "", "", "研发部", "总公司", "", "", "正式", "", "", "", "", "", "已婚", "", "", "", "", "", "", "", "", ""},
	})
	resp, err = NewImport(ctx, svcCtx, nil).Import(&types.ImportRequest{FileName: "员工.xlsx", Content: content})
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if resp.SuccessCount != 2 {
		t.Fatalf("successCount = %d, want 2", resp.SuccessCount)
	}

	list := mustList(t, svcCtx, types.ListRequest{PageRequest: types.PageRequest{Current: 1, Size: 10}})
	if list.Total != 2 {
		t.Fatalf("total after import = %d, want 2", list.Total)
	}
	for _, rec := range list.Records {
		if rec.Name == "张三" {
			if rec.EmployeeNo != "00000001" || rec.OrgUuid != dept.Uuid || rec.CompanyUuid != company.Uuid ||
				rec.PersonStatus != "2" || rec.MaritalStatus != "1" {
				t.Fatalf("imported 张三 mismatch: %+v", rec)
			}
		}
		if rec.Name == "李四" && rec.EmployeeNo != "80000001" {
			t.Fatalf("imported 李四 employeeNo = %q", rec.EmployeeNo)
		}
	}

	// 导出: base64 xlsx, 数据行数 = 员工数, 部门/单位还原为中文名
	exported, err := NewExport(ctx, svcCtx, nil).Export()
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	raw, err := base64.StdEncoding.DecodeString(exported.Content)
	if err != nil {
		t.Fatalf("decode export: %v", err)
	}
	f, err := excelize.OpenReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("open export: %v", err)
	}
	defer f.Close()
	rows, err := f.GetRows(f.GetSheetName(0))
	if err != nil {
		t.Fatalf("get export rows: %v", err)
	}
	if len(rows) != 3 {
		t.Fatalf("export rows = %d, want 3 (header + 2)", len(rows))
	}
	// 工号升序 → 第一数据行 00000001 张三, 部门列 (F) = 研发部, 单位列 (G) = 总公司
	first := rows[1]
	if first[0] != "00000001" || first[1] != "张三" || first[5] != "研发部" || first[6] != "总公司" {
		t.Fatalf("export first row = %v", first)
	}
	// 人员状态还原标签
	if first[9] != "试用期" {
		t.Fatalf("export personStatus = %q, want 试用期", first[9])
	}
}
