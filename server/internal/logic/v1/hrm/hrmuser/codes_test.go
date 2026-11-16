package hrmuser

// CanSeeAllBills DB-backed tests (final-review F4, spec 用户故事 4):
// 列表默认按我的单据过滤，管理员可看全部 — 判定口径：
//   - 超级管理员角色 → 可看全部；
//   - 任意角色被授予共享审批码 v1:hrm:bill:approve（仅审批按钮种子携带）→ 可看全部；
//   - 无任何授权 → 仅看自己的。

import (
	"context"
	"encoding/json"
	"testing"

	coreSvc "github.com/jzero-io/jzero-admin/core-engine/svc"
	"github.com/jzero-io/jzero-admin/server/internal/model/manage_menu"
	"github.com/jzero-io/jzero-admin/server/internal/model/manage_role_menu"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
)

// seedApproveGrant 菜单 permissions JSON 授予 code + 角色→菜单绑定。
// 菜单 uuid 后缀 99xx（migrations 未及 42xx，测试局部唯一）。
func seedApproveGrant(t *testing.T, env *svctest.Env, roleUuid, menuUuid, rmUuid, code string) {
	t.Helper()
	perms, err := json.Marshal([]map[string]string{{"code": code, "desc": "审批通过"}})
	if err != nil {
		t.Fatal(err)
	}
	if err := env.Model.ManageMenu.InsertV2(context.Background(), nil, &manage_menu.ManageMenu{
		Uuid:        menuUuid,
		Status:      "1",
		MenuType:    "3",
		MenuName:    "审批通过",
		Permissions: string(perms),
	}); err != nil {
		t.Fatalf("seed menu: %v", err)
	}
	if err := env.Model.ManageRoleMenu.InsertV2(context.Background(), nil, &manage_role_menu.ManageRoleMenu{
		Uuid:     rmUuid,
		RoleUuid: roleUuid,
		MenuUuid: menuUuid,
	}); err != nil {
		t.Fatalf("seed role menu: %v", err)
	}
}

func TestCanSeeAllBills(t *testing.T) {
	env := svctest.New(t)
	ctx := context.Background()
	svcCtx := &svc.ServiceContext{
		ServiceContext: &coreSvc.ServiceContext{SqlxConn: env.SqlConn},
		Model:          env.Model,
	}

	// 1. 超级管理员（种子角色 uuid）→ 可看全部，无需任何种子行。
	if !CanSeeAllBills(ctx, svcCtx, User{RoleUuids: []string{SuperAdminRoleUuid}}) {
		t.Fatal("superadmin must see all bills")
	}

	// 2. 持审批码的角色 → 可看全部。
	seedApproveGrant(t, env, "role-hr",
		"aaa10000-1111-4000-8000-000000009911",
		"aaa20000-1111-4000-8000-000000009911", BillApproveCode)
	if !CanSeeAllBills(ctx, svcCtx, User{RoleUuids: []string{"role-hr"}}) {
		t.Fatal("role granted v1:hrm:bill:approve must see all bills")
	}

	// 3. 无授权角色 → 仅看自己的。
	seedApproveGrant(t, env, "role-staff",
		"aaa10000-1111-4000-8000-000000009912",
		"aaa20000-1111-4000-8000-000000009912", "v1:hrm:transfer:list")
	if CanSeeAllBills(ctx, svcCtx, User{RoleUuids: []string{"role-staff"}}) {
		t.Fatal("role without approve grant must be mine-only")
	}

	// 4. 无角色 → 仅看自己的。
	if CanSeeAllBills(ctx, svcCtx, User{}) {
		t.Fatal("roleless user must be mine-only")
	}

	// 5. 混合角色（staff+hr）→ 可看全部。
	if !CanSeeAllBills(ctx, svcCtx, User{RoleUuids: []string{"role-staff", "role-hr"}}) {
		t.Fatal("mixed roles with one approve grant must see all bills")
	}
}
