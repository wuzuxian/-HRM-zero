package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/jzero-io/jzero-admin/core-engine/helper/auth"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/manage_menu"
	"github.com/jzero-io/jzero-admin/server/internal/model/manage_role"
	"github.com/jzero-io/jzero-admin/server/internal/model/manage_role_menu"
	"github.com/jzero-io/jzero-admin/server/internal/model/manage_user_role"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/auth"
)

type GetUserInfo struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewGetUserInfo(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *GetUserInfo {
	return &GetUserInfo{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx, r: r,
	}
}

func (l *GetUserInfo) GetUserInfo(req *types.GetUserInfoRequest) (resp *types.GetUserInfoResponse, err error) {
	info, err := auth.Info(l.ctx)
	if err != nil {
		return nil, err
	}

	user, err := l.svcCtx.Model.ManageUser.FindOneByUuid(l.ctx, nil, info.Uuid)
	if err != nil {
		return nil, err
	}

	userRoles, err := l.svcCtx.Model.ManageUserRole.FindByCondition(l.ctx, nil, condition.NewChain().
		Equal(manage_user_role.UserUuid, user.Uuid).
		Build()...)
	if err != nil {
		return nil, err
	}

	var roleUuids []string
	for _, userRole := range userRoles {
		roleUuids = append(roleUuids, userRole.RoleUuid)
	}

	roles, err := l.svcCtx.Model.ManageRole.FindByCondition(l.ctx, nil, condition.NewChain().
		In(manage_role.Uuid, roleUuids).
		Build()...)
	if err != nil {
		return nil, err
	}
	var roleCodes []string
	for _, role := range roles {
		roleCodes = append(roleCodes, role.Code)
	}

	// get role buttons
	roleMenus, err := l.svcCtx.Model.ManageRoleMenu.FindByCondition(l.ctx, nil, condition.NewChain().
		In(manage_role_menu.RoleUuid, roleUuids).
		Build()...)
	if err != nil {
		return nil, err
	}
	var menuUuids []string
	for _, roleMenu := range roleMenus {
		menuUuids = append(menuUuids, roleMenu.MenuUuid)
	}
	menus, err := l.svcCtx.Model.ManageMenu.FindByCondition(l.ctx, nil, condition.NewChain().
		In(manage_menu.Uuid, menuUuids).
		Equal(manage_menu.Status, "1").
		Equal(manage_menu.MenuType, "3").
		Build()...)
	if err != nil {
		return nil, err
	}
	buttons := buttonCodes(menus)

	return &types.GetUserInfoResponse{
		UserUuid: user.Uuid,
		Username: user.Username,
		Roles:    roleCodes,
		Buttons:  buttons,
	}, nil
}

// buttonCodeAPIVersionPrefix is the API version prefix ("v1") stored in the
// seed data for button codes (e.g. "v1:manage:user:add"), mirroring the API
// permission codes used by casbin. The frontend permission check compares bare
// codes ("manage:user:add"), so the prefix is stripped at this boundary; codes
// stored without it pass through unchanged.
const buttonCodeAPIVersionPrefix = "v1:"

// buttonCodes collects the UI button codes from button-type menus.
func buttonCodes(menus []*manage_menu.ManageMenu) []string {
	buttons := make([]string, 0, len(menus))
	for _, menu := range menus {
		buttons = append(buttons, strings.TrimPrefix(menu.ButtonCode, buttonCodeAPIVersionPrefix))
	}
	return buttons
}
