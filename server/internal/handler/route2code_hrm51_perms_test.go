package handler

import (
	"context"
	"strings"
	"testing"

	"github.com/jzero-io/jzero-admin/server/internal/svctest"
)

// HRM-51 迁移 64 行为断言: 关联/解绑 4 个员工侧端点码 + 用户侧只读反查码
// 必须随迁移落到菜单 permissions (Authx/casbin 的授予载体), 且持有相应页面的
// 角色得到 casbin 存量行。route2code 一致性守卫只查"存在某菜单授予",
// 这里钉死授予载体与存量 casbin 行, 防止迁移漏写 role_menu/permissions。

func TestHRM51MigrationSeedsPermissions(t *testing.T) {
	env := svctest.New(t)

	employeePage := "aaa10000-1111-4000-8000-000000000402"
	userPage := "b3c4d5e6-f7a8-4912-83b4-c5d6e7f8a9b0"
	superUuid := "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"

	var rows []menuRow
	if err := env.SqlConn.QueryRowsCtx(context.Background(), &rows,
		"SELECT permissions FROM manage_menu WHERE uuid = ? OR parent_uuid = ?", employeePage, employeePage); err != nil {
		t.Fatalf("query employee page menus: %v", err)
	}
	joined := make([]string, 0, len(rows))
	for _, row := range rows {
		joined = append(joined, row.Permissions)
	}
	all := strings.Join(joined, ",")

	var userPageRow menuRow
	if err := env.SqlConn.QueryRowCtx(context.Background(), &userPageRow,
		"SELECT permissions FROM manage_menu WHERE uuid = ?", userPage); err != nil {
		t.Fatalf("load user page: %v", err)
	}

	type casbinRow struct {
		V1 string `db:"v1"`
	}
	var casbins []casbinRow
	if err := env.SqlConn.QueryRowsCtx(context.Background(), &casbins,
		"SELECT v1 FROM casbin_rule WHERE p_type = 'p' AND v0 = ?", superUuid); err != nil {
		t.Fatalf("query casbin_rule: %v", err)
	}
	granted := map[string]bool{}
	for _, row := range casbins {
		granted[row.V1] = true
	}

	// 每个新码: 菜单 permissions 有授予载体 + R_SUPER 有 casbin 存量行
	for _, code := range []string{
		"v1:hrm:employee:linkUser",
		"v1:hrm:employee:unlinkUser",
		"v1:hrm:employee:linkableUserList",
		"v1:hrm:employee:createAndLinkUser",
		"v1:manage:user:linkedEmployees",
	} {
		if code != "v1:manage:user:linkedEmployees" && !strings.Contains(all, code) {
			t.Errorf("员工档案页 (含按钮) permissions 缺 %q — 迁移 64 未生效", code)
		}
		if !granted[code] {
			t.Errorf("R_SUPER casbin 缺 %q — 迁移 64 存量行未生效", code)
		}
	}
	if !strings.Contains(userPageRow.Permissions, "v1:manage:user:linkedEmployees") {
		t.Errorf("用户管理页 permissions 缺 v1:manage:user:linkedEmployees — 迁移 64 未生效")
	}
}
