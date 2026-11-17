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

// addEmployeeWithoutAccount 造一份"未关联账号"的档案 (HRM-50 后新增即开号, 直接
// addEmployee 必然带账号)。仅走外部行为: 离职补录新增 (接缝跳过开号) → 编辑转在职
// (编辑不产号) — 对应存量档案「生成用户」按钮的修复语义。
func addEmployeeWithoutAccount(t *testing.T, svcCtx *svc.ServiceContext, req *types.AddRequest) {
	t.Helper()
	resigned := PersonStatusResigned
	if req.PersonStatus == resigned {
		t.Fatal("addEmployeeWithoutAccount 需要目标在职状态, 离职请直接 addEmployee")
	}
	orig := req.PersonStatus
	req.PersonStatus = resigned
	addEmployee(t, svcCtx, req)
	row := mustFindEmployeeByNo(t, svcCtx, req.EmployeeNo)
	// Edit 是全量更新, 把已落库字段原样带回, 只翻人员状态 (编辑不产号)
	if _, err := NewEdit(context.Background(), svcCtx, nil).Edit(&types.EditRequest{
		Uuid:         row.Uuid,
		EmployeeNo:   row.EmployeeNo,
		Name:         row.Name,
		Gender:       row.Gender,
		Phone:        row.Phone,
		Email:        row.Email,
		PersonStatus: orig,
	}); err != nil {
		t.Fatalf("flip person status to on-duty: %v", err)
	}
}

// mustFindEmployeeByNo 按工号定位档案 (列表 id DESC, 不依赖顺序)。
func mustFindEmployeeByNo(t *testing.T, svcCtx *svc.ServiceContext, employeeNo string) *hrm_employee.HrmEmployee {
	t.Helper()
	row, err := svcCtx.Model.HrmEmployee.FindOneByCondition(context.Background(), nil,
		condition.NewChain().Equal(hrm_employee.EmployeeNo, employeeNo).Build()...)
	if err != nil {
		t.Fatalf("find employee by no %s: %v", employeeNo, err)
	}
	return row
}

// TestGenerateUserSingle 批量端点传一个 uuid = 单个生成:
// manage_user 行 (username=工号, nickname=姓名, 初始密码取配置) + hrm_employee.user_uuid 回写.
// HRM-50 建档即开号后, 未关联账号的档案只能由存量数据/离职补录产生, 用
// addEmployeeWithoutAccount 造 (离职新增→编辑转在职, 全外部行为)。
func TestGenerateUserSingle(t *testing.T) {
	svcCtx, _ := newSyncCtx(t)

	addEmployeeWithoutAccount(t, svcCtx, &types.AddRequest{Name: "张三", EmployeeNo: "10000001", PersonStatus: "1"})
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
		addEmployeeWithoutAccount(t, svcCtx, &types.AddRequest{Name: name, PersonStatus: "1"})
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

	// 未生成账号的员工编辑不报错 (无账号可同步): 离职补录造无账号档案
	addEmployeeWithoutAccount(t, svcCtx, &types.AddRequest{Name: "王五", PersonStatus: "1"})
	recs := mustList(t, svcCtx, types.ListRequest{})
	var noAccountUuid string
	for _, r := range recs.Records {
		if r.UserStatus == "0" {
			noAccountUuid = r.Uuid
		}
	}
	if noAccountUuid == "" {
		t.Fatal("no employee without account found (addEmployeeWithoutAccount 失效)")
	}
	if _, err := NewEdit(context.Background(), svcCtx, nil).Edit(&types.EditRequest{Uuid: noAccountUuid, Name: "王五改"}); err != nil {
		t.Fatalf("edit employee without account: %v", err)
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

	addEmployeeWithoutAccount(t, svcCtx, &types.AddRequest{Name: "孙七", PersonStatus: "1"})
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

// ---- HRM-50 建档即开号: 新增/导入入口外部行为 (ADR 0007) ----

// TestAddProvisionsAccountAndReturnsCredentials 新增入口: 建档即开号, 响应回带
// userCreated/username/initialPassword; 账号字段取档案, 挂默认角色。
func TestAddProvisionsAccountAndReturnsCredentials(t *testing.T) {
	svcCtx, _ := newSyncCtx(t)

	resp, err := NewAdd(context.Background(), svcCtx, nil).Add(&types.AddRequest{
		Name: "钱开", EmployeeNo: "10000021", Gender: "1",
		Phone: "13811110000", Email: "qian@a.com", PersonStatus: PersonStatusRegular,
	})
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if !resp.UserCreated || resp.Username != "10000021" {
		t.Fatalf("resp = %+v, want userCreated=true username=10000021", resp)
	}
	if resp.InitialPassword != "test-pwd-123" {
		t.Fatalf("initialPassword = %q, want test-pwd-123 (配置 Hrm.InitialUserPassword)", resp.InitialPassword)
	}

	emp := mustFindEmployeeByNo(t, svcCtx, "10000021")
	user := mustFindUserByUsername(t, svcCtx, "10000021")
	if user.Uuid != emp.UserUuid {
		t.Fatalf("manage_user.uuid %q != hrm_employee.user_uuid %q", user.Uuid, emp.UserUuid)
	}
	if user.Password != "test-pwd-123" || user.Status != "1" || user.Nickname != "钱开" ||
		user.Phone != "13811110000" || user.Email != "qian@a.com" || user.Gender != "1" {
		t.Fatalf("account fields mismatch: %+v", user)
	}
	// 挂默认角色 (DefaultRoleCode 空兜底 R_EMPLOYEE, 迁移 63 种子该角色)
	role := mustFindRoleByCode(t, svcCtx, "R_EMPLOYEE")
	rows := userRoleRows(t, svcCtx, user.Uuid)
	if len(rows) != 1 || rows[0].RoleUuid != role.Uuid {
		t.Fatalf("role rows = %+v, want exactly R_EMPLOYEE", rows)
	}
}

// TestAddResignedSkipsProvisioning 新增离职补录 (person_status='5') 不产账号:
// 响应 userCreated=false, 无 manage_user 行, 档案 user_uuid 空。
func TestAddResignedSkipsProvisioning(t *testing.T) {
	svcCtx, _ := newSyncCtx(t)

	resp, err := NewAdd(context.Background(), svcCtx, nil).Add(&types.AddRequest{
		Name: "离补录", EmployeeNo: "10000022", PersonStatus: PersonStatusResigned,
	})
	if err != nil {
		t.Fatalf("add resigned: %v", err)
	}
	if resp.UserCreated {
		t.Fatalf("resp = %+v, want userCreated=false (离职跳过开号)", resp)
	}
	if resp.Username != "" || resp.InitialPassword != "" {
		t.Fatalf("resigned add must not return credentials: %+v", resp)
	}
	if _, err = svcCtx.Model.ManageUser.FindOneByCondition(context.Background(), nil,
		condition.NewChain().Equal(manage_user.Username, "10000022").Build()...); err != sqlx.ErrNotFound {
		t.Fatalf("manage_user for resigned employee = %v, want ErrNotFound (不产账号)", err)
	}
	if emp := mustFindEmployeeByNo(t, svcCtx, "10000022"); emp.UserUuid != "" {
		t.Fatalf("user_uuid = %q, want empty", emp.UserUuid)
	}
}

// TestAddEmployeeNoConflictRollsBackArchive 新增时工号与已有账号用户名冲突:
// 新增失败 (10461) 且档案不落库 (同一事务整体回滚)。
func TestAddEmployeeNoConflictRollsBackArchive(t *testing.T) {
	svcCtx, _ := newSyncCtx(t)

	if err := svcCtx.Model.ManageUser.InsertV2(context.Background(), nil, &manage_user.ManageUser{
		Uuid: uuidNew(), Username: "80000001", Password: "x", Nickname: "历史脏账号", Status: "1",
	}); err != nil {
		t.Fatalf("seed conflicting account: %v", err)
	}

	_, err := NewAdd(context.Background(), svcCtx, nil).Add(&types.AddRequest{
		Name: "撞号", EmployeeNo: "80000001", PersonStatus: PersonStatusRegular,
	})
	if st := status.FromError(err); st == nil || st.Code() != errcodes.EmployeeUserNoDuplicateCode {
		t.Fatalf("err = %v, want EmployeeUserNoDuplicateCode (10461)", err)
	}
	// 档案不落库
	if _, err = svcCtx.Model.HrmEmployee.FindOneByCondition(context.Background(), nil,
		condition.NewChain().Equal(hrm_employee.EmployeeNo, "80000001").Build()...); err != sqlx.ErrNotFound {
		t.Fatalf("hrm_employee after failed add = %v, want ErrNotFound (档案回滚)", err)
	}
	// 冲突账号未被改动
	if user := mustFindUserByUsername(t, svcCtx, "80000001"); user.Nickname != "历史脏账号" {
		t.Fatalf("conflict account mutated: %+v", user)
	}
}

// TestImportProvisionsAccountsAllRows 导入入口: 每一行员工都获得账号
// (显式工号行 + 空工号自动生成行), 用户名=工号。
func TestImportProvisionsAccountsAllRows(t *testing.T) {
	svcCtx, _ := newCtx(t)

	content := buildImportXlsx(t, [][]string{
		{"90000001", "导甲", "男", "13900000001", "dj@a.com", "", "", "", "", "正式", "", "", "", "", "", "", "", "", "", "", "", "", "", "", ""},
		{"", "导乙", "女", "13900000002", "de@a.com", "", "", "", "", "正式", "", "", "", "", "", "", "", "", "", "", "", "", "", "", ""},
	})
	resp, err := NewImport(context.Background(), svcCtx, nil).Import(&types.ImportRequest{FileName: "员工.xlsx", Content: content})
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if resp.SuccessCount != 2 {
		t.Fatalf("successCount = %d, want 2", resp.SuccessCount)
	}

	recs := mustList(t, svcCtx, types.ListRequest{})
	noByUuid := map[string]string{}
	for _, rec := range recs.Records {
		noByUuid[rec.Uuid] = rec.EmployeeNo
	}
	if len(noByUuid) != 2 {
		t.Fatalf("records = %d, want 2", len(noByUuid))
	}
	for empUuid, no := range noByUuid {
		emp := mustFindEmployee(t, svcCtx, empUuid)
		if emp.UserUuid == "" {
			t.Fatalf("employee %s has no account (导入行未开号)", no)
		}
		user := mustFindUserByUsername(t, svcCtx, no)
		if user.Uuid != emp.UserUuid {
			t.Fatalf("employee %s: user_uuid pointer mismatch", no)
		}
		if user.Status != "1" {
			t.Fatalf("employee %s account status = %q, want 1", no, user.Status)
		}
	}
}

// TestImportResignedRowSkipsAccount 导入离职行不产账号, 其余行照常开号。
func TestImportResignedRowSkipsAccount(t *testing.T) {
	svcCtx, _ := newCtx(t)

	content := buildImportXlsx(t, [][]string{
		{"90000011", "在导入", "男", "", "", "", "", "", "", "正式", "", "", "", "", "", "", "", "", "", "", "", "", "", "", ""},
		{"90000012", "离导入", "女", "", "", "", "", "", "", "离职", "", "", "", "", "", "", "", "", "", "", "", "", "", "", ""},
	})
	resp, err := NewImport(context.Background(), svcCtx, nil).Import(&types.ImportRequest{FileName: "员工.xlsx", Content: content})
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if resp.SuccessCount != 2 {
		t.Fatalf("successCount = %d, want 2", resp.SuccessCount)
	}

	user := mustFindUserByUsername(t, svcCtx, "90000011")
	if user.Uuid == "" {
		t.Fatal("on-duty import row has no account")
	}
	if _, err = svcCtx.Model.ManageUser.FindOneByCondition(context.Background(), nil,
		condition.NewChain().Equal(manage_user.Username, "90000012").Build()...); err != sqlx.ErrNotFound {
		t.Fatalf("resigned import row got an account, want none: %v", err)
	}
	if emp := mustFindEmployeeByNo(t, svcCtx, "90000012"); emp.UserUuid != "" {
		t.Fatalf("resigned row user_uuid = %q, want empty", emp.UserUuid)
	}
}

// TestImportConflictRollsBackWholeTable 导入语义收紧 (ADR 0007): 任一行开号失败
// (工号撞已有账号用户名) 整表回滚 — 本来能成功的行也不落库。
func TestImportConflictRollsBackWholeTable(t *testing.T) {
	svcCtx, _ := newCtx(t)

	if err := svcCtx.Model.ManageUser.InsertV2(context.Background(), nil, &manage_user.ManageUser{
		Uuid: uuidNew(), Username: "90000021", Password: "x", Nickname: "冲突账号", Status: "1",
	}); err != nil {
		t.Fatalf("seed conflicting account: %v", err)
	}

	// 行1 (能成功) + 行2 (工号撞已有账号用户名) → 整表回滚
	content := buildImportXlsx(t, [][]string{
		{"90000022", "能成功行", "男", "", "", "", "", "", "", "正式", "", "", "", "", "", "", "", "", "", "", "", "", "", "", ""},
		{"90000021", "撞号行", "女", "", "", "", "", "", "", "正式", "", "", "", "", "", "", "", "", "", "", "", "", "", "", ""},
	})
	_, err := NewImport(context.Background(), svcCtx, nil).Import(&types.ImportRequest{FileName: "员工.xlsx", Content: content})
	if st := status.FromError(err); st == nil || st.Code() != errcodes.EmployeeUserNoDuplicateCode {
		t.Fatalf("err = %v, want EmployeeUserNoDuplicateCode (10461)", err)
	}

	// 整表回滚: 档案一行都不落库
	recs := mustList(t, svcCtx, types.ListRequest{})
	if len(recs.Records) != 0 {
		t.Fatalf("records after failed import = %d, want 0 (整表回滚, 含能成功的行)", len(recs.Records))
	}
	// 行1 的账号也未留下
	if _, err = svcCtx.Model.ManageUser.FindOneByCondition(context.Background(), nil,
		condition.NewChain().Equal(manage_user.Username, "90000022").Build()...); err != sqlx.ErrNotFound {
		t.Fatalf("row1 account survived rollback: %v, want ErrNotFound", err)
	}
	// 冲突账号未被改动
	if user := mustFindUserByUsername(t, svcCtx, "90000021"); user.Nickname != "冲突账号" {
		t.Fatalf("conflict account mutated: %+v", user)
	}
}

// TestEditEmployeeNoKeepsLoginName 改工号不重命名账号 (ADR 0007 用户名与工号解耦):
// 编辑工号后 manage_user.username 不变, 列表 loginName 与新工号不一致可见。
func TestEditEmployeeNoKeepsLoginName(t *testing.T) {
	svcCtx, _ := newSyncCtx(t)

	addEmployee(t, svcCtx, &types.AddRequest{Name: "改号工", EmployeeNo: "10000031", PersonStatus: PersonStatusRegular})
	emp := mustFindEmployeeByNo(t, svcCtx, "10000031")
	if emp.UserUuid == "" {
		t.Fatal("add did not provision account")
	}
	origUser := mustFindUserByUsername(t, svcCtx, "10000031")

	if _, err := NewEdit(context.Background(), svcCtx, nil).Edit(&types.EditRequest{
		Uuid: emp.Uuid, EmployeeNo: "10000099", Name: "改号工", PersonStatus: PersonStatusRegular,
	}); err != nil {
		t.Fatalf("edit employeeNo: %v", err)
	}

	// 账号原地未动: 同一 uuid, username 仍是旧工号
	user := mustFindUserByUsername(t, svcCtx, "10000031")
	if user.Uuid != origUser.Uuid {
		t.Fatal("manage_user.uuid changed (账号被重建?)")
	}
	// 新工号下无同名账号
	if _, err := svcCtx.Model.ManageUser.FindOneByCondition(context.Background(), nil,
		condition.NewChain().Equal(manage_user.Username, "10000099").Build()...); err != sqlx.ErrNotFound {
		t.Fatalf("account renamed to new employeeNo (登录凭证不得漂移): %v", err)
	}
	// 档案工号已更新; 列表 loginName 仍旧账号名 (二者可不一致)
	empAfter := mustFindEmployeeByNo(t, svcCtx, "10000099")
	if empAfter.UserUuid != emp.UserUuid {
		t.Fatal("user_uuid pointer lost on employeeNo edit")
	}
	rec := mustDetail(t, svcCtx, emp.Uuid)
	if rec.LoginName != "10000031" {
		t.Fatalf("loginName = %q, want 10000031 (与工号解耦, 改号不重命名)", rec.LoginName)
	}
	if rec.EmployeeNo != "10000099" {
		t.Fatalf("employeeNo = %q, want 10000099", rec.EmployeeNo)
	}
}
