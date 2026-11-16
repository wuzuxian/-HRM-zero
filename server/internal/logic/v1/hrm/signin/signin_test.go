package signin

import (
	"context"
	"strings"
	"testing"

	coresvc "github.com/jzero-io/jzero-admin/core-engine/svc"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_plan"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/signin"
	sessiontypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/training_session"
	sessionlogic "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/training_session"

	"github.com/google/uuid"
)

// Seams under test (HRM-29 acceptance criteria):
//  1. 课堂动态码轮换: refresh 生成新 token（TTL 内有效），旧 token 立即失效（单活跃 token）。
//  2. 扫码签到 DoSignin: token 有效 + 名册成员 → 自签成功（方式=1/操作人=本人）；
//     非名册成员拒绝；已签幂等（不覆盖首次时间）；token 过期/轮换后拒绝。
//  3. 管理员补签 SigninForParticipant: 方式=2/操作人=当前管理员/原因落库；未到名单行可补；
//     幂等（已签不覆盖）。
//  4. 统计口径: Expected=名册数 / Signed=sign_in_time 非空 / Absent=差值。
//  5. 明细与导出: 明细含 姓名/工号/部门/签到时间/方式/操作人/原因；导出 xlsx round-trip。

func testSvcCtx(t *testing.T, env *svctest.Env) *svc.ServiceContext {
	t.Helper()
	return &svc.ServiceContext{ServiceContext: testSvcCore(env), Model: env.Model}
}

func testSvcCore(env *svctest.Env) *coresvc.ServiceContext {
	return &coresvc.ServiceContext{SqlxConn: env.SqlConn}
}

// seedPublishedPlan creates a published training plan (session prerequisite).
func seedPublishedPlan(t *testing.T, env *svctest.Env, published int64) string {
	t.Helper()
	u := uuid.New().String()
	err := hrm_training_plan.NewHrmTrainingPlanModel(env.SqlConn).InsertV2(context.Background(), nil,
		&hrm_training_plan.HrmTrainingPlan{
			Uuid:      u,
			BillNo:    "HR216-2026010100009",
			Status:    "approved",
			PlanName:  "测试培训计划",
			PlanYear:  "2026",
			PlanType:  "special",
			StartDate: "2026-01-01",
			EndDate:   "2026-12-31",
			Published: published,
		})
	if err != nil {
		t.Fatalf("seed plan: %v", err)
	}
	return u
}

// seedUser creates a manage_user row (signin participant).
func seedUser(t *testing.T, env *svctest.Env, account string) string {
	t.Helper()
	u := uuid.New().String()
	fixture := manage_user_fixture(account, u)
	err := env.Model.ManageUser.InsertV2(context.Background(), nil, &fixture)
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return u
}

// addSessionWithRoster creates a session via the real Add + SetSessionParticipants logic
// (roster = 应到名单) and returns (sessionUuid, participantUuids by user order).
func addSessionWithRoster(t *testing.T, svcCtx *svc.ServiceContext, userUuids []string) (string, []string) {
	t.Helper()
	planUuid := seedPublishedPlanT(t, svcCtx)
	if _, err := sessionlogic.NewAdd(context.Background(), svcCtx, nil).Add(&sessiontypes.AddRequest{
		SessionName: "签到测试班次",
		PlanUuid:    planUuid,
		StartTime:   "2026-11-20 09:00:00",
		EndTime:     "2026-11-21 17:00:00",
	}); err != nil {
		t.Fatalf("Add session: %v", err)
	}
	list, err := sessionlogic.NewList(context.Background(), svcCtx, nil).List(&sessiontypes.ListRequest{
		PageRequest: sessiontypes.PageRequest{Current: 1, Size: 10},
		PlanUuid:    planUuid,
	})
	if err != nil || len(list.Records) != 1 {
		t.Fatalf("List after Add: err=%v records=%d", err, len(list.Records))
	}
	sessionUuid := list.Records[0].Uuid
	if _, err = sessionlogic.NewSetSessionParticipants(context.Background(), svcCtx, nil).SetSessionParticipants(
		&sessiontypes.SetSessionParticipantsRequest{SessionUuid: sessionUuid, UserUuids: userUuids}); err != nil {
		t.Fatalf("SetSessionParticipants: %v", err)
	}

	details, err := NewGetSigninDetailList(context.Background(), svcCtx, nil).GetSigninDetailList(
		&types.GetSigninDetailListRequest{SessionUuid: sessionUuid})
	if err != nil {
		t.Fatalf("GetSigninDetailList: %v", err)
	}
	byUser := make(map[string]string, len(details.Records))
	for _, rec := range details.Records {
		byUser[rec.UserUuid] = rec.Uuid
	}
	pids := make([]string, 0, len(userUuids))
	for _, u := range userUuids {
		pid, ok := byUser[u]
		if !ok {
			t.Fatalf("roster missing user %s", u)
		}
		pids = append(pids, pid)
	}
	return sessionUuid, pids
}

func userCtx(userUuid, username string) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "uuid", userUuid)
	ctx = context.WithValue(ctx, "username", username)
	ctx = context.WithValue(ctx, "role_uuids", []any{})
	return ctx
}

func adminCtx() context.Context {
	// 1c2d3e4f... = 迁移 1 种子的 admin（超级管理员）user_uuid，操作人快照按它回查 nickname
	return userCtx("1c2d3e4f-5a6b-7c8d-9e0f-1a2b3c4d5e6f", "admin")
}

// TestRefreshSigninTokenRotation 课堂动态码：刷新轮换（新 token 活跃，旧 token 立即失效）。
func TestRefreshSigninTokenRotation(t *testing.T) {
	env := svctest.New(t)
	svcCtx := testSvcCtx(t, env)

	u1 := seedUser(t, env, "zhangsan")
	sessionUuid, _ := addSessionWithRoster(t, svcCtx, []string{u1})

	// token 1
	r1, err := NewRefreshSigninToken(adminCtx(), svcCtx, nil).RefreshSigninToken(&types.RefreshSigninTokenRequest{SessionUuid: sessionUuid})
	if err != nil {
		t.Fatalf("refresh #1: %v", err)
	}
	if r1.Token == "" || len(r1.Token) < 16 {
		t.Fatalf("token #1 too short/empty: %q", r1.Token)
	}
	if !strings.HasPrefix(r1.SigninUrl, "/hrm/training/signin/scan?token=") {
		t.Fatalf("signinUrl %q missing scan path prefix", r1.SigninUrl)
	}
	if r1.ExpireTime == "" {
		t.Fatal("expireTime empty")
	}

	// token 2 (rotation): old token must die
	r2, err := NewRefreshSigninToken(adminCtx(), svcCtx, nil).RefreshSigninToken(&types.RefreshSigninTokenRequest{SessionUuid: sessionUuid})
	if err != nil {
		t.Fatalf("refresh #2: %v", err)
	}
	if r2.Token == r1.Token {
		t.Fatal("rotation returned the same token")
	}
	for _, tok := range []string{r1.Token, r2.Token} {
		row, ferr := FindActiveSigninToken(context.Background(), svcCtx.Model.HrmTrainingSigninToken, tok, timeNow())
		if ferr != nil {
			t.Fatalf("FindActiveSigninToken: %v", ferr)
		}
		if tok == r1.Token {
			if row != nil {
				t.Fatal("old token still active after rotation")
			}
		} else if row == nil {
			t.Fatal("new token not active")
		}
	}

	// finished session refuses refresh
	if _, err = sessionlogic.NewFinish(context.Background(), svcCtx, nil).Finish(&sessiontypes.FinishRequest{Uuid: sessionUuid}); err != nil {
		t.Fatalf("Finish: %v", err)
	}
	if _, err = NewRefreshSigninToken(adminCtx(), svcCtx, nil).RefreshSigninToken(&types.RefreshSigninTokenRequest{SessionUuid: sessionUuid}); err == nil {
		t.Fatal("refresh on finished session should fail")
	}
}

// TestDoSigninHappyAndIdempotent 扫码签到：成功自签 + 重复扫码幂等（首次时间保留）。
func TestDoSigninHappyAndIdempotent(t *testing.T) {
	env := svctest.New(t)
	svcCtx := testSvcCtx(t, env)

	u1 := seedUser(t, env, "zhangsan")
	sessionUuid, _ := addSessionWithRoster(t, svcCtx, []string{u1})

	r1, err := NewRefreshSigninToken(adminCtx(), svcCtx, nil).RefreshSigninToken(&types.RefreshSigninTokenRequest{SessionUuid: sessionUuid})
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}

	// first scan: signed in, method=1 self
	resp1, err := NewDoSignin(userCtx(u1, "zhangsan"), svcCtx, nil).DoSignin(&types.DoSigninRequest{Token: r1.Token})
	if err != nil {
		t.Fatalf("DoSignin #1: %v", err)
	}
	if resp1.SessionUuid != sessionUuid || resp1.SignInTime == "" {
		t.Fatalf("DoSignin #1 resp = %+v", resp1)
	}

	details, err := NewGetSigninDetailList(context.Background(), svcCtx, nil).GetSigninDetailList(&types.GetSigninDetailListRequest{SessionUuid: sessionUuid})
	if err != nil {
		t.Fatalf("detail list: %v", err)
	}
	if len(details.Records) != 1 {
		t.Fatalf("detail records = %d", len(details.Records))
	}
	rec := details.Records[0]
	if rec.SignInTime == "" || rec.SignInMethod != SignInMethodSelf {
		t.Fatalf("after self signin: method=%q time=%q", rec.SignInMethod, rec.SignInTime)
	}
	if rec.SignInBy != u1 || rec.SignInByName != "zhangsan姓名" {
		t.Fatalf("self signin operator = %q / %q", rec.SignInBy, rec.SignInByName)
	}
	firstTime := rec.SignInTime

	// second scan: idempotent — keeps first time
	resp2, err := NewDoSignin(userCtx(u1, "zhangsan"), svcCtx, nil).DoSignin(&types.DoSigninRequest{Token: r1.Token})
	if err != nil {
		t.Fatalf("DoSignin #2 (idempotent): %v", err)
	}
	if resp2.SignInTime != firstTime {
		t.Fatalf("idempotent DoSignin returned %q, want original %q", resp2.SignInTime, firstTime)
	}

	// stats: 1 expected / 1 signed / 0 absent
	stats, err := NewGetSigninStats(context.Background(), svcCtx, nil).GetSigninStats(&types.GetSigninStatsRequest{SessionUuid: sessionUuid})
	if err != nil {
		t.Fatalf("stats: %v", err)
	}
	if stats.Expected != 1 || stats.Signed != 1 || stats.Absent != 0 {
		t.Fatalf("stats = %+v, want expected=1 signed=1 absent=0", stats)
	}
}

// TestDoSigninGuards 扫码签到拒绝路径：非名册成员 / 无效 token / 轮换后旧码 / 过期码 / 未登录。
func TestDoSigninGuards(t *testing.T) {
	env := svctest.New(t)
	svcCtx := testSvcCtx(t, env)

	u1 := seedUser(t, env, "zhangsan")
	u2 := seedUser(t, env, "lisi")
	sessionUuid, _ := addSessionWithRoster(t, svcCtx, []string{u1})

	r1, err := NewRefreshSigninToken(adminCtx(), svcCtx, nil).RefreshSigninToken(&types.RefreshSigninTokenRequest{SessionUuid: sessionUuid})
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}

	// non-roster user rejected
	if _, err = NewDoSignin(userCtx(u2, "lisi"), svcCtx, nil).DoSignin(&types.DoSigninRequest{Token: r1.Token}); err == nil {
		t.Fatal("non-roster member should be rejected")
	}

	// unauthenticated rejected
	if _, err = NewDoSignin(context.Background(), svcCtx, nil).DoSignin(&types.DoSigninRequest{Token: r1.Token}); err == nil {
		t.Fatal("unauthenticated DoSignin should be rejected")
	}

	// invalid token rejected
	if _, err = NewDoSignin(userCtx(u1, "zhangsan"), svcCtx, nil).DoSignin(&types.DoSigninRequest{Token: "no-such-token"}); err == nil {
		t.Fatal("invalid token should be rejected")
	}

	// rotation kills old token
	if _, err = NewRefreshSigninToken(adminCtx(), svcCtx, nil).RefreshSigninToken(&types.RefreshSigninTokenRequest{SessionUuid: sessionUuid}); err != nil {
		t.Fatalf("refresh #2: %v", err)
	}
	if _, err = NewDoSignin(userCtx(u1, "zhangsan"), svcCtx, nil).DoSignin(&types.DoSigninRequest{Token: r1.Token}); err == nil {
		t.Fatal("rotated-out (old) token should be rejected")
	}

	// expired token rejected
	expired := seedTokenRow(t, env, sessionUuid, "expired-token", timeNow().Add(-1*signinTokenTTL))
	if expired == "" {
		t.Fatal("seed expired token failed")
	}
	if _, err = NewDoSignin(userCtx(u1, "zhangsan"), svcCtx, nil).DoSignin(&types.DoSigninRequest{Token: "expired-token"}); err == nil {
		t.Fatal("expired token should be rejected")
	}

	// finished session rejects signin
	if _, err = sessionlogic.NewFinish(context.Background(), svcCtx, nil).Finish(&sessiontypes.FinishRequest{Uuid: sessionUuid}); err != nil {
		t.Fatalf("Finish: %v", err)
	}
	// fresh token on finished session is refused by refresh; seed one directly
	_ = seedTokenRow(t, env, sessionUuid, "finished-token", timeNow().Add(signinTokenTTL))
	if _, err = NewDoSignin(userCtx(u1, "zhangsan"), svcCtx, nil).DoSignin(&types.DoSigninRequest{Token: "finished-token"}); err == nil {
		t.Fatal("signin on finished session should be rejected")
	}
}

// TestSigninForParticipantRecord 管理员补签：方式=2/操作人=当前管理员/原因落库；幂等不覆盖。
func TestSigninForParticipantRecord(t *testing.T) {
	env := svctest.New(t)
	svcCtx := testSvcCtx(t, env)

	u1 := seedUser(t, env, "zhangsan")
	u2 := seedUser(t, env, "lisi")
	sessionUuid, pids := addSessionWithRoster(t, svcCtx, []string{u1, u2})

	// 补签 u1 (absent row), reason recorded
	if _, err := NewSigninForParticipant(adminCtx(), svcCtx, nil).SigninForParticipant(
		&types.SigninForParticipantRequest{Uuid: pids[0], Reason: "迟到 10 分钟，课后补签"}); err != nil {
		t.Fatalf("SigninForParticipant: %v", err)
	}

	details, err := NewGetSigninDetailList(context.Background(), svcCtx, nil).GetSigninDetailList(&types.GetSigninDetailListRequest{SessionUuid: sessionUuid})
	if err != nil {
		t.Fatalf("detail list: %v", err)
	}
	rec := details.Records[0]
	if rec.SignInMethod != SignInMethodFor || rec.SignInBy != "1c2d3e4f-5a6b-7c8d-9e0f-1a2b3c4d5e6f" || rec.SignInByName != "超级管理员" {
		t.Fatalf("补签记录 method=%q by=%q byName=%q", rec.SignInMethod, rec.SignInBy, rec.SignInByName)
	}
	if rec.SignInReason != "迟到 10 分钟，课后补签" {
		t.Fatalf("补签原因 = %q", rec.SignInReason)
	}

	// repeat 补签: idempotent, keeps first record
	if _, err = NewSigninForParticipant(userCtx("another-admin", "admin2"), svcCtx, nil).SigninForParticipant(
		&types.SigninForParticipantRequest{Uuid: pids[0], Reason: "第二次补签"}); err != nil {
		t.Fatalf("repeat SigninForParticipant: %v", err)
	}
	details, err = NewGetSigninDetailList(context.Background(), svcCtx, nil).GetSigninDetailList(&types.GetSigninDetailListRequest{SessionUuid: sessionUuid})
	if err != nil {
		t.Fatalf("detail list #2: %v", err)
	}
	rec = details.Records[0]
	if rec.SignInByName != "超级管理员" || rec.SignInReason != "迟到 10 分钟，课后补签" {
		t.Fatalf("idempotent 补签 overwrote first record: %+v", rec)
	}

	// stats: 2 expected / 1 signed / 1 absent (u2 still absent)
	stats, err := NewGetSigninStats(context.Background(), svcCtx, nil).GetSigninStats(&types.GetSigninStatsRequest{SessionUuid: sessionUuid})
	if err != nil {
		t.Fatalf("stats: %v", err)
	}
	if stats.Expected != 2 || stats.Signed != 1 || stats.Absent != 1 {
		t.Fatalf("stats = %+v, want expected=2 signed=1 absent=1", stats)
	}

	// unknown participant row rejected
	if _, err = NewSigninForParticipant(adminCtx(), svcCtx, nil).SigninForParticipant(
		&types.SigninForParticipantRequest{Uuid: "no-such-row"}); err == nil {
		t.Fatal("unknown participant row should be rejected")
	}
}

// TestSigninDetailEnrichment 明细口径：姓名/工号/部门来自 user+employee 档案快照。
func TestSigninDetailEnrichment(t *testing.T) {
	env := svctest.New(t)
	svcCtx := testSvcCtx(t, env)

	u1 := seedUser(t, env, "zhangsan")
	ensureOrg(t, env)
	seedEmployee(t, env, u1, "E1001", "张三", orgUuid)
	sessionUuid, _ := addSessionWithRoster(t, svcCtx, []string{u1})

	details, err := NewGetSigninDetailList(context.Background(), svcCtx, nil).GetSigninDetailList(&types.GetSigninDetailListRequest{SessionUuid: sessionUuid})
	if err != nil {
		t.Fatalf("detail list: %v", err)
	}
	rec := details.Records[0]
	if rec.NickName != "zhangsan姓名" || rec.Username != "zhangsan" {
		t.Fatalf("detail user fields = %+v", rec)
	}
	if rec.EmployeeNo != "E1001" {
		t.Fatalf("employeeNo = %q, want E1001", rec.EmployeeNo)
	}
	if rec.OrgName != "测试部门" {
		t.Fatalf("orgName = %q, want 测试部门", rec.OrgName)
	}
}

// TestExportSigninDetailRoundTrip 导出 round-trip：xlsx 解析回读与明细一致（含 补签原因 列）。
func TestExportSigninDetailRoundTrip(t *testing.T) {
	env := svctest.New(t)
	svcCtx := testSvcCtx(t, env)

	u1 := seedUser(t, env, "zhangsan")
	u2 := seedUser(t, env, "lisi")
	sessionUuid, pids := addSessionWithRoster(t, svcCtx, []string{u1, u2})

	// one self signin + one 补签
	if _, err := NewDoSignin(userCtx(u1, "zhangsan"), svcCtx, nil).DoSignin(&types.DoSigninRequest{Token: mustRefresh(t, svcCtx, sessionUuid)}); err != nil {
		t.Fatalf("DoSignin: %v", err)
	}
	if _, err := NewSigninForParticipant(adminCtx(), svcCtx, nil).SigninForParticipant(
		&types.SigninForParticipantRequest{Uuid: pids[1], Reason: "请假后补签"}); err != nil {
		t.Fatalf("补签: %v", err)
	}

	exp, err := NewExportSigninDetail(context.Background(), svcCtx, nil).ExportSigninDetail(&types.ExportSigninDetailRequest{SessionUuid: sessionUuid})
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	if !strings.HasSuffix(exp.FileName, ".xlsx") || exp.Content == "" {
		t.Fatalf("export resp fileName=%q contentLen=%d", exp.FileName, len(exp.Content))
	}
	verifyExportedXlsx(t, exp.Content, []map[string]string{
		{"姓名": "zhangsan姓名", "登录账号": "zhangsan", "签到方式": "扫码签到"},
		{"姓名": "lisi姓名", "签到方式": "管理员补签", "补签原因": "请假后补签"},
	})
}
