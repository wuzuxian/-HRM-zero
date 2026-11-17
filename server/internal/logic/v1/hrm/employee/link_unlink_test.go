package employee

import (
	"context"
	"strings"
	"testing"

	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/manage_user"
	"github.com/jzero-io/jzero-admin/server/internal/model/manage_user_role"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/employee"
)

// HRM-51 关联/解绑与账号侧可见性。断言只落在对外可见行为:
// 指针指向谁、错误码、账号行与角色绑定是否还在 —— 不断言内部调用顺序。

func mustEmployeeByNo(t *testing.T, svcCtx *svc.ServiceContext, employeeNo string) types.HrmEmployee {
	t.Helper()
	list := mustList(t, svcCtx, types.ListRequest{EmployeeNo: employeeNo})
	if len(list.Records) != 1 {
		t.Fatalf("employee %s: got %d records, want 1", employeeNo, len(list.Records))
	}
	return list.Records[0]
}

// stripLink 摘除档案的账号指针 (模拟存量档案: 账号已存在但档案尚未关联)。
func stripLink(t *testing.T, svcCtx *svc.ServiceContext, employeeUuid string) {
	t.Helper()
	row, err := svcCtx.Model.HrmEmployee.FindOneByUuid(context.Background(), nil, employeeUuid)
	if err != nil {
		t.Fatalf("load employee %s: %v", employeeUuid, err)
	}
	row.UserUuid = ""
	if err = svcCtx.Model.HrmEmployee.Update(context.Background(), nil, row); err != nil {
		t.Fatalf("strip user_uuid: %v", err)
	}
}

// seedBareAccount 直种一个未关联档案的存量账号 (不挂角色), 模拟账号侧既有数据。
func seedBareAccount(t *testing.T, svcCtx *svc.ServiceContext, username, nickname string) string {
	t.Helper()
	uuid := uuidNew()
	if err := svcCtx.Model.ManageUser.InsertV2(context.Background(), nil, &manage_user.ManageUser{
		Uuid: uuid, Username: username, Password: "seed-pass", Nickname: nickname, Status: "1",
	}); err != nil {
		t.Fatalf("seed account %s: %v", username, err)
	}
	return uuid
}

func mustUserByUuid(t *testing.T, svcCtx *svc.ServiceContext, userUuid string) *manage_user.ManageUser {
	t.Helper()
	row, err := svcCtx.Model.ManageUser.FindOneByUuid(context.Background(), nil, userUuid)
	if err != nil {
		t.Fatalf("manage_user %s: %v", userUuid, err)
	}
	return row
}

func mustRolesOf(t *testing.T, svcCtx *svc.ServiceContext, userUuid string) []string {
	t.Helper()
	rows, err := svcCtx.Model.ManageUserRole.FindByCondition(context.Background(), nil, condition.Condition{
		Field: manage_user_role.UserUuid, Operator: condition.Equal, Value: userUuid,
	})
	if err != nil {
		t.Fatalf("user roles of %s: %v", userUuid, err)
	}
	roleUuids := make([]string, 0, len(rows))
	for _, row := range rows {
		roleUuids = append(roleUuids, row.RoleUuid)
	}
	return roleUuids
}

func wantCode(t *testing.T, err error, code status.Code, label string) {
	t.Helper()
	if err == nil {
		t.Fatalf("want %s error, got nil", label)
	}
	st := status.FromError(err)
	if st == nil || st.Code() != code {
		t.Fatalf("err = %v, want %s (%d)", err, label, code)
	}
}

// mustLinkableList 候选账号查询; 返回 (当前页 records, total)。
func mustLinkableList(t *testing.T, svcCtx *svc.ServiceContext, req types.LinkableUserListRequest) ([]types.LinkableUser, int64) {
	t.Helper()
	if req.Current == 0 {
		req.Current = 1
	}
	if req.Size == 0 {
		req.Size = 10
	}
	resp, err := NewLinkableUserList(context.Background(), svcCtx, nil).LinkableUserList(&req)
	if err != nil {
		t.Fatalf("linkable user list: %v", err)
	}
	return resp.Records, resp.Total
}

// TestLinkUserAndUnlink 关联成功 → 指针指向候选账号、用户状态翻转;
// 解绑 → 指针清空但账号与角色绑定原样保留。
func TestLinkUserAndUnlink(t *testing.T) {
	svcCtx, _ := newCtx(t)

	addEmployee(t, svcCtx, &types.AddRequest{Name: "张三", EmployeeNo: "00000001"})
	emp := mustEmployeeByNo(t, svcCtx, "00000001")
	stripLink(t, svcCtx, emp.Uuid)

	spareUuid := seedBareAccount(t, svcCtx, "spare-account", "备胎账号")

	// 关联成功
	if _, err := NewLinkUser(context.Background(), svcCtx, nil).LinkUser(&types.LinkUserRequest{
		EmployeeUuid: emp.Uuid, UserUuid: spareUuid,
	}); err != nil {
		t.Fatalf("link: %v", err)
	}
	after := mustEmployeeByNo(t, svcCtx, "00000001")
	if after.UserUuid != spareUuid {
		t.Fatalf("userUuid = %q, want %q (指针指向候选账号)", after.UserUuid, spareUuid)
	}
	if after.UserStatus != "1" {
		t.Fatalf("userStatus = %q, want 1", after.UserStatus)
	}
	if after.LoginName != "spare-account" {
		t.Fatalf("loginName = %q, want spare-account", after.LoginName)
	}

	// 解绑: 只清指针, 账号与角色绑定保留
	if _, err := NewUnlinkUser(context.Background(), svcCtx, nil).UnlinkUser(&types.UnlinkUserRequest{
		EmployeeUuid: emp.Uuid,
	}); err != nil {
		t.Fatalf("unlink: %v", err)
	}
	after = mustEmployeeByNo(t, svcCtx, "00000001")
	if after.UserUuid != "" || after.UserStatus != "0" || after.LoginName != "" {
		t.Fatalf("after unlink: userUuid=%q userStatus=%q loginName=%q, want 全清空", after.UserUuid, after.UserStatus, after.LoginName)
	}
	user := mustUserByUuid(t, svcCtx, spareUuid)
	if user.Username != "spare-account" || user.Status != "1" {
		t.Fatalf("account after unlink: %+v (账号应原样保留)", user)
	}
}

// TestLinkUserOccupied 账号已被其他档案持有 → 拒绝并报出占用者 (工号/姓名出现在错误消息)。
func TestLinkUserOccupied(t *testing.T) {
	svcCtx, _ := newCtx(t)

	addEmployee(t, svcCtx, &types.AddRequest{Name: "张三", EmployeeNo: "00000001"})
	addEmployee(t, svcCtx, &types.AddRequest{Name: "李四", EmployeeNo: "00000002"})
	a := mustEmployeeByNo(t, svcCtx, "00000001")
	b := mustEmployeeByNo(t, svcCtx, "00000002")
	stripLink(t, svcCtx, b.Uuid) // B 先回到未关联, 才走到占用校验 (否则先撞已关联)

	_, err := NewLinkUser(context.Background(), svcCtx, nil).LinkUser(&types.LinkUserRequest{
		EmployeeUuid: b.Uuid, UserUuid: a.UserUuid,
	})
	wantCode(t, err, errcodes.EmployeeUserOccupiedCode, "EmployeeUserOccupiedCode")
	if !strings.Contains(status.FromError(err).Message(), "00000001") || !strings.Contains(status.FromError(err).Message(), "张三") {
		t.Fatalf("occupied error should name the holder (工号+姓名), got %q", status.FromError(err).Message())
	}
	// 拒绝后指针不被写坏
	b = mustEmployeeByNo(t, svcCtx, "00000002")
	if b.UserUuid == a.UserUuid {
		t.Fatal("occupied link must not overwrite pointer")
	}
}

// TestLinkUserAlreadyLinked 员工已关联其他账号 → 再次关联被拒; 关联同一账号幂等成功。
func TestLinkUserAlreadyLinked(t *testing.T) {
	svcCtx, _ := newCtx(t)

	addEmployee(t, svcCtx, &types.AddRequest{Name: "张三", EmployeeNo: "00000001"})
	a := mustEmployeeByNo(t, svcCtx, "00000001")
	spareUuid := seedBareAccount(t, svcCtx, "spare-2", "另一个账号")

	// 已关联账号 A, 又尝试关联账号 B → 明确拒绝
	_, err := NewLinkUser(context.Background(), svcCtx, nil).LinkUser(&types.LinkUserRequest{
		EmployeeUuid: a.Uuid, UserUuid: spareUuid,
	})
	wantCode(t, err, errcodes.EmployeeUserAlreadyLinkedCode, "EmployeeUserAlreadyLinkedCode")

	// 关联到当前已持有的同一账号 → 幂等成功, 指针不变
	if _, err = NewLinkUser(context.Background(), svcCtx, nil).LinkUser(&types.LinkUserRequest{
		EmployeeUuid: a.Uuid, UserUuid: a.UserUuid,
	}); err != nil {
		t.Fatalf("relink same account should be idempotent success: %v", err)
	}
	a = mustEmployeeByNo(t, svcCtx, "00000001")
	if a.UserUuid == spareUuid {
		t.Fatal("idempotent relink must not repoint")
	}
}

// TestLinkUserAccountNotFound 关联不存在的账号 uuid → 明确的"账号不存在"错误。
func TestLinkUserAccountNotFound(t *testing.T) {
	svcCtx, _ := newCtx(t)

	addEmployee(t, svcCtx, &types.AddRequest{Name: "张三", EmployeeNo: "00000001"})
	a := mustEmployeeByNo(t, svcCtx, "00000001")
	stripLink(t, svcCtx, a.Uuid)

	_, err := NewLinkUser(context.Background(), svcCtx, nil).LinkUser(&types.LinkUserRequest{
		EmployeeUuid: a.Uuid, UserUuid: "no-such-user-uuid",
	})
	wantCode(t, err, errcodes.EmployeeUserAccountNotFoundCode, "EmployeeUserAccountNotFoundCode")

	// 员工本身不存在 → 沿用 10460
	_, err = NewLinkUser(context.Background(), svcCtx, nil).LinkUser(&types.LinkUserRequest{
		EmployeeUuid: "no-such-employee", UserUuid: "whatever",
	})
	wantCode(t, err, errcodes.EmployeeUserNotFoundCode, "EmployeeUserNotFoundCode")
}

// TestLinkableUserList 候选搜索只返回未关联账号; 关键字命中用户名或昵称; 分页正确。
func TestLinkableUserList(t *testing.T) {
	svcCtx, _ := newCtx(t)

	addEmployee(t, svcCtx, &types.AddRequest{Name: "张三", EmployeeNo: "00000001"}) // 开号 → 占用 00000001
	addEmployee(t, svcCtx, &types.AddRequest{Name: "李四", EmployeeNo: "00000002"}) // 开号 → 占用 00000002
	seedBareAccount(t, svcCtx, "spare-one", "王小明")
	seedBareAccount(t, svcCtx, "spare-two", "赵大宝")

	// 全量: 未关联账号 = 迁移种子的 admin + 两个 spare; 已被档案占用的 00000001/00000002 不出现
	records, total := mustLinkableList(t, svcCtx, types.LinkableUserListRequest{})
	if total != 3 || len(records) != 3 {
		t.Fatalf("total=%d len=%d, want 3/3 (admin + 两个未关联账号)", total, len(records))
	}
	seen := map[string]bool{}
	for _, rec := range records {
		seen[rec.Username] = true
		if rec.Uuid == "" || rec.Nickname == "" {
			t.Fatalf("candidate row missing uuid/nickname: %+v", rec)
		}
	}
	if !seen["spare-one"] || !seen["spare-two"] {
		t.Fatalf("candidates = %v, want spare-one & spare-two", seen)
	}
	if seen["00000001"] || seen["00000002"] {
		t.Fatalf("occupied accounts must not appear in candidates: %v", seen)
	}

	// 关键字按昵称命中
	records, _ = mustLinkableList(t, svcCtx, types.LinkableUserListRequest{Keyword: "大宝"})
	if len(records) != 1 || records[0].Username != "spare-two" {
		t.Fatalf("keyword by nickname = %+v, want spare-two", records)
	}
	// 关键字按用户名命中
	records, _ = mustLinkableList(t, svcCtx, types.LinkableUserListRequest{Keyword: "one"})
	if len(records) != 1 || records[0].Username != "spare-one" {
		t.Fatalf("keyword by username = %+v, want spare-one", records)
	}

	// 关联掉一个后, 候选立即缩水
	emp := mustEmployeeByNo(t, svcCtx, "00000001")
	stripLink(t, svcCtx, emp.Uuid)
	spareList, _ := mustLinkableList(t, svcCtx, types.LinkableUserListRequest{})
	var spareUuid string
	for _, rec := range spareList {
		if rec.Username == "spare-one" {
			spareUuid = rec.Uuid
		}
	}
	if _, err := NewLinkUser(context.Background(), svcCtx, nil).LinkUser(&types.LinkUserRequest{
		EmployeeUuid: emp.Uuid, UserUuid: spareUuid,
	}); err != nil {
		t.Fatalf("link: %v", err)
	}
	_, total = mustLinkableList(t, svcCtx, types.LinkableUserListRequest{})
	// 关联后候选 = admin + spare-two + 00000001 的孤儿账号 (指针被摘的账号回到候选,
	// 占用口径只看档案指针, 不看账号是否曾经关联过)
	if total != 3 {
		t.Fatalf("total after link = %d, want 3 (刚关联的账号从候选消失, 孤儿账号回到候选)", total)
	}
}

// TestCreateAndLinkUser 新建账号并关联: 自定义用户名/密码、空值默认 (用户名=工号, 密码取配置)、
// 工号撞已有账号报错、已关联拒绝、离职不开号。
func TestCreateAndLinkUser(t *testing.T) {
	svcCtx, _ := newCtx(t)

	addEmployee(t, svcCtx, &types.AddRequest{Name: "张三", EmployeeNo: "00000001"})
	emp := mustEmployeeByNo(t, svcCtx, "00000001")
	stripLink(t, svcCtx, emp.Uuid)

	// 自定义用户名+密码
	resp, err := NewCreateAndLinkUser(context.Background(), svcCtx, nil).CreateAndLinkUser(&types.CreateAndLinkUserRequest{
		EmployeeUuid: emp.Uuid, Username: "custom-user", Password: "pw-9999",
	})
	if err != nil {
		t.Fatalf("createAndLink: %v", err)
	}
	if resp.UserUuid == "" || resp.Username != "custom-user" || resp.InitialPassword != "pw-9999" {
		t.Fatalf("resp = %+v, want 自定义凭据回带", resp)
	}
	created := mustUserByUuid(t, svcCtx, resp.UserUuid)
	if created.Username != "custom-user" || created.Password != "pw-9999" {
		t.Fatalf("created account = %s/%s, want custom-user/pw-9999", created.Username, created.Password)
	}
	emp = mustEmployeeByNo(t, svcCtx, "00000001")
	if emp.UserUuid != resp.UserUuid || emp.LoginName != "custom-user" {
		t.Fatalf("pointer not written: userUuid=%q loginName=%q", emp.UserUuid, emp.LoginName)
	}

	// 已关联 → 拒绝
	_, err = NewCreateAndLinkUser(context.Background(), svcCtx, nil).CreateAndLinkUser(&types.CreateAndLinkUserRequest{
		EmployeeUuid: emp.Uuid,
	})
	wantCode(t, err, errcodes.EmployeeUserAlreadyLinkedCode, "EmployeeUserAlreadyLinkedCode")

	// 空值默认: 用户名=工号, 密码取配置 (newCtx 注入 test-pwd-123)。
	// 先把李四的开号孤儿账号从账号侧移除 (模拟账号被单独清理的存量场景),
	// 否则建档即开号已占用 工号=用户名, createAndLink 默认取工号必撞 10461。
	addEmployee(t, svcCtx, &types.AddRequest{Name: "李四", EmployeeNo: "00000002"})
	emp2 := mustEmployeeByNo(t, svcCtx, "00000002")
	stripLink(t, svcCtx, emp2.Uuid)
	if err := svcCtx.Model.ManageUser.DeleteByCondition(context.Background(), nil,
		condition.NewChain().Equal(manage_user.Uuid, emp2.UserUuid).Build()...); err != nil {
		t.Fatalf("remove orphan account: %v", err)
	}
	resp2, err := NewCreateAndLinkUser(context.Background(), svcCtx, nil).CreateAndLinkUser(&types.CreateAndLinkUserRequest{
		EmployeeUuid: emp2.Uuid,
	})
	if err != nil {
		t.Fatalf("createAndLink default: %v", err)
	}
	if resp2.Username != "00000002" || resp2.InitialPassword != "test-pwd-123" {
		t.Fatalf("default resp = %+v, want username=工号 password=配置初始密码", resp2)
	}

	// 工号撞已有账号用户名 (00000003 的账号还在, 档案指针被摘) → 10461
	addEmployee(t, svcCtx, &types.AddRequest{Name: "王五", EmployeeNo: "00000003"})
	emp3 := mustEmployeeByNo(t, svcCtx, "00000003")
	stripLink(t, svcCtx, emp3.Uuid)
	_, err = NewCreateAndLinkUser(context.Background(), svcCtx, nil).CreateAndLinkUser(&types.CreateAndLinkUserRequest{
		EmployeeUuid: emp3.Uuid,
	})
	wantCode(t, err, errcodes.EmployeeUserNoDuplicateCode, "EmployeeUserNoDuplicateCode")

	// 离职补录不开号 → 明确报错 (不是静默成功)
	addEmployee(t, svcCtx, &types.AddRequest{Name: "赵六", EmployeeNo: "80000001", PersonStatus: "5"})
	emp4 := mustEmployeeByNo(t, svcCtx, "80000001")
	_, err = NewCreateAndLinkUser(context.Background(), svcCtx, nil).CreateAndLinkUser(&types.CreateAndLinkUserRequest{
		EmployeeUuid: emp4.Uuid,
	})
	wantCode(t, err, errcodes.EmployeeUserResignedCode, "EmployeeUserResignedCode")
}

// TestUnlinkKeepsAccountAndRoles 解绑保留账号与角色绑定 (验收: 账号仍存在且仍能登录的前提)。
func TestUnlinkKeepsAccountAndRoles(t *testing.T) {
	svcCtx, _ := newCtx(t)

	// 建档即开号: 账号挂 R_EMPLOYEE (迁移 63 种子)
	addEmployee(t, svcCtx, &types.AddRequest{Name: "李四", EmployeeNo: "00000002"})
	emp := mustEmployeeByNo(t, svcCtx, "00000002")
	rolesBefore := mustRolesOf(t, svcCtx, emp.UserUuid)
	if len(rolesBefore) == 0 {
		t.Fatal("auto-provisioned account should carry default role (R_EMPLOYEE)")
	}

	if _, err := NewUnlinkUser(context.Background(), svcCtx, nil).UnlinkUser(&types.UnlinkUserRequest{
		EmployeeUuid: emp.Uuid,
	}); err != nil {
		t.Fatalf("unlink: %v", err)
	}

	user := mustUserByUuid(t, svcCtx, emp.UserUuid)
	if user.Username != "00000002" || user.Status != "1" {
		t.Fatalf("account after unlink: %s/%s (账号必须原样保留)", user.Username, user.Status)
	}
	rolesAfter := mustRolesOf(t, svcCtx, emp.UserUuid)
	if len(rolesAfter) != len(rolesBefore) {
		t.Fatalf("roles after unlink = %v, want 原样保留 %v", rolesAfter, rolesBefore)
	}
}
