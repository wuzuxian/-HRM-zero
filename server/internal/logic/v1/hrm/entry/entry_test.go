package entry

import (
	"bytes"
	"context"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/xuri/excelize/v2"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	coreSvc "github.com/jzero-io/jzero-admin/core-engine/svc"
	"github.com/jzero-io/jzero-admin/server/internal/config"
	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee_education"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee_family"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee_work"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_entry_bill"
	"github.com/jzero-io/jzero-admin/server/internal/model/manage_role"
	"github.com/jzero-io/jzero-admin/server/internal/model/manage_user"
	"github.com/jzero-io/jzero-admin/server/internal/model/manage_user_role"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	empTypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/employee"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/entry"
)

// Seams under test (per ticket acceptance criteria), all at the logic-method
// seam against the real migrations (svctest):
//  1. Add: 单据编号 HR201-... generated in-tx, 未提交 status, 申请人 snapshot,
//     预计转正缺省=入职日期+试用期月, 三子表快照落库.
//  2. 审批通过自动建档 (demo-inventory 3.1): SubmitBill → ApproveBill creates a
//     hrm_employee archive inside the approval tx — 工号自动生成 (HRM-03 seam),
//     person_status=试用期 when probation>0, formal_date=预计转正, 三子表落档案,
//     employee_uuid 回写幂等 (重复审批不重复建档).
//  3. 拒绝后不建档: RejectBill leaves employee_uuid empty and no archive row.
//  4. Lifecycle guards: edit/delete rejected outside editable statuses;
//     archived bills (employee_uuid set) locked forever.
//  5. 「我的单据」 default filter with admin bypass + 单据编号/状态/姓名 filters.
func testCtxFor(uuid, username string, roles ...string) context.Context {
	roleUuids := make([]any, len(roles))
	for i, r := range roles {
		roleUuids[i] = r
	}
	ctx := context.Background()
	ctx = context.WithValue(ctx, "uuid", uuid)
	ctx = context.WithValue(ctx, "username", username)
	ctx = context.WithValue(ctx, "role_uuids", roleUuids)
	return ctx
}

func newSvcCtx(t *testing.T, env *svctest.Env) *svc.ServiceContext {
	t.Helper()
	// HRM-50 建档即开号: 审批通过 OnApproved 在建档事务内接开号接缝, 接缝读
	// ConfigCenter (初始密码/默认角色码), 测试环境必须注入 StaticConfigCenter;
	// 审批路径用 WithSvcCtx 把本环境带入 OnApproved 闭包 (生产走全局兜底)。
	return &svc.ServiceContext{
		ServiceContext: &coreSvc.ServiceContext{SqlxConn: env.SqlConn},
		ConfigCenter:   config.NewStaticConfigCenter(config.Config{Hrm: config.HrmConf{InitialUserPassword: "test-pwd-123"}}),
		Model:          env.Model,
	}
}

func addEntry(t *testing.T, svcCtx *svc.ServiceContext, ctx context.Context, mutate func(*types.AddRequest)) (string, *types.AddRequest) {
	t.Helper()
	req := &types.AddRequest{
		Name:            "王入职",
		Phone:           "13900000001",
		Email:           "wang@example.com",
		Gender:          "1",
		Nation:          "汉族",
		MaritalStatus:   "1",
		CultureLevel:    "本科",
		Position:        "后端工程师",
		Post:            "技术专家",
		ProbationMonths: 3,
		HireDate:        "2026-12-01",
		Works: []types.HrmEntryBillWork{
			{Company: "腾讯", Position: "开发工程师", StartDate: "2020-07-01", EndDate: "2024-06-30"},
		},
		Educations: []types.HrmEntryBillEducation{
			{School: "华中科技大学", Major: "软件工程", Education: "本科", StartDate: "2012-09-01", EndDate: "2016-06-30"},
		},
		Families: []types.HrmEntryBillFamily{
			{Name: "王父", Relation: "父亲", Phone: "13800000002", WorkUnit: "退休"},
		},
	}
	if mutate != nil {
		mutate(req)
	}
	if _, err := NewAdd(ctx, svcCtx, nil).Add(req); err != nil {
		t.Fatalf("Add: %v", err)
	}
	// fetch through the public list seam
	resp, err := NewList(ctx, svcCtx, nil).List(&types.ListRequest{PageRequest: types.PageRequest{Current: 1, Size: 10}, Mine: 1})
	if err != nil {
		t.Fatalf("List after Add: %v", err)
	}
	if len(resp.Records) == 0 {
		t.Fatal("added entry not visible in 我的单据 list")
	}
	rec := resp.Records[0]
	if !strings.HasPrefix(rec.BillNo, "HR201-") {
		t.Fatalf("billNo %q does not follow HR201- format", rec.BillNo)
	}
	if rec.Status != bill.StatusUnsubmitted {
		t.Fatalf("new bill status = %q, want unsubmitted", rec.Status)
	}
	if rec.CreatorName != "张三HR" || rec.CreatorUuid == "" {
		t.Fatalf("creator snapshot wrong: uuid=%q name=%q", rec.CreatorUuid, rec.CreatorName)
	}
	return rec.Uuid, req
}

// TestAddGeneratesBillNoAndDefaults covers Add seam: HR201 编号, 未提交,
// 申请人快照, 预计转正=入职日期+试用期月 缺省计算, 三子表快照落库.
func TestAddGeneratesBillNoAndDefaults(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR", "some-role")

	billUuid, req := addEntry(t, svcCtx, ctx, nil)

	row := mustFindEntry(t, svcCtx, billUuid)
	if row.ExpectFormalDate != "2027-03-01" {
		t.Fatalf("expect_formal_date = %q, want 2027-03-01 (入职 2026-12-01 + 3 个月)", row.ExpectFormalDate)
	}
	if row.EmployeeUuid != "" {
		t.Fatalf("employee_uuid = %q before approval, want empty", row.EmployeeUuid)
	}
	if row.Name != req.Name || row.Phone != req.Phone {
		t.Fatalf("basic info snapshot lost: %+v", row)
	}

	// 三子表快照: detail seam reads them back
	detail, err := NewDetail(ctx, svcCtx, nil).Detail(&types.DetailRequest{Uuid: billUuid})
	if err != nil {
		t.Fatalf("Detail: %v", err)
	}
	if len(detail.Entry.Works) != 1 || detail.Entry.Works[0].Company != "腾讯" {
		t.Fatalf("work snapshot missing: %+v", detail.Entry.Works)
	}
	if len(detail.Entry.Educations) != 1 || detail.Entry.Educations[0].School != "华中科技大学" {
		t.Fatalf("education snapshot missing: %+v", detail.Entry.Educations)
	}
	if len(detail.Entry.Families) != 1 || detail.Entry.Families[0].Relation != "父亲" {
		t.Fatalf("family snapshot missing: %+v", detail.Entry.Families)
	}
}

func TestAddValidation(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")

	cases := []struct {
		name  string
		mutat func(*types.AddRequest)
	}{
		{"missing name", func(r *types.AddRequest) { r.Name = "" }},
		{"bad gender", func(r *types.AddRequest) { r.Gender = "3" }},
		{"bad marital status", func(r *types.AddRequest) { r.MaritalStatus = "5" }},
		{"negative probation", func(r *types.AddRequest) { r.ProbationMonths = -1 }},
		{"missing hire date", func(r *types.AddRequest) { r.HireDate = "" }},
		{"bad hire date format", func(r *types.AddRequest) { r.HireDate = "2026/12/01" }},
		{"bad expect formal format", func(r *types.AddRequest) { r.ExpectFormalDate = "2027.03.01" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := &types.AddRequest{
				Name:     "王入职",
				HireDate: "2026-12-01",
			}
			tc.mutat(req)
			if _, err := NewAdd(ctx, svcCtx, nil).Add(req); err == nil {
				t.Fatal("expected validation error, got nil")
			}
		})
	}
}

// TestApproveAutoCreatesArchive is the core acceptance: 审批通过自动建档.
func TestApproveAutoCreatesArchive(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")

	billUuid, _ := addEntry(t, svcCtx, ctx, nil)

	op := bill.Operator{UserUuid: "approver-1", Name: "李审批"}
	if err := bill.SubmitBill(context.Background(), env.SqlConn, BillTypeCode, billUuid, op, ""); err != nil {
		t.Fatalf("SubmitBill: %v", err)
	}
	if err := bill.ApproveBill(WithSvcCtx(context.Background(), svcCtx), env.SqlConn, BillTypeCode, billUuid, op, "同意入职"); err != nil {
		t.Fatalf("ApproveBill: %v", err)
	}

	row := mustFindEntry(t, svcCtx, billUuid)
	if row.Status != bill.StatusApproved {
		t.Fatalf("status = %q, want approved", row.Status)
	}
	if row.EmployeeUuid == "" {
		t.Fatal("employee_uuid not written back after approval — 自动建档 missing")
	}

	// 员工档案 assertions: 工号自动生成 / 试用期 / 预计转正 / 基本信息落档.
	emp := mustFindEmployee(t, svcCtx, row.EmployeeUuid)
	if emp.EmployeeNo == "" {
		t.Fatal("archive employee_no empty — 工号自动生成 missing")
	}
	if len(emp.EmployeeNo) != 8 || emp.EmployeeNo != "00000001" {
		t.Fatalf("archive employee_no = %q, want 00000001 (空表最大工号+1)", emp.EmployeeNo)
	}
	if emp.Name != "王入职" || emp.Phone != "13900000001" || emp.Position != "后端工程师" || emp.Post != "技术专家" {
		t.Fatalf("archive basic info not copied from 入职单: %+v", emp)
	}
	if emp.PersonStatus != employee.PersonStatusProbation {
		t.Fatalf("person_status = %q, want 试用期 (probation 3 月)", emp.PersonStatus)
	}
	if emp.HireDate != "2026-12-01" {
		t.Fatalf("hire_date = %q, want 2026-12-01", emp.HireDate)
	}
	if emp.FormalDate != "2027-03-01" {
		t.Fatalf("formal_date = %q, want 2027-03-01 (预计转正)", emp.FormalDate)
	}
	if emp.OrgUuid != "" || emp.CompanyUuid != "" {
		// 无组织时保持空; 有组织时由 OnApproved 透传
		t.Logf("org/company uuid: %q / %q", emp.OrgUuid, emp.CompanyUuid)
	}

	// 三子表落档案: count via the archive sub-tables.
	works, err := svcCtx.Model.HrmEmployeeWork.FindByCondition(context.Background(), nil,
		archiveChildrenChain(hrm_employee_work.EmployeeUuid, row.EmployeeUuid).Build()...)
	if err != nil {
		t.Fatalf("find archive works: %v", err)
	}
	if len(works) != 1 || works[0].Company != "腾讯" {
		t.Fatalf("archive works wrong: %+v", works)
	}
	educations, err := svcCtx.Model.HrmEmployeeEducation.FindByCondition(context.Background(), nil,
		archiveChildrenChain(hrm_employee_education.EmployeeUuid, row.EmployeeUuid).Build()...)
	if err != nil {
		t.Fatalf("find archive educations: %v", err)
	}
	if len(educations) != 1 || educations[0].School != "华中科技大学" {
		t.Fatalf("archive educations wrong: %+v", educations)
	}
	families, err := svcCtx.Model.HrmEmployeeFamily.FindByCondition(context.Background(), nil,
		archiveChildrenChain(hrm_employee_family.EmployeeUuid, row.EmployeeUuid).Build()...)
	if err != nil {
		t.Fatalf("find archive families: %v", err)
	}
	if len(families) != 1 || families[0].Name != "王父" {
		t.Fatalf("archive families wrong: %+v", families)
	}
}

// TestApproveArchiveIdempotent: 重复审批/回写重放不得重复建档 — employee_uuid
// 回写后再次 onApproved 必须是 no-op, 档案表不新增行, 工号不递增.
func TestApproveArchiveIdempotent(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")

	billUuid, _ := addEntry(t, svcCtx, ctx, nil)

	op := bill.Operator{UserUuid: "approver-1", Name: "李审批"}
	if err := bill.SubmitBill(context.Background(), env.SqlConn, BillTypeCode, billUuid, op, ""); err != nil {
		t.Fatalf("SubmitBill: %v", err)
	}
	if err := bill.ApproveBill(WithSvcCtx(context.Background(), svcCtx), env.SqlConn, BillTypeCode, billUuid, op, ""); err != nil {
		t.Fatalf("ApproveBill: %v", err)
	}
	row := mustFindEntry(t, svcCtx, billUuid)
	firstUuid := row.EmployeeUuid
	if firstUuid == "" {
		t.Fatal("first approval did not create archive")
	}

	// Replay the write-back directly (重复审批兜底) — must be a no-op.
	if err := onApproved(WithSvcCtx(context.Background(), svcCtx), env.SqlConn, billUuid); err != nil {
		t.Fatalf("onApproved replay: %v", err)
	}
	if err := onApproved(WithSvcCtx(context.Background(), svcCtx), env.SqlConn, billUuid); err != nil {
		t.Fatalf("onApproved replay 2: %v", err)
	}

	row = mustFindEntry(t, svcCtx, billUuid)
	if row.EmployeeUuid != firstUuid {
		t.Fatalf("employee_uuid changed on replay: %q → %q (重复建档!)", firstUuid, row.EmployeeUuid)
	}

	// exactly one archive row for this 入职单 name
	employees, err := svcCtx.Model.HrmEmployee.FindByCondition(context.Background(), nil)
	if err != nil {
		t.Fatalf("find employees: %v", err)
	}
	count := 0
	for _, e := range employees {
		if e.Name == "王入职" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("archive rows for 王入职 = %d, want 1 (重复建档)", count)
	}
}

// TestRejectDoesNotArchive: 审批驳回不建档 — employee_uuid stays empty, no
// hrm_employee row; resubmit → approve still archives.
func TestRejectDoesNotArchive(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")

	billUuid, _ := addEntry(t, svcCtx, ctx, nil)

	op := bill.Operator{UserUuid: "approver-1", Name: "李审批"}
	if err := bill.SubmitBill(context.Background(), env.SqlConn, BillTypeCode, billUuid, op, ""); err != nil {
		t.Fatalf("SubmitBill: %v", err)
	}
	if err := bill.RejectBill(context.Background(), env.SqlConn, BillTypeCode, billUuid, op, "材料不全"); err != nil {
		t.Fatalf("RejectBill: %v", err)
	}

	row := mustFindEntry(t, svcCtx, billUuid)
	if row.Status != bill.StatusRejected {
		t.Fatalf("status = %q, want rejected", row.Status)
	}
	if row.EmployeeUuid != "" {
		t.Fatalf("employee_uuid = %q after rejection — 驳回后不应建档", row.EmployeeUuid)
	}
	employees, err := svcCtx.Model.HrmEmployee.FindByCondition(context.Background(), nil)
	if err != nil {
		t.Fatalf("find employees: %v", err)
	}
	if len(employees) != 0 {
		t.Fatalf("hrm_employee rows = %d after rejection, want 0 (拒绝后不建档)", len(employees))
	}

	// 审批不通过 → 重新提交 → 通过 → 建档成功 (含第二个入职单工号递增 00000002)
	billUuid2, _ := addEntry(t, svcCtx, ctx, func(r *types.AddRequest) {
		r.Name = "李入职"
		r.ProbationMonths = 0
	})
	if err := bill.SubmitBill(context.Background(), env.SqlConn, BillTypeCode, billUuid, op, ""); err != nil {
		t.Fatalf("SubmitBill (resubmit): %v", err)
	}
	if err := bill.ApproveBill(WithSvcCtx(context.Background(), svcCtx), env.SqlConn, BillTypeCode, billUuid, op, ""); err != nil {
		t.Fatalf("ApproveBill (resubmit): %v", err)
	}
	row = mustFindEntry(t, svcCtx, billUuid)
	if row.EmployeeUuid == "" {
		t.Fatal("archive not created after resubmit+approve")
	}
	emp := mustFindEmployee(t, svcCtx, row.EmployeeUuid)
	if emp.EmployeeNo == "" {
		t.Fatal("employee_no empty")
	}

	// 无试用期的第二单: person_status = 正式
	_ = billUuid2
	if err := bill.SubmitBill(context.Background(), env.SqlConn, BillTypeCode, billUuid2, op, ""); err != nil {
		t.Fatalf("SubmitBill 2: %v", err)
	}
	if err := bill.ApproveBill(WithSvcCtx(context.Background(), svcCtx), env.SqlConn, BillTypeCode, billUuid2, op, ""); err != nil {
		t.Fatalf("ApproveBill 2: %v", err)
	}
	row2 := mustFindEntry(t, svcCtx, billUuid2)
	emp2 := mustFindEmployee(t, svcCtx, row2.EmployeeUuid)
	if emp2.PersonStatus != employee.PersonStatusRegular {
		t.Fatalf("second archive person_status = %q, want 正式 (probation 0)", emp2.PersonStatus)
	}
}

// TestEditDeleteGuardedByStatus covers lifecycle guards incl. archived lock.
func TestEditDeleteGuardedByStatus(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")
	op := bill.Operator{UserUuid: "approver-1", Name: "李审批"}

	billUuid, _ := addEntry(t, svcCtx, ctx, nil)

	// 审批中 (approving) is locked: edit and delete must be rejected.
	if err := bill.SubmitBill(context.Background(), env.SqlConn, BillTypeCode, billUuid, op, ""); err != nil {
		t.Fatalf("SubmitBill: %v", err)
	}
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid:     billUuid,
		Name:     "王改",
		HireDate: "2026-12-01",
	}); err == nil {
		t.Fatal("edit during 审批中 should fail")
	}
	if _, err := NewDelete(ctx, svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{billUuid}}); err == nil {
		t.Fatal("delete during 审批中 should fail")
	}

	// 审批通过 (approved) is locked too — 已自动建档.
	if err := bill.ApproveBill(WithSvcCtx(context.Background(), svcCtx), env.SqlConn, BillTypeCode, billUuid, op, ""); err != nil {
		t.Fatalf("ApproveBill: %v", err)
	}
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid:     billUuid,
		Name:     "王改",
		HireDate: "2026-12-01",
	}); err == nil {
		t.Fatal("edit after 审批通过 (已建档) should fail")
	}
	if _, err := NewDelete(ctx, svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{billUuid}}); err == nil {
		t.Fatal("delete after 审批通过 (已建档) should fail")
	}

	// 已撤销 (revoked) is editable/deletable again.
	billUuid2, _ := addEntry(t, svcCtx, ctx, func(r *types.AddRequest) { r.Name = "赵撤销" })
	if err := bill.SubmitBill(context.Background(), env.SqlConn, BillTypeCode, billUuid2, op, ""); err != nil {
		t.Fatalf("SubmitBill 2: %v", err)
	}
	if err := bill.RevokeBill(context.Background(), env.SqlConn, BillTypeCode, billUuid2, op, "撤回"); err != nil {
		t.Fatalf("RevokeBill: %v", err)
	}
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid:     billUuid2,
		Name:     "赵撤销改",
		HireDate: "2026-12-02",
	}); err != nil {
		t.Fatalf("edit after revoke: %v", err)
	}
	if _, err := NewDelete(ctx, svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{billUuid2}}); err != nil {
		t.Fatalf("delete after revoke: %v", err)
	}
}

// TestListMineFilterAndAdminBypass + 单据编号/单据状态/姓名 filters (demo-inventory 3.1).
func TestListMineFilterAndAdminBypass(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)

	userCtx := testCtxFor("user-1", "张三HR")
	addEntry(t, svcCtx, userCtx, nil)
	addEntry(t, svcCtx, userCtx, func(r *types.AddRequest) { r.Name = "李入职" })
	adminCtx := testCtxFor("admin-1", "管理员", superAdminRoleUuid())

	mineResp, err := NewList(userCtx, svcCtx, nil).List(&types.ListRequest{PageRequest: types.PageRequest{Current: 1, Size: 10}})
	if err != nil {
		t.Fatalf("List(mine): %v", err)
	}
	if mineResp.Total != 2 {
		t.Fatalf("non-admin sees %d bills, want own 2", mineResp.Total)
	}

	allResp, err := NewList(adminCtx, svcCtx, nil).List(&types.ListRequest{PageRequest: types.PageRequest{Current: 1, Size: 10}})
	if err != nil {
		t.Fatalf("List(admin): %v", err)
	}
	if allResp.Total != 2 {
		t.Fatalf("admin should see all bills: total=%d", allResp.Total)
	}

	forcedMine, err := NewList(adminCtx, svcCtx, nil).List(&types.ListRequest{PageRequest: types.PageRequest{Current: 1, Size: 10}, Mine: 1})
	if err != nil {
		t.Fatalf("List(admin, mine=1): %v", err)
	}
	if forcedMine.Total != 0 {
		t.Fatalf("admin with mine=1 should see none of user-1's bills: total=%d", forcedMine.Total)
	}

	// 单据编号 filter (取自 add 后的编号前缀 — 用 HR201- 模糊匹配两单)
	noFilter, err := NewList(adminCtx, svcCtx, nil).List(&types.ListRequest{
		PageRequest: types.PageRequest{Current: 1, Size: 10},
		BillNo:      "HR201-",
	})
	if err != nil {
		t.Fatalf("List(billNo filter): %v", err)
	}
	if noFilter.Total != 2 {
		t.Fatalf("billNo filter: total=%d, want 2", noFilter.Total)
	}

	// 单据状态 filter
	stFilter, err := NewList(adminCtx, svcCtx, nil).List(&types.ListRequest{
		PageRequest: types.PageRequest{Current: 1, Size: 10},
		Status:      bill.StatusUnsubmitted,
	})
	if err != nil {
		t.Fatalf("List(status filter): %v", err)
	}
	if stFilter.Total != 2 {
		t.Fatalf("status filter: total=%d, want 2", stFilter.Total)
	}

	// 姓名 filter
	nameFilter, err := NewList(adminCtx, svcCtx, nil).List(&types.ListRequest{
		PageRequest: types.PageRequest{Current: 1, Size: 10},
		Name:        "李入职",
	})
	if err != nil {
		t.Fatalf("List(name filter): %v", err)
	}
	if nameFilter.Total != 1 || nameFilter.Records[0].Name != "李入职" {
		t.Fatalf("name filter: total=%d, want 1 李入职", nameFilter.Total)
	}
}

func mustFindEntry(t *testing.T, svcCtx *svc.ServiceContext, billUuid string) *hrm_entry_bill.HrmEntryBill {
	t.Helper()
	row, err := svcCtx.Model.HrmEntryBill.FindOneByUuid(context.Background(), nil, billUuid)
	if err != nil {
		t.Fatalf("find entry %s: %v", billUuid, err)
	}
	return row
}

func mustFindEmployee(t *testing.T, svcCtx *svc.ServiceContext, uuid string) *hrm_employee.HrmEmployee {
	t.Helper()
	row, err := svcCtx.Model.HrmEmployee.FindOneByUuid(context.Background(), nil, uuid)
	if err != nil {
		t.Fatalf("find employee %s: %v", uuid, err)
	}
	return row
}

func archiveChildrenChain(field conditionField, uuid string) conditionChain {
	return condition.NewChain().Equal(field, uuid)
}

// tiny aliases keep the archive-children assertions readable without importing
// condition twice.
type conditionField = condition.Field
type conditionChain = condition.Chain

func superAdminRoleUuid() string {
	// keep the seeded R_SUPER uuid local to tests; production logic uses hrmuser
	return "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
}

var _ = sqlx.ErrNotFound

// ---- HRM-50 建档即开号: 入职单审批入口 (ADR 0007) ----

// submitAndApproveEntry 提交并审批通过一张入职单 (审批事务内自动建档+开号)。
func submitAndApproveEntry(t *testing.T, svcCtx *svc.ServiceContext, billUuid string) {
	t.Helper()
	op := bill.Operator{UserUuid: "approver-1", Name: "李审批"}
	if err := bill.SubmitBill(WithSvcCtx(context.Background(), svcCtx), envSqlConnOf(svcCtx), BillTypeCode, billUuid, op, ""); err != nil {
		t.Fatalf("SubmitBill: %v", err)
	}
	if err := bill.ApproveBill(WithSvcCtx(context.Background(), svcCtx), envSqlConnOf(svcCtx), BillTypeCode, billUuid, op, "同意入职"); err != nil {
		t.Fatalf("ApproveBill: %v", err)
	}
}

// envSqlConnOf 取测试环境的连接 (WithSvcCtx 注入的 svcCtx 与 ApproveBill 必须同一 DB)。
func envSqlConnOf(svcCtx *svc.ServiceContext) sqlx.SqlConn {
	return svcCtx.SqlxConn
}

// mustFindRoleByCodeEntry 按角色码定位 manage_role 行 (entry 包测试本地助手)。
func mustFindRoleByCodeEntry(t *testing.T, svcCtx *svc.ServiceContext, code string) *manage_role.ManageRole {
	t.Helper()
	row, err := svcCtx.Model.ManageRole.FindOneByCondition(context.Background(), nil,
		condition.NewChain().Equal(manage_role.Code, code).Build()...)
	if err != nil {
		t.Fatalf("find manage_role by code %q: %v", code, err)
	}
	return row
}

// accountShape 账号外部可见快照: password/nickname/phone/email/status/挂角色 uuid 集合。
type accountShape struct {
	password  string
	nickname  string
	phone     string
	email     string
	status    string
	roleUuids map[string]bool
}

func accountShapeOf(t *testing.T, svcCtx *svc.ServiceContext, userUuid string) accountShape {
	t.Helper()
	user, err := svcCtx.Model.ManageUser.FindOneByUuid(context.Background(), nil, userUuid)
	if err != nil {
		t.Fatalf("find manage_user %s: %v", userUuid, err)
	}
	rows, err := svcCtx.Model.ManageUserRole.FindByCondition(context.Background(), nil,
		condition.NewChain().Equal(manage_user_role.UserUuid, userUuid).Build()...)
	if err != nil {
		t.Fatalf("find user roles: %v", err)
	}
	roles := make(map[string]bool, len(rows))
	for _, r := range rows {
		roles[r.RoleUuid] = true
	}
	return accountShape{
		password:  user.Password,
		nickname:  user.Nickname,
		phone:     user.Phone,
		email:     user.Email,
		status:    user.Status,
		roleUuids: roles,
	}
}

// TestThreeEntryProvisionsEquivalentAccounts 三入口账号等价 (ticket 验收): 同一份
// 档案数据经 手工新增 / 批量导入 / 入职单审批自动建档 三个入口产出账号, 其
// 用户名(=工号)、初始密码(取配置)、昵称/手机/邮箱(取档案)、状态(启用)、
// 默认角色(R_EMPLOYEE) 完全一致。
func TestThreeEntryProvisionsEquivalentAccounts(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)

	// 入口一: 手工新增 (在职, 与入职单同基本资料)
	addResp, err := employee.NewAdd(context.Background(), svcCtx, nil).Add(&empTypes.AddRequest{
		Name: "王入职", Gender: "1", Phone: "13900000001", Email: "wang@example.com",
		PersonStatus: employee.PersonStatusRegular,
	})
	if err != nil {
		t.Fatalf("add entry: %v", err)
	}
	if !addResp.UserCreated {
		t.Fatalf("add resp = %+v, want userCreated=true", addResp)
	}

	// 入口二: 批量导入 (同一模板, 空工号自动生成)
	content := buildEmployeeImportXlsx(t, [][]string{
		{"", "王入职", "男", "13900000001", "wang@example.com", "", "", "", "", "正式", "", "", "", "", "", "", "", "", "", "", "", "", "", "", ""},
	})
	importResp, err := employee.NewImport(context.Background(), svcCtx, nil).Import(&empTypes.ImportRequest{
		FileName: "员工.xlsx", Content: content,
	})
	if err != nil {
		t.Fatalf("import entry: %v", err)
	}
	if importResp.SuccessCount != 1 {
		t.Fatalf("import successCount = %d, want 1", importResp.SuccessCount)
	}

	// 入口三: 入职单审批通过自动建档
	billUuid, _ := addEntry(t, svcCtx, testCtxFor("user-1", "张三HR"), nil)
	submitAndApproveEntry(t, svcCtx, billUuid)
	entryRow := mustFindEntry(t, svcCtx, billUuid)
	if entryRow.EmployeeUuid == "" {
		t.Fatal("approval did not create archive")
	}

	// 三个档案的账号逐项等价
	addEmp := mustFindEmployeeByName(t, svcCtx, "王入职", addResp.Username)
	archEmp := mustFindEmployee(t, svcCtx, entryRow.EmployeeUuid)

	userByUuid := func(t *testing.T, uuid string) *manage_user.ManageUser {
		t.Helper()
		row, err := svcCtx.Model.ManageUser.FindOneByUuid(context.Background(), nil, uuid)
		if err != nil {
			t.Fatalf("find manage_user %s: %v", uuid, err)
		}
		return row
	}

	// 新增入口的档案: 工号即账号用户名
	addUser := userByUuid(t, addEmp.UserUuid)
	shape1 := accountShapeOf(t, svcCtx, addEmp.UserUuid)

	// 入职单入口档案
	archUser := userByUuid(t, archEmp.UserUuid)
	shape3 := accountShapeOf(t, svcCtx, archEmp.UserUuid)

	// 用户名 = 各自档案工号
	if addUser.Username != addEmp.EmployeeNo {
		t.Fatalf("add entry username = %q, want 工号 %q", addUser.Username, addEmp.EmployeeNo)
	}
	if archUser.Username != archEmp.EmployeeNo {
		t.Fatalf("entry-bill username = %q, want 工号 %q", archUser.Username, archEmp.EmployeeNo)
	}

	role := mustFindRoleByCodeEntry(t, svcCtx, "R_EMPLOYEE")

	// 逐字段等价: 密码/昵称/手机/邮箱/状态/角色
	if shape1.password != "test-pwd-123" || shape3.password != "test-pwd-123" {
		t.Fatalf("password mismatch: add=%q arch=%q, want test-pwd-123", shape1.password, shape3.password)
	}
	if shape1.nickname != "王入职" || shape3.nickname != "王入职" {
		t.Fatalf("nickname mismatch: %q / %q, want 王入职 (取档案姓名)", shape1.nickname, shape3.nickname)
	}
	if shape1.phone != "13900000001" || shape3.phone != "13900000001" {
		t.Fatalf("phone mismatch: %q / %q", shape1.phone, shape3.phone)
	}
	if shape1.email != "wang@example.com" || shape3.email != "wang@example.com" {
		t.Fatalf("email mismatch: %q / %q", shape1.email, shape3.email)
	}
	if shape1.status != "1" || shape3.status != "1" {
		t.Fatalf("status mismatch: %q / %q, want 1 (启用)", shape1.status, shape3.status)
	}
	if len(shape1.roleUuids) != 1 || !shape1.roleUuids[role.Uuid] {
		t.Fatalf("add entry roles = %+v, want exactly R_EMPLOYEE", shape1.roleUuids)
	}
	if len(shape3.roleUuids) != 1 || !shape3.roleUuids[role.Uuid] {
		t.Fatalf("entry-bill roles = %+v, want exactly R_EMPLOYEE", shape3.roleUuids)
	}

	// 导入入口的账号同样等价 (工号 = 除新增/审批两个工号外的第三条记录)
	addRespNo := addResp.Username
	archNo := archEmp.EmployeeNo
	importEmp := findEmployeeByExcludedNos(t, svcCtx, []string{addRespNo, archNo})
	if importEmp == nil || importEmp.UserUuid == "" {
		t.Fatal("import row did not provision account")
	}
	importUser := userByUuid(t, importEmp.UserUuid)
	if importUser.Username != importEmp.EmployeeNo {
		t.Fatalf("import username = %q, want 工号 %q", importUser.Username, importEmp.EmployeeNo)
	}
	shape2 := accountShapeOf(t, svcCtx, importEmp.UserUuid)
	if shape2.password != "test-pwd-123" || shape2.nickname != "王入职" ||
		shape2.phone != "13900000001" || shape2.email != "wang@example.com" || shape2.status != "1" {
		t.Fatalf("import account shape mismatch: %+v", shape2)
	}
	if len(shape2.roleUuids) != 1 || !shape2.roleUuids[role.Uuid] {
		t.Fatalf("import roles = %+v, want exactly R_EMPLOYEE", shape2.roleUuids)
	}
}

// buildEmployeeImportXlsx entry 包本地的员工导入模板 (与 employee.importHeader 同一列序)。
func buildEmployeeImportXlsx(t *testing.T, rows [][]string) string {
	t.Helper()
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	header := []string{
		"工号", "姓名", "性别", "手机号", "邮箱",
		"所属部门", "所属单位", "职位", "职务",
		"人员状态", "入职日期", "转正日期",
		"生日", "民族", "政治面貌", "婚姻状况", "籍贯", "身份证号",
		"户口地址", "现住址", "紧急联系人", "紧急联系人电话",
		"工资开户行", "工资卡账户", "备注",
	}
	if err := f.SetSheetRow(sheet, "A1", &header); err != nil {
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
		t.Fatalf("write xlsx: %v", err)
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

// mustFindEmployeeByName 按姓名定位档案 (可选工号精确过滤; 测试数据受控)。
func mustFindEmployeeByName(t *testing.T, svcCtx *svc.ServiceContext, name, employeeNo string) *hrm_employee.HrmEmployee {
	t.Helper()
	chain := condition.NewChain().Equal(hrm_employee.Name, name)
	if employeeNo != "" {
		chain = chain.Equal(hrm_employee.EmployeeNo, employeeNo)
	}
	rows, err := svcCtx.Model.HrmEmployee.FindByCondition(context.Background(), nil, chain.Build()...)
	if err != nil || len(rows) == 0 {
		t.Fatalf("find employee by name %s (no=%s): %v, rows=%d", name, employeeNo, err, len(rows))
	}
	return rows[0]
}

// findEmployeeByExcludedNos 取工号不在排除列表内的唯一档案 (三入口同姓名场景定位导入行)。
func findEmployeeByExcludedNos(t *testing.T, svcCtx *svc.ServiceContext, excluded []string) *hrm_employee.HrmEmployee {
	t.Helper()
	rows, err := svcCtx.Model.HrmEmployee.FindByCondition(context.Background(), nil)
	if err != nil {
		t.Fatalf("find employees: %v", err)
	}
	excl := map[string]bool{}
	for _, no := range excluded {
		excl[no] = true
	}
	for _, row := range rows {
		if !excl[row.EmployeeNo] {
			return row
		}
	}
	t.Fatalf("no employee outside excluded nos %v", excluded)
	return nil
}

// TestApproveProvisionFailureRollsBackApproval 开号失败则审批整体回滚 (ticket 验收):
// 自动生成的工号撞已有账号用户名 → ApproveBill 报 10461, 单据状态留在 审批中,
// 档案不落库, employee_uuid 不回写。
func TestApproveProvisionFailureRollsBackApproval(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)

	// 空表下审批建档的自动工号是 00000001; 预先种一个同名账号制造开号冲突
	if err := svcCtx.Model.ManageUser.InsertV2(context.Background(), nil, &manage_user.ManageUser{
		Uuid: uuid.New().String(), Username: "00000001", Password: "x", Nickname: "冲突账号", Status: "1",
	}); err != nil {
		t.Fatalf("seed conflicting account: %v", err)
	}

	billUuid, _ := addEntry(t, svcCtx, testCtxFor("user-1", "张三HR"), nil)
	op := bill.Operator{UserUuid: "approver-1", Name: "李审批"}
	if err := bill.SubmitBill(WithSvcCtx(context.Background(), svcCtx), envSqlConnOf(svcCtx), BillTypeCode, billUuid, op, ""); err != nil {
		t.Fatalf("SubmitBill: %v", err)
	}
	err := bill.ApproveBill(WithSvcCtx(context.Background(), svcCtx), envSqlConnOf(svcCtx), BillTypeCode, billUuid, op, "同意入职")
	if st := status.FromError(err); st == nil || st.Code() != errcodes.EmployeeUserNoDuplicateCode {
		t.Fatalf("approve err = %v, want EmployeeUserNoDuplicateCode (10461)", err)
	}

	// 审批整体回滚: 状态留在审批中, 员工档案未生成
	row := mustFindEntry(t, svcCtx, billUuid)
	if row.Status != bill.StatusApproving {
		t.Fatalf("status after failed approval = %q, want approving (回滚)", row.Status)
	}
	if row.EmployeeUuid != "" {
		t.Fatalf("employee_uuid = %q, want empty (档案未落库)", row.EmployeeUuid)
	}
	employees, err := svcCtx.Model.HrmEmployee.FindByCondition(context.Background(), nil)
	if err != nil {
		t.Fatalf("find employees: %v", err)
	}
	if len(employees) != 0 {
		t.Fatalf("archives after failed approval = %d, want 0 (同事务回滚)", len(employees))
	}
}
