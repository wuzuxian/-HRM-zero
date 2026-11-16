// Role → permission-code resolution shared by HRM logics that must decide
// "does this user hold code X" at the logic layer (authx middleware casbin only
// enforces the endpoint route code; the Auth helper carries no permission-code
// list). Resolution follows the same chain as custom.InitCasbinRule:
// manage_role_menu (role_uuid → menu_uuid) → manage_menu.permissions JSON
// ([{code,desc}]) → flat []string of codes.
package hrmuser

import (
	"context"
	"encoding/json"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/samber/lo"

	"github.com/jzero-io/jzero-admin/server/internal/model/manage_menu"
	"github.com/jzero-io/jzero-admin/server/internal/model/manage_role_menu"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	menutypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/manage/menu"
)

// BillApproveCode is the shared 审批通过 button code seeded on every 人事单据
// approval button (grep seeds: v1:hrm:bill:approve). Only roles granted the
// approval buttons (管理员/HR) carry it, so it doubles as the "may see all
// documents" capability marker (spec 用户故事 4).
const BillApproveCode = "v1:hrm:bill:approve"

// CodesForRoles resolves the union of permission codes granted to the given
// roles. Unknown roles / no grants → empty slice. No DB error path returns
// codes silently truncated: resolution failures are surfaced to the caller.
func CodesForRoles(ctx context.Context, svcCtx *svc.ServiceContext, roleUuids []string) ([]string, error) {
	if len(roleUuids) == 0 {
		return nil, nil
	}
	roleMenus, err := svcCtx.Model.ManageRoleMenu.FindByCondition(ctx, nil,
		condition.NewChain().In(manage_role_menu.RoleUuid, roleUuids).Build()...)
	if err != nil {
		return nil, err
	}
	menuUuids := lo.Uniq(lo.Map(roleMenus, func(rm *manage_role_menu.ManageRoleMenu, _ int) string {
		return rm.MenuUuid
	}))
	if len(menuUuids) == 0 {
		return nil, nil
	}
	menus, err := svcCtx.Model.ManageMenu.FindByCondition(ctx, nil,
		condition.NewChain().In(manage_menu.Uuid, menuUuids).Build()...)
	if err != nil {
		return nil, err
	}
	codes := make([]string, 0, len(menus))
	for _, m := range menus {
		var permissions []menutypes.Permission
		if err := json.Unmarshal([]byte(m.Permissions), &permissions); err != nil {
			continue // corrupt JSON row → skip, keep other menus' codes
		}
		for _, p := range permissions {
			if p.Code != "" {
				codes = append(codes, p.Code)
			}
		}
	}
	return lo.Uniq(codes), nil
}

// RolesHoldCode reports whether any of the roles has been granted a menu whose
// permissions include code (CodesForRoles subset check, no slice build).
func RolesHoldCode(ctx context.Context, svcCtx *svc.ServiceContext, roleUuids []string, code string) (bool, error) {
	codes, err := CodesForRoles(ctx, svcCtx, roleUuids)
	if err != nil {
		return false, err
	}
	for _, c := range codes {
		if c == code {
			return true, nil
		}
	}
	return false, nil
}
