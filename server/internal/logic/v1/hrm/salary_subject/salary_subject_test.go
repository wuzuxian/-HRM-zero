package salary_subject

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_org"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_subject"
)

func newCtx(t *testing.T) *svc.ServiceContext {
	t.Helper()
	env := svctest.New(t)
	return &svc.ServiceContext{Model: env.Model}
}

// seedCompany / seedDepartment 直接插组织行（org_type 1=公司 2=部门），
// 供主体关联公司校验使用（Add/Edit 只接受公司节点）.
func seedCompany(t *testing.T, svcCtx *svc.ServiceContext, name string) string {
	t.Helper()
	u := uuid.New().String()
	if _, err := svcCtx.Model.HrmOrg.Insert(t.Context(), nil, &hrm_org.HrmOrg{
		Uuid: u, ParentUuid: "", OrgName: name, OrgType: "1", OrderNum: 1, Status: "1",
	}); err != nil {
		t.Fatalf("seed company %s: %v", name, err)
	}
	return u
}

func seedDepartment(t *testing.T, svcCtx *svc.ServiceContext, name string) string {
	t.Helper()
	u := uuid.New().String()
	if _, err := svcCtx.Model.HrmOrg.Insert(t.Context(), nil, &hrm_org.HrmOrg{
		Uuid: u, ParentUuid: "", OrgName: name, OrgType: "2", OrderNum: 2, Status: "1",
	}); err != nil {
		t.Fatalf("seed department %s: %v", name, err)
	}
	return u
}

func mustList(t *testing.T, svcCtx *svc.ServiceContext, req types.ListRequest) []types.HrmSalarySubject {
	t.Helper()
	req.Current = 1
	req.Size = 100
	resp, err := NewList(context.Background(), svcCtx, nil).List(&req)
	if err != nil {
		t.Fatalf("list salary subjects: %v", err)
	}
	return resp.Records
}

func findByName(records []types.HrmSalarySubject, name string) *types.HrmSalarySubject {
	for i := range records {
		if records[i].SubjectName == name {
			return &records[i]
		}
	}
	return nil
}

// TestSalarySubjectCrud walks the CRUD seam (HRM-33 user story 54):
// 新增（关联公司名称快照）→ 编辑（换公司/启停）→ 删除, plus 名称筛选.
func TestSalarySubjectCrud(t *testing.T) {
	svcCtx := newCtx(t)
	ctx := context.Background()

	companyA := seedCompany(t, svcCtx, "甲公司")
	companyB := seedCompany(t, svcCtx, "乙公司")

	// 新增: 关联公司快照
	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		SubjectName: "甲公司薪酬主体", OrgUuid: companyA,
		PayAccount: "6222-0000-1111", PayBank: "工商银行",
	}); err != nil {
		t.Fatalf("add salary subject: %v", err)
	}
	records := mustList(t, svcCtx, types.ListRequest{})
	if len(records) != 1 {
		t.Fatalf("expected 1 subject after add, got %d", len(records))
	}
	if records[0].OrgName != "甲公司" || records[0].Status != "1" || records[0].PayBank != "工商银行" {
		t.Fatalf("add snapshot/defaults not applied: %+v", records[0])
	}

	// 名称筛选
	if named := mustList(t, svcCtx, types.ListRequest{SubjectName: "甲公司薪酬主体"}); len(named) != 1 {
		t.Fatalf("name filter failed: %+v", named)
	}
	if named := mustList(t, svcCtx, types.ListRequest{SubjectName: "不存在"}); len(named) != 0 {
		t.Fatalf("name filter should be empty: %+v", named)
	}

	// 编辑: 换关联公司 + 停用
	row := findByName(mustList(t, svcCtx, types.ListRequest{}), "甲公司薪酬主体")
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid: row.Uuid, SubjectName: "甲公司薪酬主体", OrgUuid: companyB,
		PayAccount: "6222-0000-2222", PayBank: "建设银行", Status: "0",
	}); err != nil {
		t.Fatalf("edit salary subject: %v", err)
	}
	row = findByName(mustList(t, svcCtx, types.ListRequest{}), "甲公司薪酬主体")
	if row.OrgName != "乙公司" || row.Status != "0" || row.PayAccount != "6222-0000-2222" {
		t.Fatalf("edit not applied: %+v", row)
	}

	// 编辑不存在的主体
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid: "no-such-uuid", SubjectName: "幽灵主体", OrgUuid: companyA, Status: "1",
	}); err == nil {
		t.Fatalf("editing unknown subject should fail")
	}

	// 删除
	if _, err := NewDelete(ctx, svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{row.Uuid}}); err != nil {
		t.Fatalf("delete salary subject: %v", err)
	}
	if left := mustList(t, svcCtx, types.ListRequest{}); len(left) != 0 {
		t.Fatalf("expected empty after delete, got %+v", left)
	}
}

// TestSalarySubjectValidation 非法参数守卫: 名称空/启停非法/关联公司不存在/关联非公司节点.
func TestSalarySubjectValidation(t *testing.T) {
	svcCtx := newCtx(t)
	ctx := context.Background()

	company := seedCompany(t, svcCtx, "丙公司")
	department := seedDepartment(t, svcCtx, "研发部")

	cases := []struct {
		name string
		req  types.AddRequest
	}{
		{"empty name", types.AddRequest{SubjectName: "", OrgUuid: company}},
		{"bad status", types.AddRequest{SubjectName: "X主体", OrgUuid: company, Status: "2"}},
		{"missing org", types.AddRequest{SubjectName: "X主体", OrgUuid: ""}},
		{"unknown org", types.AddRequest{SubjectName: "X主体", OrgUuid: "no-such-org"}},
		{"org not company", types.AddRequest{SubjectName: "X主体", OrgUuid: department}},
	}
	for _, tc := range cases {
		if _, err := NewAdd(ctx, svcCtx, nil).Add(&tc.req); err == nil {
			t.Fatalf("%s: expected error", tc.name)
		}
	}

	// 合法路径兜底确认（确保上面的失败不是环境性的）
	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{SubjectName: "丙公司主体", OrgUuid: company}); err != nil {
		t.Fatalf("valid add should pass: %v", err)
	}
}
