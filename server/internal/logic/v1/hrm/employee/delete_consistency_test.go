package employee

import (
	"context"
	"testing"

	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	coreConfig "github.com/jzero-io/jzero-admin/core-engine/config"
	coreSvc "github.com/jzero-io/jzero-admin/core-engine/svc"
	authlogic "github.com/jzero-io/jzero-admin/server/internal/logic/v1/auth"
	manageuser "github.com/jzero-io/jzero-admin/server/internal/logic/v1/manage/user"
	"github.com/jzero-io/jzero-admin/server/internal/config"
	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	authTypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/auth"
	manageUserTypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/manage/user"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/employee"
)

// HRM-52 删除一致性测试: 删员工清账号+角色绑定 (不留孤儿行), 删被引用账号拒绝,
// 解绑后可删, 批量全或无 (ADR 0007 生命周期一致性; 删除语义由应用层保证,
// 不加外键/唯一索引)。

// newDeleteCtx 删除一致性测试环境: SqlxConn (事务) + ConfigCenter
// (初始密码配置 + Jwt 登录配置 — 静态配置中心不应用 struct default, 需显式给值).
func newDeleteCtx(t *testing.T) (*svc.ServiceContext, *svctest.Env) {
	t.Helper()
	env := svctest.New(t)
	return &svc.ServiceContext{
		ServiceContext: &coreSvc.ServiceContext{SqlxConn: env.SqlConn},
		ConfigCenter: config.NewStaticConfigCenter(config.Config{
			Config: coreConfig.Config{Jwt: coreConfig.JwtConf{AccessSecret: "test-secret", AccessExpire: 3600, RefreshExpire: 86400}},
			Hrm:    config.HrmConf{InitialUserPassword: "test-pwd-123"},
		}),
		Model: env.Model,
	}, env
}

// TestDeleteEmployeeCascadesAccountAndRoleBindings 删除已关联账号的员工:
// 同一事务内账号与它的 manage_user_role 行一起消失, 不留孤儿绑定。
func TestDeleteEmployeeCascadesAccountAndRoleBindings(t *testing.T) {
	svcCtx, _ := newDeleteCtx(t)

	addEmployee(t, svcCtx, &types.AddRequest{Name: "删一致", EmployeeNo: "10000041", PersonStatus: PersonStatusRegular})
	emp := mustFindEmployeeByNo(t, svcCtx, "10000041")
	userUuid := emp.UserUuid
	if userUuid == "" {
		t.Fatal("add did not provision account (建档即开号)")
	}
	if len(userRoleRows(t, svcCtx, userUuid)) == 0 {
		t.Fatal("precondition: provisioned account has no role bindings")
	}

	if _, err := NewDelete(context.Background(), svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{emp.Uuid}}); err != nil {
		t.Fatalf("delete employee: %v", err)
	}

	if _, err := svcCtx.Model.ManageUser.FindOneByUuid(context.Background(), nil, userUuid); err != sqlx.ErrNotFound {
		t.Fatalf("manage_user after delete = %v, want ErrNotFound (账号级联删除)", err)
	}
	if rows := userRoleRows(t, svcCtx, userUuid); len(rows) != 0 {
		t.Fatalf("manage_user_role after delete = %d rows, want 0 (不留孤儿角色绑定)", len(rows))
	}
	if _, err := svcCtx.Model.HrmEmployee.FindOneByUuid(context.Background(), nil, emp.Uuid); err != sqlx.ErrNotFound {
		t.Fatalf("hrm_employee after delete = %v, want ErrNotFound", err)
	}
}

// TestDeleteReferencedAccountRejected 用户管理删除被员工档案引用的账号 → 拒绝
// (EmployeeUserReferencedCode); 被拒后账号仍在、仍能登录、档案关联指针未变。
func TestDeleteReferencedAccountRejected(t *testing.T) {
	svcCtx, _ := newDeleteCtx(t)

	addEmployee(t, svcCtx, &types.AddRequest{Name: "引用甲", EmployeeNo: "10000042", PersonStatus: PersonStatusRegular})
	emp := mustFindEmployeeByNo(t, svcCtx, "10000042")
	userUuid := emp.UserUuid
	if userUuid == "" {
		t.Fatal("add did not provision account (建档即开号)")
	}

	_, err := manageuser.NewDelete(context.Background(), svcCtx, nil).Delete(&manageUserTypes.DeleteRequest{Uuids: []string{userUuid}})
	if st := status.FromError(err); st == nil || st.Code() != errcodes.EmployeeUserReferencedCode {
		t.Fatalf("err = %v, want EmployeeUserReferencedCode (10469, 账号被员工档案引用不可删除)", err)
	}

	// 账号仍然存在且未被改动
	user := mustFindUserByUsername(t, svcCtx, "10000042")
	if user.Uuid != userUuid || user.Status != "1" || user.Password != "test-pwd-123" {
		t.Fatalf("account after rejected delete = uuid %q status %q, want unchanged", user.Uuid, user.Status)
	}
	// 档案关联指针未变
	if after := mustFindEmployee(t, svcCtx, emp.Uuid); after.UserUuid != userUuid {
		t.Fatalf("user_uuid after rejected delete = %q, want %q (指针未变)", after.UserUuid, userUuid)
	}
	// 仍然能登录
	if _, err := authlogic.NewPwdLogin(context.Background(), svcCtx, nil).PwdLogin(&authTypes.PwdLoginRequest{
		Username: "10000042", Password: "test-pwd-123",
	}); err != nil {
		t.Fatalf("login after rejected delete: %v (账号应仍可登录)", err)
	}
}

// TestUnboundAccountDeletableAfterUnlink 档案解绑 (只清 user_uuid 指针) 后,
// 同一账号可从用户管理正常删除, 账号与角色绑定一起清掉, 档案不受影响。
func TestUnboundAccountDeletableAfterUnlink(t *testing.T) {
	svcCtx, _ := newDeleteCtx(t)

	addEmployee(t, svcCtx, &types.AddRequest{Name: "解绑乙", EmployeeNo: "10000043", PersonStatus: PersonStatusRegular})
	emp := mustFindEmployeeByNo(t, svcCtx, "10000043")
	userUuid := emp.UserUuid
	if userUuid == "" {
		t.Fatal("add did not provision account (建档即开号)")
	}

	// 构造"已解绑"状态: 只清档案指针 (解绑端点归 #53, 本票手工清 user_uuid)
	if err := svcCtx.Model.HrmEmployee.UpdateFieldsByCondition(context.Background(), nil,
		map[string]any{"user_uuid": ""},
		condition.NewChain().Equal(hrm_employee.Uuid, emp.Uuid).Build()...); err != nil {
		t.Fatalf("clear user_uuid pointer: %v", err)
	}

	if _, err := manageuser.NewDelete(context.Background(), svcCtx, nil).Delete(&manageUserTypes.DeleteRequest{Uuids: []string{userUuid}}); err != nil {
		t.Fatalf("delete unbound account: %v (解绑后应可删除)", err)
	}
	if _, err := svcCtx.Model.ManageUser.FindOneByUuid(context.Background(), nil, userUuid); err != sqlx.ErrNotFound {
		t.Fatalf("manage_user after delete = %v, want ErrNotFound", err)
	}
	if rows := userRoleRows(t, svcCtx, userUuid); len(rows) != 0 {
		t.Fatalf("manage_user_role after delete = %d rows, want 0 (账号与角色绑定一起清掉)", len(rows))
	}
	// 档案仍在 (删账号不动档案), 指针保持空
	if after := mustFindEmployee(t, svcCtx, emp.Uuid); after.UserUuid != "" {
		t.Fatalf("user_uuid after account delete = %q, want empty (删账号不动档案)", after.UserUuid)
	}
}

// TestBatchDeleteAccountsAllOrNothing 批量删除混合"被引用"与"未被引用"账号:
// 整批拒绝 (全或无, 不部分成功); 单独删除未被引用的账号照常成功。
func TestBatchDeleteAccountsAllOrNothing(t *testing.T) {
	svcCtx, _ := newDeleteCtx(t)

	// 被引用账号: 建档即开号
	addEmployee(t, svcCtx, &types.AddRequest{Name: "引用丙", EmployeeNo: "10000044", PersonStatus: PersonStatusRegular})
	emp := mustFindEmployeeByNo(t, svcCtx, "10000044")
	referenced := emp.UserUuid
	if referenced == "" {
		t.Fatal("add did not provision account (建档即开号)")
	}
	// 未引用账号: 用户管理新增 (挂 R_EMPLOYEE, 迁移 63 种子角色)
	if _, err := manageuser.NewAdd(context.Background(), svcCtx, nil).Add(&manageUserTypes.AddRequest{
		Username: "loose-10000045", Password: "loose-pwd", NickName: "散账号", Status: "1", UserRoles: []string{"R_EMPLOYEE"},
	}); err != nil {
		t.Fatalf("add loose account: %v", err)
	}
	loose := mustFindUserByUsername(t, svcCtx, "loose-10000045").Uuid
	if len(userRoleRows(t, svcCtx, loose)) == 0 {
		t.Fatal("precondition: loose account has no role bindings")
	}

	// 混合批量 → 整批拒绝
	_, err := manageuser.NewDelete(context.Background(), svcCtx, nil).Delete(&manageUserTypes.DeleteRequest{Uuids: []string{loose, referenced}})
	if st := status.FromError(err); st == nil || st.Code() != errcodes.EmployeeUserReferencedCode {
		t.Fatalf("err = %v, want EmployeeUserReferencedCode (混合批量应整批拒绝)", err)
	}
	// 不出现部分成功: 两个账号都还在, 未引用账号的角色绑定也还在
	if _, err := svcCtx.Model.ManageUser.FindOneByUuid(context.Background(), nil, referenced); err != nil {
		t.Fatalf("referenced account after rejected batch: %v (整批拒绝, 无部分删除)", err)
	}
	if _, err := svcCtx.Model.ManageUser.FindOneByUuid(context.Background(), nil, loose); err != nil {
		t.Fatalf("unreferenced account after rejected batch: %v (整批拒绝, 无部分删除)", err)
	}
	if rows := userRoleRows(t, svcCtx, loose); len(rows) == 0 {
		t.Fatal("unreferenced account lost role bindings on rejected batch (无部分成功)")
	}

	// 只删未被引用的账号 → 照常成功 (账号与角色绑定一起清掉)
	if _, err := manageuser.NewDelete(context.Background(), svcCtx, nil).Delete(&manageUserTypes.DeleteRequest{Uuids: []string{loose}}); err != nil {
		t.Fatalf("delete unreferenced account: %v", err)
	}
	if _, err := svcCtx.Model.ManageUser.FindOneByUuid(context.Background(), nil, loose); err != sqlx.ErrNotFound {
		t.Fatalf("manage_user after delete = %v, want ErrNotFound", err)
	}
	if rows := userRoleRows(t, svcCtx, loose); len(rows) != 0 {
		t.Fatalf("manage_user_role after delete = %d rows, want 0", len(rows))
	}
	// 被引用账号不受影响
	if _, err := svcCtx.Model.ManageUser.FindOneByUuid(context.Background(), nil, referenced); err != nil {
		t.Fatalf("referenced account must survive: %v", err)
	}
}
