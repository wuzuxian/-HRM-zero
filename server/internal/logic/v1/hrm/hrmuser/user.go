// Package hrmuser resolves the current JWT user and role facts shared by HRM
// document list logics (spec 用户故事 4: 列表默认按「我的单据」过滤，管理员可看全部).
package hrmuser

import (
	"context"

	"github.com/jzero-io/jzero-admin/core-engine/helper/auth"
)

// SuperAdminRoleUuid is the seeded 超级管理员 (R_SUPER) role uuid — the role
// that is allowed to see all documents instead of only its own.
const SuperAdminRoleUuid = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"

// User is the current JWT user as seen by HRM logic.
type User struct {
	Uuid      string
	Username  string
	RoleUuids []string
}

// Current returns the JWT user; a zero-value User (with a logged error) when
// the context carries no auth info — e.g. logic tests.
func Current(ctx context.Context) User {
	info, err := auth.Info(ctx)
	if err != nil {
		return User{}
	}
	return User{
		Uuid:      info.Uuid,
		Username:  info.Username,
		RoleUuids: info.RoleUuids,
	}
}

// IsSuperAdmin reports whether the user carries the 超级管理员 role.
func (u User) IsSuperAdmin() bool {
	for _, r := range u.RoleUuids {
		if r == SuperAdminRoleUuid {
			return true
		}
	}
	return false
}
