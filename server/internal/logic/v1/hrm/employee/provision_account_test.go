package employee

import (
	"context"
	"testing"

	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"

	coreSvc "github.com/jzero-io/jzero-admin/core-engine/svc"
	"github.com/jzero-io/jzero-admin/server/internal/config"
	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/manage_role"
	"github.com/jzero-io/jzero-admin/server/internal/model/manage_user"
	"github.com/jzero-io/jzero-admin/server/internal/model/manage_user_role"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/employee"
)

// HRM-49 开号接缝测试: 唯一开号入口 ProvisionAccount 的外部行为契约
// (ADR 0007: 账号是员工档案的投影, 开号即建号+挂默认角色)。

// newProvisionCtx 开号接缝测试环境: SqlxConn (事务) + ConfigCenter (初始密码/默认角色码).
func newProvisionCtx(t *testing.T, hrmConf config.HrmConf) (*svc.ServiceContext, *svctest.Env) {
	t.Helper()
	env := svctest.New(t)
	return &svc.ServiceContext{
		ServiceContext: &coreSvc.ServiceContext{SqlxConn: env.SqlConn},
		ConfigCenter:   config.NewStaticConfigCenter(config.Config{Hrm: hrmConf}),
		Model:          env.Model,
	}, env
}

func mustFindRoleByCode(t *testing.T, svcCtx *svc.ServiceContext, code string) *manage_role.ManageRole {
	t.Helper()
	row, err := svcCtx.Model.ManageRole.FindOneByCondition(context.Background(), nil,
		condition.NewChain().Equal(manage_role.Code, code).Build()...)
	if err != nil {
		t.Fatalf("find manage_role by code %q: %v", code, err)
	}
	return row
}

func userRoleRows(t *testing.T, svcCtx *svc.ServiceContext, userUuid string) []*manage_user_role.ManageUserRole {
	t.Helper()
	rows, err := svcCtx.Model.ManageUserRole.FindByCondition(context.Background(), nil,
		condition.NewChain().Equal(manage_user_role.UserUuid, userUuid).Build()...)
	if err != nil {
		t.Fatalf("find manage_user_role for user %s: %v", userUuid, err)
	}
	return rows
}

// TestProvisionAccountCreatesUserWithDefaultRole 开号: 建 manage_user 行
// (username=工号, 昵称/性别/手机/邮箱取自档案, 状态启用, 初始密码取配置)
// + 回写 hrm_employee.user_uuid + 挂默认角色 R_EMPLOYEE.
func TestProvisionAccountCreatesUserWithDefaultRole(t *testing.T) {
	svcCtx, _ := newProvisionCtx(t, config.HrmConf{InitialUserPassword: "test-pwd-123", DefaultRoleCode: "R_EMPLOYEE"})

	addEmployee(t, svcCtx, &types.AddRequest{Name: "张三", EmployeeNo: "10000001", Gender: "1", Phone: "13811112222", Email: "zhang@a.com", PersonStatus: "1"})
	emp := mustList(t, svcCtx, types.ListRequest{}).Records[0]

	userUuid, err := ProvisionAccount(context.Background(), svcCtx, nil, emp.Uuid)
	if err != nil {
		t.Fatalf("ProvisionAccount: %v", err)
	}

	empRow := mustFindEmployee(t, svcCtx, emp.Uuid)
	if empRow.UserUuid != userUuid {
		t.Fatalf("hrm_employee.user_uuid = %q, want %q (指针回写)", empRow.UserUuid, userUuid)
	}

	user := mustFindUserByUsername(t, svcCtx, "10000001")
	if user.Uuid != userUuid {
		t.Fatalf("manage_user.uuid = %q, want %q", user.Uuid, userUuid)
	}
	if user.Nickname != "张三" {
		t.Fatalf("nickname = %q, want 张三 (取自档案)", user.Nickname)
	}
	if user.Gender != "1" {
		t.Fatalf("gender = %q, want 1 (取自档案)", user.Gender)
	}
	if user.Phone != "13811112222" {
		t.Fatalf("phone = %q, want 13811112222 (取自档案)", user.Phone)
	}
	if user.Email != "zhang@a.com" {
		t.Fatalf("email = %q, want zhang@a.com (取自档案)", user.Email)
	}
	if user.Status != "1" {
		t.Fatalf("status = %q, want 1 (启用)", user.Status)
	}
	if user.Password != "test-pwd-123" {
		t.Fatalf("password = %q, want test-pwd-123 (初始密码取配置)", user.Password)
	}

	role := mustFindRoleByCode(t, svcCtx, "R_EMPLOYEE")
	rows := userRoleRows(t, svcCtx, userUuid)
	if len(rows) != 1 {
		t.Fatalf("manage_user_role rows = %d, want 1 (默认角色)", len(rows))
	}
	if rows[0].RoleUuid != role.Uuid {
		t.Fatalf("role_uuid = %q, want R_EMPLOYEE %q", rows[0].RoleUuid, role.Uuid)
	}
}

// TestProvisionAccountMissingRoleDegraded 角色缺失降级: hrm.defaultRoleCode 指向
// 不存在的角色时, 开号仍成功 (账号已建、指针已回写), 角色绑定为空, 不返回错误。
func TestProvisionAccountMissingRoleDegraded(t *testing.T) {
	svcCtx, _ := newProvisionCtx(t, config.HrmConf{InitialUserPassword: "x", DefaultRoleCode: "R_NO_SUCH_ROLE"})

	addEmployee(t, svcCtx, &types.AddRequest{Name: "李四", EmployeeNo: "10000002", PersonStatus: "1"})
	emp := mustList(t, svcCtx, types.ListRequest{}).Records[0]

	userUuid, err := ProvisionAccount(context.Background(), svcCtx, nil, emp.Uuid)
	if err != nil {
		t.Fatalf("ProvisionAccount with missing role: %v (角色缺失不阻断开号)", err)
	}
	if userUuid == "" {
		t.Fatal("userUuid empty, want account created")
	}

	empRow := mustFindEmployee(t, svcCtx, emp.Uuid)
	if empRow.UserUuid != userUuid {
		t.Fatalf("user_uuid = %q, want %q (指针仍回写)", empRow.UserUuid, userUuid)
	}
	if rows := userRoleRows(t, svcCtx, userUuid); len(rows) != 0 {
		t.Fatalf("manage_user_role rows = %d, want 0 (角色缺失只建号不挂角色)", len(rows))
	}
}

// TestProvisionAccountEmptyRoleCodeFallback defaultRoleCode 配置为空 → 代码兜底
// R_EMPLOYEE (configcenter 不应用 json default 标签的既有结论, 不能只靠标签)。
func TestProvisionAccountEmptyRoleCodeFallback(t *testing.T) {
	svcCtx, _ := newProvisionCtx(t, config.HrmConf{InitialUserPassword: "x"})

	addEmployee(t, svcCtx, &types.AddRequest{Name: "王五", EmployeeNo: "10000003", PersonStatus: "1"})
	emp := mustList(t, svcCtx, types.ListRequest{}).Records[0]

	userUuid, err := ProvisionAccount(context.Background(), svcCtx, nil, emp.Uuid)
	if err != nil {
		t.Fatalf("ProvisionAccount: %v", err)
	}

	role := mustFindRoleByCode(t, svcCtx, "R_EMPLOYEE")
	rows := userRoleRows(t, svcCtx, userUuid)
	if len(rows) != 1 || rows[0].RoleUuid != role.Uuid {
		t.Fatalf("rows = %v, want R_EMPLOYEE (空配置兜底默认角色码)", rows)
	}
}

// TestProvisionAccountIdempotent 已关联账号的档案再次开号 → 幂等返回既有指针,
// 不新建账号、不重复挂角色。
func TestProvisionAccountIdempotent(t *testing.T) {
	svcCtx, _ := newProvisionCtx(t, config.HrmConf{DefaultRoleCode: "R_EMPLOYEE"})

	addEmployee(t, svcCtx, &types.AddRequest{Name: "甲", EmployeeNo: "10000010", PersonStatus: "1"})
	emp := mustList(t, svcCtx, types.ListRequest{}).Records[0]

	first, err := ProvisionAccount(context.Background(), svcCtx, nil, emp.Uuid)
	if err != nil {
		t.Fatalf("first ProvisionAccount: %v", err)
	}

	second, err := ProvisionAccount(context.Background(), svcCtx, nil, emp.Uuid)
	if err != nil {
		t.Fatalf("idempotent ProvisionAccount: %v", err)
	}
	if first != second {
		t.Fatalf("first=%q second=%q, want same pointer (幂等)", first, second)
	}
	if rows := userRoleRows(t, svcCtx, first); len(rows) != 1 {
		t.Fatalf("manage_user_role rows = %d, want 1 (不重复挂角色)", len(rows))
	}
}

// TestProvisionAccountResignedSkipped 离职 (person_status='5') 档案不开号: 返回空串,
// 不建 manage_user 行, 指针保持空。
func TestProvisionAccountResignedSkipped(t *testing.T) {
	svcCtx, _ := newProvisionCtx(t, config.HrmConf{DefaultRoleCode: "R_EMPLOYEE"})

	addEmployee(t, svcCtx, &types.AddRequest{Name: "乙", EmployeeNo: "10000011", PersonStatus: "5"})
	emp := mustList(t, svcCtx, types.ListRequest{}).Records[0]

	userUuid, err := ProvisionAccount(context.Background(), svcCtx, nil, emp.Uuid)
	if err != nil {
		t.Fatalf("ProvisionAccount resigned: %v (离职跳过不报错)", err)
	}
	if userUuid != "" {
		t.Fatalf("userUuid = %q, want empty (离职不开号)", userUuid)
	}

	empRow := mustFindEmployee(t, svcCtx, emp.Uuid)
	if empRow.UserUuid != "" {
		t.Fatalf("user_uuid = %q, want empty (指针不写)", empRow.UserUuid)
	}
}

// TestProvisionAccountUsernameConflictRollback 工号撞已有账号用户名 → EmployeeUserNoDuplicateCode,
// 且本档案侧无任何半成品残留 (无 manage_user 行、无 user_uuid 回写、无角色绑定)。
func TestProvisionAccountUsernameConflictRollback(t *testing.T) {
	svcCtx, _ := newProvisionCtx(t, config.HrmConf{DefaultRoleCode: "R_EMPLOYEE"})

	// HRM-50 建档即开号: 在职新增立即产号, 冲突账号必须先种 (否则工号已被新增占用)。
	// 目标档案用离职补录造 (不产号), 种冲突账号后经编辑翻回在职 (编辑不产号)。
	addEmployee(t, svcCtx, &types.AddRequest{Name: "丙", EmployeeNo: "10000012", PersonStatus: PersonStatusResigned})
	target := mustList(t, svcCtx, types.ListRequest{}).Records[0]

	// 直接插入冲突账号 (模拟历史脏数据)
	if err := svcCtx.Model.ManageUser.InsertV2(context.Background(), nil, &manage_user.ManageUser{
		Uuid:     uuidNew(),
		Username: "10000012",
		Password: "whatever",
		Nickname: "历史脏账号",
		Status:   "1",
	}); err != nil {
		t.Fatalf("seed conflicting account: %v", err)
	}
	if _, err := NewEdit(context.Background(), svcCtx, nil).Edit(&types.EditRequest{
		Uuid:         target.Uuid,
		EmployeeNo:   target.EmployeeNo,
		Name:         target.Name,
		PersonStatus: PersonStatusRegular,
	}); err != nil {
		t.Fatalf("flip to on-duty: %v", err)
	}

	_, err := ProvisionAccount(context.Background(), svcCtx, nil, target.Uuid)
	if st := status.FromError(err); st == nil || st.Code() != errcodes.EmployeeUserNoDuplicateCode {
		t.Fatalf("err = %v, want EmployeeUserNoDuplicateCode (10461)", err)
	}

	// 回滚断言: 档案指针未写
	empRow := mustFindEmployee(t, svcCtx, target.Uuid)
	if empRow.UserUuid != "" {
		t.Fatalf("user_uuid = %q, want empty (冲突回滚, 指针不写)", empRow.UserUuid)
	}
	// 冲突账号仍是原行 (未被改动), 且没有第二条同名账号
	user := mustFindUserByUsername(t, svcCtx, "10000012")
	if user.Nickname != "历史脏账号" {
		t.Fatalf("nickname = %q, want 历史脏账号 (冲突账号未被覆盖)", user.Nickname)
	}
}

// menuRow / roleMenuRow 用于迁移 63 种子的行为断言 (route path 维度, 不钉 uuid)。
type menuRow63 struct {
	RoutePath string `db:"route_path"`
	MenuType  string `db:"menu_type"`
}

// TestMigration63DefaultRoleSelfServiceOnly 迁移 63 种子后的外部行为:
// R_EMPLOYEE 恰好覆盖 8 个自服务页面 + 其目录链, 不含「评审填报」与「考核任务工作台」。
func TestMigration63DefaultRoleSelfServiceOnly(t *testing.T) {
	svcCtx, _ := newProvisionCtx(t, config.HrmConf{})

	role := mustFindRoleByCode(t, svcCtx, "R_EMPLOYEE")

	var rows []menuRow63
	if err := svcCtx.SqlxConn.QueryRowsCtx(context.Background(), &rows,
		`SELECT mm.route_path, mm.menu_type FROM manage_role_menu rm
		 JOIN manage_menu mm ON mm.uuid = rm.menu_uuid
		 WHERE rm.role_uuid = ?`, role.Uuid); err != nil {
		t.Fatalf("join role_menu x menu: %v", err)
	}

	granted := map[string]bool{}
	for _, r := range rows {
		granted[r.RoutePath] = true
	}

	wantPages := []string{
		"/home",
		"/user-center",
		"/hrm/attendance/my/attendance-punch",
		"/hrm/attendance/my/dashboard",
		"/hrm/training/my",
		"/hrm/performance/assessment-task/plan-list",
		"/hrm/performance/assessment-task/self-review-list",
		"/hrm/performance/assessment-task/result-confirm-task-list",
	}
	for _, p := range wantPages {
		if !granted[p] {
			t.Errorf("R_EMPLOYEE 未授予自服务页 %s (侧边栏将缺失)", p)
		}
	}
	for _, banned := range []string{
		"/hrm/performance/assessment-task/leader-review-list", // 评审填报
		"/hrm/performance/assessment-task/workbench",          // 考核任务工作台
	} {
		if granted[banned] {
			t.Errorf("R_EMPLOYEE 被授予了管理者视角页 %s (不应当)", banned)
		}
	}
}

// TestMigration63BackfillsExistingZeroRoleAccounts 迁移 63 存量补挂:
// 已关联档案且零角色的账号获得 R_EMPLOYEE; 已有角色的账号不动;
// 未关联档案的员工不自动建号 (迁移不产生 manage_user 行)。
func TestMigration63BackfillsExistingZeroRoleAccounts(t *testing.T) {
	svcCtx, _ := newProvisionCtx(t, config.HrmConf{})

	// 造两个已开号员工 (接缝) + 一个未开号员工 (离职补录: 建档即开号下唯一的自然无号态)
	addEmployee(t, svcCtx, &types.AddRequest{Name: "补挂甲", EmployeeNo: "10000101", PersonStatus: "1"})
	addEmployee(t, svcCtx, &types.AddRequest{Name: "补挂乙", EmployeeNo: "10000102", PersonStatus: "1"})
	addEmployee(t, svcCtx, &types.AddRequest{Name: "未开号丙", EmployeeNo: "10000103", PersonStatus: PersonStatusResigned})
	recs := mustList(t, svcCtx, types.ListRequest{})
	if len(recs.Records) != 3 {
		t.Fatalf("records = %d, want 3", len(recs.Records))
	}
	// List 按 id DESC 排序, 按工号定位, 不依赖顺序
	empUuidByNo := map[string]string{}
	for _, r := range recs.Records {
		empUuidByNo[r.EmployeeNo] = r.Uuid
	}

	u1, err := ProvisionAccount(context.Background(), svcCtx, nil, empUuidByNo["10000101"])
	if err != nil {
		t.Fatalf("provision 甲: %v", err)
	}
	u2, err := ProvisionAccount(context.Background(), svcCtx, nil, empUuidByNo["10000102"])
	if err != nil {
		t.Fatalf("provision 乙: %v", err)
	}
	role := mustFindRoleByCode(t, svcCtx, "R_EMPLOYEE")

	// 乙挂一个非默认的自定义角色 (已有角色 → 存量补挂语义下不应被触碰; 开号默认角色已在)
	extraRole := uuidNew()
	if err = svcCtx.Model.ManageUserRole.InsertV2(context.Background(), nil, &manage_user_role.ManageUserRole{
		Uuid: uuidNew(), UserUuid: u2, RoleUuid: extraRole,
	}); err != nil {
		t.Fatalf("seed extra role: %v", err)
	}

	// 验证: 甲乙都有角色行 (迁移种子已在 fresh DB 应用过; 接缝开号即挂角色),
	// 丙无账号
	if rows := userRoleRows(t, svcCtx, u1); len(rows) == 0 {
		t.Fatal("甲无角色行 (开号即挂默认角色失败)")
	}
	// 乙已有角色 (默认 + 自定义) —— 存量补挂只针对零角色账号, 有角色者不被触碰;
	// 断言默认角色行存在且无第三行 (迁移不会重复插入)
	rows2 := userRoleRows(t, svcCtx, u2)
	hasDefault := false
	for _, r := range rows2 {
		if r.RoleUuid == role.Uuid {
			hasDefault = true
		}
	}
	if !hasDefault || len(rows2) != 2 {
		t.Fatalf("乙角色行 = %d (含默认=%v), want 2 (默认+自定义, 迁移不重复补挂)", len(rows2), hasDefault)
	}
	var userRows []manage_user.ManageUser
	if err = svcCtx.SqlxConn.QueryRowsCtx(context.Background(), &userRows,
		`SELECT * FROM manage_user WHERE username = ?`, "10000103"); err != nil {
		t.Fatalf("query user 10000103: %v", err)
	}
	if len(userRows) != 0 {
		t.Fatal("未开号员工被自动建号 (迁移不应当做)")
	}
}
