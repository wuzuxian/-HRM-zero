package handler

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/jzero-io/jzero-admin/server/internal/svctest"
)

// 权限码三处一致性守卫。
//
// 菜单 permissions JSON 有两个消费者：casbin（role set_menus 运行时据此重建
// casbin_rule）与前端按钮授权；而 Authx 中间件（core-engine）用本包 route2code
// 的生成映射查 casbin。两个来源各写各的，一旦漂移——码名不一致或漏授——
// casbin 全表无此策略，端点对任何角色裸 403（空 body）。
// 纯后端逻辑单测不经过 Authx 发现不了，因此在这里在 DB 层钉死不变式。
//
// bug 实例（全部是本测试拦截后不再复发的一类）：
//   - a232d16 考核任务 detail/review/calibrate
//   - 迁移 60 doSignin（无菜单授予）
//   - 迁移 61 perf_result list/sign/submit
//   - 迁移 62 perf_period:advance / training_session:finish / maintain_year
//
// 豁免见 exemptFromMenuGrant：确属"尚未接线的产品端点"，接线时随菜单种子
// 一并授予并从清单移除；新增豁免必须带理由注释。

// exemptFromMenuGrant 存在于 route2code 但允许不出现在任何菜单 permissions 的码。
// （v1:auth:* / v1:route:* 等非 jwt 网关码不在本守卫范围内，无需豁免。）
var exemptFromMenuGrant = map[string]string{
	// 员工自助工资条端点：前端尚无调用方（typings/service 已生成，页面未建），
	// 授予载体（"我的工资条"菜单）待产品接线时一并落。
	"v1:hrm:salary_payslip:mineList": "员工自助端点，前端未接线",
	"v1:hrm:salary_payslip:read":     "员工自助端点，前端未接线",
}

type menuRow struct {
	Permissions string `db:"permissions"`
}

type permissionItem struct {
	Code string `json:"code"`
}

func loadMenuPermissionCodes(t *testing.T, env *svctest.Env) map[string]int {
	t.Helper()

	var rows []menuRow
	if err := env.SqlConn.QueryRowsCtx(context.Background(), &rows,
		"SELECT permissions FROM manage_menu"); err != nil {
		t.Fatalf("query manage_menu.permissions: %v", err)
	}

	codes := make(map[string]int)
	for _, row := range rows {
		if strings.TrimSpace(row.Permissions) == "" {
			continue
		}
		var items []permissionItem
		if err := json.Unmarshal([]byte(row.Permissions), &items); err != nil {
			t.Fatalf("unmarshal permissions %q: %v", row.Permissions, err)
		}
		for _, item := range items {
			if item.Code != "" {
				codes[item.Code]++
			}
		}
	}
	return codes
}

func isHRMCode(code string) bool {
	return strings.HasPrefix(code, "v1:hrm:")
}

// TestRoute2CodeCoveredByMenuPermissions 正向：Authx 守卫的每个 hrm 端点码
// 必须能通过某个菜单授予（存在于至少一个菜单的 permissions JSON），
// 否则 casbin 永远无此策略 → 端点对全员 403。
func TestRoute2CodeCoveredByMenuPermissions(t *testing.T) {
	env := svctest.New(t)
	menuCodes := loadMenuPermissionCodes(t, env)

	for _, code := range RoutesCodesMap {
		if !isHRMCode(code) {
			continue
		}
		if _, ok := exemptFromMenuGrant[code]; ok {
			continue
		}
		if menuCodes[code] == 0 {
			t.Errorf("route2code 码 %q 不在任何菜单 permissions 中 → Authx 必 403；"+
				"请把它补进承载该功能的菜单 permissions 种子（新迁移，三方言同步），"+
				"或确认属产品未接线端点后加入 exemptFromMenuGrant（带理由）", code)
		}
	}
}

// TestMenuPermissionCodeIsRealRouteCode 反向：菜单 permissions 里的每个 hrm 码
// 必须是 route2code 生成映射中的真实码，否则是死授予——casbin 永远不会被
// 该码命中（a232d16 即此类：种子写了生成器不会产生的码名）。
func TestMenuPermissionCodeIsRealRouteCode(t *testing.T) {
	env := svctest.New(t)
	menuCodes := loadMenuPermissionCodes(t, env)

	routeCodes := make(map[string]struct{}, len(RoutesCodesMap))
	for _, code := range RoutesCodesMap {
		routeCodes[code] = struct{}{}
	}

	for code := range menuCodes {
		if !isHRMCode(code) {
			continue
		}
		if _, ok := routeCodes[code]; !ok {
			t.Errorf("菜单 permissions 码 %q 不是 route2code 的任何生成码（死授予）；"+
				"请对齐 route2code.go 的实际映射（新迁移，三方言同步）", code)
		}
	}
}
