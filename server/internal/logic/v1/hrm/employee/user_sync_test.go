package employee

import (
	"context"
	"testing"

	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	coreSvc "github.com/jzero-io/jzero-admin/core-engine/svc"
	"github.com/jzero-io/jzero-admin/server/internal/config"
	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/manage_user"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/employee"
)

// newSyncCtx 员工↔账号同步测试环境: 需要 SqlxConn (事务) + ConfigCenter (初始密码配置).
func newSyncCtx(t *testing.T) (*svc.ServiceContext, *svctest.Env) {
	t.Helper()
	env := svctest.New(t)
	return &svc.ServiceContext{
		ServiceContext: &coreSvc.ServiceContext{SqlxConn: env.SqlConn},
		ConfigCenter:   config.NewStaticConfigCenter(config.Config{Hrm: config.HrmConf{InitialUserPassword: "test-pwd-123"}}),
		Model:          env.Model,
	}, env
}

func mustFindUserByUsername(t *testing.T, svcCtx *svc.ServiceContext, username string) *manage_user.ManageUser {
	t.Helper()
	row, err := svcCtx.Model.ManageUser.FindOneByCondition(context.Background(), nil,
		condition.NewChain().Equal(manage_user.Username, username).Build()...)
	if err != nil {
		t.Fatalf("find manage_user by username %q: %v", username, err)
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

// TestGenerateUserSingle 批量端点传一个 uuid = 单个生成:
// manage_user 行 (username=工号, nickname=姓名, 初始密码取配置) + hrm_employee.user_uuid 回写.
func TestGenerateUserSingle(t *testing.T) {
	svcCtx, _ := newSyncCtx(t)

	addEmployee(t, svcCtx, &types.AddRequest{Name: "张三", EmployeeNo: "10000001", PersonStatus: "1"})
	list := mustList(t, svcCtx, types.ListRequest{})
	if len(list.Records) != 1 {
		t.Fatalf("records = %d, want 1", len(list.Records))
	}
	empUuid := list.Records[0].Uuid

	resp, err := NewGenerateUser(context.Background(), svcCtx, nil).GenerateUser(&types.GenerateUserRequest{Uuids: []string{empUuid}})
	if err != nil {
		t.Fatalf("generateUser: %v", err)
	}
	if resp.Generated != 1 || resp.Skipped != 0 {
		t.Fatalf("generated=%d skipped=%d, want 1/0", resp.Generated, resp.Skipped)
	}

	emp := mustFindEmployee(t, svcCtx, empUuid)
	if emp.UserUuid == "" {
		t.Fatal("user_uuid not written back to hrm_employee")
	}

	user := mustFindUserByUsername(t, svcCtx, "10000001")
	if user.Nickname != "张三" {
		t.Fatalf("nickname = %q, want 张三", user.Nickname)
	}
	if user.Password != "test-pwd-123" {
		t.Fatalf("password = %q, want test-pwd-123 (初始密码取配置)", user.Password)
	}
	if user.Status != "1" {
		t.Fatalf("status = %q, want 1 (启用)", user.Status)
	}
	if user.Uuid != emp.UserUuid {
		t.Fatalf("manage_user.uuid %q != hrm_employee.user_uuid %q", user.Uuid, emp.UserUuid)
	}

	// 档案上的用户状态由 user_uuid 派生为已生成
	rec := mustDetail(t, svcCtx, empUuid)
	if rec.UserStatus != "1" {
		t.Fatalf("userStatus = %q, want 1", rec.UserStatus)
	}
}

// TestGenerateUserBatchIdempotent 批量生成 + 幂等: 已生成过账号的员工自动跳过.
func TestGenerateUserBatchIdempotent(t *testing.T) {
	svcCtx, _ := newSyncCtx(t)

	for _, name := range []string{"甲", "乙", "丙"} {
		addEmployee(t, svcCtx, &types.AddRequest{Name: name, PersonStatus: "1"})
	}
	list := mustList(t, svcCtx, types.ListRequest{})
	if len(list.Records) != 3 {
		t.Fatalf("records = %d, want 3", len(list.Records))
	}

	// 单号自动生成: 00000001 起; 找出对应 uuid→工号
	noByUuid := map[string]string{}
	for _, rec := range list.Records {
		noByUuid[rec.Uuid] = rec.EmployeeNo
	}

	resp, err := NewGenerateUser(context.Background(), svcCtx, nil).GenerateUser(&types.GenerateUserRequest{Uuids: uuidsOf(noByUuid)})
	if err != nil {
		t.Fatalf("batch generateUser: %v", err)
	}
	if resp.Generated != 3 || resp.Skipped != 0 {
		t.Fatalf("generated=%d skipped=%d, want 3/0", resp.Generated, resp.Skipped)
	}

	// 再次生成全部: 全部跳过 (幂等)
	resp2, err := NewGenerateUser(context.Background(), svcCtx, nil).GenerateUser(&types.GenerateUserRequest{Uuids: uuidsOf(noByUuid)})
	if err != nil {
		t.Fatalf("idempotent generateUser: %v", err)
	}
	if resp2.Generated != 0 || resp2.Skipped != 3 {
		t.Fatalf("generated=%d skipped=%d, want 0/3", resp2.Generated, resp2.Skipped)
	}

	// 用户名 = 工号
	for empUuid, no := range noByUuid {
		emp := mustFindEmployee(t, svcCtx, empUuid)
		user := mustFindUserByUsername(t, svcCtx, no)
		if user.Uuid != emp.UserUuid {
			t.Fatalf("employee %s: manage_user.uuid != user_uuid", no)
		}
		if user.Nickname == "" {
			t.Fatalf("employee %s: nickname empty", no)
		}
	}
}

// TestGenerateUserUnknownUuid 生成账号时传入不存在的员工 uuid → EmployeeUserNotFoundCode.
func TestGenerateUserUnknownUuid(t *testing.T) {
	svcCtx, _ := newSyncCtx(t)

	_, err := NewGenerateUser(context.Background(), svcCtx, nil).GenerateUser(&types.GenerateUserRequest{Uuids: []string{"no-such-uuid"}})
	if st := status.FromError(err); st == nil || st.Code() != errcodes.EmployeeUserNotFoundCode {
		t.Fatalf("err = %v, want EmployeeUserNotFoundCode (10460)", err)
	}
}

// TestEditSyncToUser 员工更新时同步账号资料 (昵称/手机号/邮箱).
func TestEditSyncToUser(t *testing.T) {
	svcCtx, _ := newSyncCtx(t)

	addEmployee(t, svcCtx, &types.AddRequest{Name: "李四", Phone: "13811112222", Email: "lisi@a.com", PersonStatus: "1"})
	rec := mustList(t, svcCtx, types.ListRequest{}).Records[0]
	empUuid := rec.Uuid
	origUserUuid := ""

	if _, err := NewGenerateUser(context.Background(), svcCtx, nil).GenerateUser(&types.GenerateUserRequest{Uuids: []string{empUuid}}); err != nil {
		t.Fatalf("generateUser: %v", err)
	}
	origUserUuid = mustFindEmployee(t, svcCtx, empUuid).UserUuid

	// 编辑员工资料 → 账号资料同步
	editReq := &types.EditRequest{
		Uuid:  empUuid,
		Name:  "李四改",
		Phone: "13933334444",
		Email: "lisi-new@b.com",
	}
	if _, err := NewEdit(context.Background(), svcCtx, nil).Edit(editReq); err != nil {
		t.Fatalf("edit: %v", err)
	}

	user := mustFindUserByUsername(t, svcCtx, rec.EmployeeNo)
	if user.Nickname != "李四改" {
		t.Fatalf("nickname = %q, want 李四改 (同步昵称)", user.Nickname)
	}
	if user.Phone != "13933334444" {
		t.Fatalf("phone = %q, want 13933334444 (同步手机号)", user.Phone)
	}
	if user.Email != "lisi-new@b.com" {
		t.Fatalf("email = %q, want lisi-new@b.com (同步邮箱)", user.Email)
	}
	if user.Uuid != origUserUuid {
		t.Fatal("manage_user.uuid changed on edit (应原地更新, 非重建)")
	}

	// 未生成账号的员工编辑不报错 (无账号可同步)
	addEmployee(t, svcCtx, &types.AddRequest{Name: "王五", PersonStatus: "1"})
	recs := mustList(t, svcCtx, types.ListRequest{})
	var noAccountUuid string
	for _, r := range recs.Records {
		if r.UserStatus == "0" {
			noAccountUuid = r.Uuid
		}
	}
	if noAccountUuid != "" {
		if _, err := NewEdit(context.Background(), svcCtx, nil).Edit(&types.EditRequest{Uuid: noAccountUuid, Name: "王五改"}); err != nil {
			t.Fatalf("edit employee without account: %v", err)
		}
	}
}

// TestDeleteCascadeUser 删除员工时级联删除账号.
func TestDeleteCascadeUser(t *testing.T) {
	svcCtx, _ := newSyncCtx(t)

	addEmployee(t, svcCtx, &types.AddRequest{Name: "赵六", PersonStatus: "1"})
	empUuid := mustList(t, svcCtx, types.ListRequest{}).Records[0].Uuid
	if _, err := NewGenerateUser(context.Background(), svcCtx, nil).GenerateUser(&types.GenerateUserRequest{Uuids: []string{empUuid}}); err != nil {
		t.Fatalf("generateUser: %v", err)
	}
	userUuid := mustFindEmployee(t, svcCtx, empUuid).UserUuid

	if _, err := NewDelete(context.Background(), svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{empUuid}}); err != nil {
		t.Fatalf("delete: %v", err)
	}

	if _, err := svcCtx.Model.ManageUser.FindOneByUuid(context.Background(), nil, userUuid); err != sqlx.ErrNotFound {
		t.Fatalf("manage_user after delete = %v, want ErrNotFound (级联删除)", err)
	}
}

// TestDeleteWithoutAccount 删除未生成账号的员工不报错.
func TestDeleteWithoutAccount(t *testing.T) {
	svcCtx, _ := newSyncCtx(t)

	addEmployee(t, svcCtx, &types.AddRequest{Name: "孙七", PersonStatus: "1"})
	empUuid := mustList(t, svcCtx, types.ListRequest{}).Records[0].Uuid
	if _, err := NewDelete(context.Background(), svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{empUuid}}); err != nil {
		t.Fatalf("delete employee without account: %v", err)
	}
}

func uuidsOf(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
