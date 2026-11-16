package perf_assessment

// DB-backed integration test for the production code-resolution path (fix F1):
// handlers never call WithCodes, so resolveActor must self-resolve the user's
// button codes from role→menu permissions (hrmuser.CodesForRoles, same chain
// as custom.InitCasbinRule). This test seeds a role granted
// v1:hrm:perf_assessment:review via manage_menu.permissions + manage_role_menu,
// puts a JWT-like auth payload into ctx (no WithCodes), and proves
// LeaderReview succeeds through resolveActor — the exact production path.

import (
	"context"
	"testing"

	"github.com/jzero-io/jzero-admin/server/internal/model/manage_menu"
	"github.com/jzero-io/jzero-admin/server/internal/model/manage_role_menu"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
)

// authCtx builds a context that core-engine/helper/auth.Info reads as a logged
// in user (keys "uuid" / "username" / "role_uuids" as []any — same shape the
// authx middleware puts into request context).
func authCtx(userUuid string, roleUuids ...string) context.Context {
	roles := make([]any, len(roleUuids))
	for i, r := range roleUuids {
		roles[i] = r
	}
	return context.WithValue(context.WithValue(context.WithValue(context.Background(),
		"uuid", userUuid), "username", "reviewer"), "role_uuids", roles)
}

// seedReviewRoleGrant seeds a manage_menu row whose permissions JSON grants
// code, plus a manage_role_menu row binding role→menu. Menu uuid suffix 99xx (migrations stay below 42xx; test-local uniqueness)
func seedReviewRoleGrant(t *testing.T, env *svctest.Env, roleUuid, menuUuid, rmUuid, code string) {
	t.Helper()
	mustNoErr(t, env.Model.ManageMenu.InsertV2(testCtx, nil, &manage_menu.ManageMenu{
		Uuid:        menuUuid,
		Status:      "1",
		MenuType:    "3",
		MenuName:    "评审提交",
		Permissions: `[{"code":"` + code + `","desc":"上级评审"}]`,
	}))
	mustNoErr(t, env.Model.ManageRoleMenu.InsertV2(testCtx, nil, &manage_role_menu.ManageRoleMenu{
		Uuid:     rmUuid,
		RoleUuid: roleUuid,
		MenuUuid: menuUuid,
	}))
}

// snapshotItemUuids reads the two indicator uuids out of the task snapshot.
func snapshotItemUuids(t *testing.T, svcCtx *svc.ServiceContext, taskUuid string) []string {
	t.Helper()
	task, err := FindAssessmentByUuid(testCtx, svcCtx.SqlxConn, taskUuid)
	mustNoErr(t, err)
	snapshot, err := parseItemsSnapshot(task.ItemsSnapshot)
	mustNoErr(t, err)
	return []string{snapshot[0].IndicatorUuid, snapshot[1].IndicatorUuid}
}

// itemsOf builds two-item score rows (80 each) from the snapshot uuids.
func itemsOf(uuids []string, score float64) []ScoreItem {
	return []ScoreItem{
		{IndicatorUuid: uuids[0], Score: score},
		{IndicatorUuid: uuids[1], Score: score},
	}
}

// TestResolveActorFromRoleGrants walks the real handler-context path: reviewer
// user with NO WithCodes injection, codes resolved from role grants.
func TestResolveActorFromRoleGrants(t *testing.T) {
	svcCtx, env := newSvc(t)
	tpl := seedTemplate(t, env)
	scheme := seedScheme(t, env, tpl, "percent")
	emp := seedEmployee(t, env, "user-emp-self")
	periodUuid := seedOngoingPeriod(t, env, scheme, emp)
	mustGenerate(t, svcCtx, periodUuid)
	task := mustFindTask(t, svcCtx, periodUuid)
	uuids := snapshotItemUuids(t, svcCtx, task.Uuid)

	// Reviewer user bound to a role granted the review code via menu seed.
	seedReviewRoleGrant(t, env, "role-reviewer",
		"aaa10000-1111-4000-8000-000000009901",
		"aaa20000-1111-4000-8000-000000009901", PermAssessmentReview)

	// Drive emp's task plan → self (owner actor injected directly, out of scope here).
	mustNoErr(t, ConfirmPlanForTask(testCtx, svcCtx.SqlxConn, task.Uuid, "", actorFor(emp)))
	mustNoErr(t, SubmitSelfReviewForTask(testCtx, svcCtx, task.Uuid, itemsOf(uuids, 80), "", actorFor(emp)))

	// Production path: JWT-like ctx, resolveActor self-resolves codes.
	ctx := authCtx("user-reviewer", "role-reviewer")
	actor, err := resolveActor(ctx, svcCtx)
	mustNoErr(t, err)
	if !hasCode(actor.Codes, PermAssessmentReview) {
		t.Fatalf("expected resolved codes to contain %s, got %v", PermAssessmentReview, actor.Codes)
	}
	mustNoErr(t, SubmitLeaderReviewForTask(ctx, svcCtx, task.Uuid, itemsOf(uuids, 90), "", actor))

	// Negative control: a role without the review grant stays denied
	// (proves resolution isn't accidentally granting everything).
	seedReviewRoleGrant(t, env, "role-plain",
		"aaa10000-1111-4000-8000-000000009902",
		"aaa20000-1111-4000-8000-000000009902", "v1:hrm:perf_assessment:detail")
	plainActor, err := resolveActor(authCtx("user-plain", "role-plain"), svcCtx)
	mustNoErr(t, err)
	if hasCode(plainActor.Codes, PermAssessmentReview) {
		t.Fatalf("plain role must not carry %s, got %v", PermAssessmentReview, plainActor.Codes)
	}
}
