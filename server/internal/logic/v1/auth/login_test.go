package auth

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	jzerocache "github.com/jzero-io/jzero/core/stores/cache"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/stores/redis"

	coreConfig "github.com/jzero-io/jzero-admin/core-engine/config"
	coreSvc "github.com/jzero-io/jzero-admin/core-engine/svc"
	"github.com/jzero-io/jzero-admin/server/internal/config"
	"github.com/jzero-io/jzero-admin/server/internal/constant"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/manage_user"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/auth"
	employeetypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/employee"
)

// Seams under test (HRM-53, ADR 0007 生命周期一致性): 登录是账号状态的真实开关,
// pwd_login 与 code_login 两条路径都必须校验 manage_user.status:
//   - '1'(启用) → 既有行为不变, 正常发 token
//   - '2'(禁用) → 明确拒绝 ("账号已禁用"), 不混淆成"用户名或密码错误"
//   - 员工离职 (person_status='5') → 关联账号被禁用 → 登录被拒; 重新启用后恢复登录
//
// 全部走 logic 方法缝 + 真实迁移 (svctest); 账号状态经 model 读 manage_user.status
// (它是状态真值列, 不是内部实现细节)。

// newLoginCtx 登录逻辑测试环境: 需要 Cache (code_login 读验证码) 与 Jwt 配置
// (StaticConfigCenter 不应用 struct default, AccessSecret 必须显式给出)。
func newLoginCtx(t *testing.T) (*svc.ServiceContext, *svctest.Env) {
	t.Helper()
	env := svctest.New(t)
	mr := miniredis.RunT(t)
	rds := redis.MustNewRedis(redis.RedisConf{Type: redis.NodeType, Host: mr.Addr()})
	return &svc.ServiceContext{
		ServiceContext: &coreSvc.ServiceContext{
			SqlxConn: env.SqlConn,
			Cache:    jzerocache.NewRedisNode(rds, errors.New("cache not found")),
		},
		ConfigCenter: config.NewStaticConfigCenter(config.Config{
			Config: coreConfig.Config{Jwt: coreConfig.JwtConf{AccessSecret: "test-secret", AccessExpire: 7200}},
			Hrm:    config.HrmConf{InitialUserPassword: "test-pwd-123"},
		}),
		Model: env.Model,
	}, env
}

func seedUser(t *testing.T, svcCtx *svc.ServiceContext, username, password, email, status string) *manage_user.ManageUser {
	t.Helper()
	row := &manage_user.ManageUser{
		Uuid:     uuid.NewString(),
		Username: username,
		Password: password,
		Nickname: username,
		Email:    email,
		Status:   status,
	}
	if err := svcCtx.Model.ManageUser.InsertV2(context.Background(), nil, row); err != nil {
		t.Fatalf("seed manage_user %s: %v", username, err)
	}
	return row
}

func mustUserByUuid(t *testing.T, svcCtx *svc.ServiceContext, userUuid string) *manage_user.ManageUser {
	t.Helper()
	row, err := svcCtx.Model.ManageUser.FindOneByUuid(context.Background(), nil, userUuid)
	if err != nil {
		t.Fatalf("find manage_user %s: %v", userUuid, err)
	}
	return row
}

func setUserStatus(t *testing.T, svcCtx *svc.ServiceContext, userUuid, status string) {
	t.Helper()
	row := mustUserByUuid(t, svcCtx, userUuid)
	row.Status = status
	if err := svcCtx.Model.ManageUser.Update(context.Background(), nil, row); err != nil {
		t.Fatalf("set manage_user.status: %v", err)
	}
}

func seedVerificationCode(t *testing.T, svcCtx *svc.ServiceContext, verificationUuid, code string) {
	t.Helper()
	if err := svcCtx.Cache.SetWithExpireCtx(context.Background(),
		fmt.Sprintf("%s:%s", constant.CacheVerificationCodePrefix, verificationUuid), code, time.Minute); err != nil {
		t.Fatalf("seed verification code: %v", err)
	}
}

func pwdLogin(svcCtx *svc.ServiceContext, username, password string) (any, error) {
	resp, err := NewPwdLogin(context.Background(), svcCtx, nil).PwdLogin(&types.PwdLoginRequest{
		Username: username, Password: password,
	})
	return resp, err
}

// TestPwdLoginEnabledAccountSucceeds 启用账号 + 密码正确 → 正常发 token (既有行为不变).
func TestPwdLoginEnabledAccountSucceeds(t *testing.T) {
	svcCtx, _ := newLoginCtx(t)
	seedUser(t, svcCtx, "emp001", "pwd-123", "emp001@a.com", "1")

	resp, err := NewPwdLogin(context.Background(), svcCtx, nil).PwdLogin(&types.PwdLoginRequest{
		Username: "emp001", Password: "pwd-123",
	})
	if err != nil {
		t.Fatalf("pwd login enabled account: %v", err)
	}
	if resp.Token == "" || resp.RefreshToken == "" {
		t.Fatalf("login response missing tokens: %+v", resp)
	}
}

// TestPwdLoginDisabledAccountRejected 禁用账号 + 密码正确 → 明确拒绝:
// 错误文案含 "账号已禁用", 不得混淆成 "用户名或密码错误" (HRM-53 验收标准)。
// 库里已存在的禁用账号从此真的登不进去 — 这是 ADR 0007 点名的预期行为变更。
func TestPwdLoginDisabledAccountRejected(t *testing.T) {
	svcCtx, _ := newLoginCtx(t)
	user := seedUser(t, svcCtx, "emp002", "pwd-123", "emp002@a.com", "1")
	setUserStatus(t, svcCtx, user.Uuid, "2")

	_, err := NewPwdLogin(context.Background(), svcCtx, nil).PwdLogin(&types.PwdLoginRequest{
		Username: "emp002", Password: "pwd-123",
	})
	if err == nil {
		t.Fatal("disabled account with correct password must not log in")
	}
	msg := err.Error()
	if !strings.Contains(msg, "账号已禁用") {
		t.Fatalf("err = %q, want 明确的账号已禁用文案 (不得让用户误试密码)", msg)
	}
	if strings.Contains(msg, "用户名或密码错误") {
		t.Fatalf("err = %q, disabled account must not be confused with wrong-password error", msg)
	}
}

// TestPwdLoginWrongPasswordStillGenericError 密码错误 → 既有通用文案不变.
func TestPwdLoginWrongPasswordStillGenericError(t *testing.T) {
	svcCtx, _ := newLoginCtx(t)
	seedUser(t, svcCtx, "emp003", "pwd-123", "emp003@a.com", "1")

	_, err := NewPwdLogin(context.Background(), svcCtx, nil).PwdLogin(&types.PwdLoginRequest{
		Username: "emp003", Password: "wrong",
	})
	if err == nil || !strings.Contains(err.Error(), "用户名或密码错误") {
		t.Fatalf("err = %v, want 既有 用户名或密码错误 文案不变", err)
	}
}

// TestCodeLoginEnabledAccountSucceeds 验证码登录路径: 启用账号 → 正常发 token.
func TestCodeLoginEnabledAccountSucceeds(t *testing.T) {
	svcCtx, _ := newLoginCtx(t)
	seedUser(t, svcCtx, "emp011", "pwd-123", "emp011@a.com", "1")
	seedVerificationCode(t, svcCtx, "v-uuid-ok", "654321")

	resp, err := NewCodeLogin(context.Background(), svcCtx, nil).CodeLogin(&types.CodeLoginRequest{
		Email: "emp011@a.com", VerificationUuid: "v-uuid-ok", VerificationCode: "654321",
	})
	if err != nil {
		t.Fatalf("code login enabled account: %v", err)
	}
	if resp.Token == "" || resp.RefreshToken == "" {
		t.Fatalf("code login response missing tokens: %+v", resp)
	}
}

// TestCodeLoginDisabledAccountRejected 验证码登录路径同样校验状态:
// 禁用账号 → 明确拒绝, 两条登录路径行为一致 (HRM-53 验收标准).
func TestCodeLoginDisabledAccountRejected(t *testing.T) {
	svcCtx, _ := newLoginCtx(t)
	user := seedUser(t, svcCtx, "emp012", "pwd-123", "emp012@a.com", "1")
	setUserStatus(t, svcCtx, user.Uuid, "2")
	seedVerificationCode(t, svcCtx, "v-uuid-disabled", "654321")

	_, err := NewCodeLogin(context.Background(), svcCtx, nil).CodeLogin(&types.CodeLoginRequest{
		Email: "emp012@a.com", VerificationUuid: "v-uuid-disabled", VerificationCode: "654321",
	})
	if err == nil {
		t.Fatal("disabled account must not log in via code login")
	}
	if !strings.Contains(err.Error(), "账号已禁用") {
		t.Fatalf("err = %q, want 明确的账号已禁用文案 (与 pwd_login 一致)", err.Error())
	}
}

// TestReEnabledAccountCanLogin 账号 '2'→'1' 后恢复正常登录 (重新启用恢复, HRM-53).
func TestReEnabledAccountCanLogin(t *testing.T) {
	svcCtx, _ := newLoginCtx(t)
	user := seedUser(t, svcCtx, "emp021", "pwd-123", "emp021@a.com", "1")
	setUserStatus(t, svcCtx, user.Uuid, "2")

	if _, err := pwdLogin(svcCtx, "emp021", "pwd-123"); err == nil {
		t.Fatal("precondition: disabled account should be rejected")
	}

	// 重新启用: 用户管理编辑 status='1' 的直接效果
	setUserStatus(t, svcCtx, user.Uuid, "1")
	resp, err := NewPwdLogin(context.Background(), svcCtx, nil).PwdLogin(&types.PwdLoginRequest{
		Username: "emp021", Password: "pwd-123",
	})
	if err != nil {
		t.Fatalf("re-enabled account must log in again: %v", err)
	}
	if resp.Token == "" {
		t.Fatalf("re-enabled login missing token: %+v", resp)
	}
}

// TestResignedEmployeeLoginRejected 人员状态置离职('5')后关联账号被禁用, 登录被拒:
// 走员工新增(建档即开号)→员工编辑(置离职) 全外部入口组合.
func TestResignedEmployeeLoginRejected(t *testing.T) {
	svcCtx, _ := newLoginCtx(t)

	if _, err := employee.NewAdd(context.Background(), svcCtx, nil).Add(&employeetypes.AddRequest{
		Name: "离职工", EmployeeNo: "EMP-R01", PersonStatus: employee.PersonStatusRegular,
	}); err != nil {
		t.Fatalf("add employee: %v", err)
	}
	emp, err := svcCtx.Model.HrmEmployee.FindOneByCondition(context.Background(), nil,
		condition.NewChain().Equal(hrm_employee.EmployeeNo, "EMP-R01").Build()...)
	if err != nil {
		t.Fatalf("find employee: %v", err)
	}
	if emp.UserUuid == "" {
		t.Fatal("precondition: 建档即开号 should have provisioned an account")
	}

	if _, err = employee.NewEdit(context.Background(), svcCtx, nil).Edit(&employeetypes.EditRequest{
		Uuid: emp.Uuid, Name: "离职工", PersonStatus: employee.PersonStatusResigned,
	}); err != nil {
		t.Fatalf("edit employee to resigned: %v", err)
	}

	user := mustUserByUuid(t, svcCtx, emp.UserUuid)
	if user.Status != "2" {
		t.Fatalf("resigned employee account status = %q, want 2 (离职置禁用)", user.Status)
	}

	_, err = NewPwdLogin(context.Background(), svcCtx, nil).PwdLogin(&types.PwdLoginRequest{
		Username: "EMP-R01", Password: "test-pwd-123",
	})
	if err == nil || !strings.Contains(err.Error(), "账号已禁用") {
		t.Fatalf("err = %v, want 离职员工登录被拒 (账号已禁用)", err)
	}
}
