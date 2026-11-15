package training_session

import (
	"context"
	"strings"
	"testing"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_session_participant"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/training_session"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"

	"github.com/jzero-io/jzero/core/stores/condition"
)

// DB-backed seams (svctest against the real migrations): plan→session linkage,
// status 流转, participant binding. Shared helpers live in sessionno_test.go.

func testSvcCtx(t *testing.T, env *svctest.Env) *svc.ServiceContext {
	t.Helper()
	return &svc.ServiceContext{ServiceContext: testSvcCore(env), Model: env.Model}
}

func addSession(t *testing.T, svcCtx *svc.ServiceContext, planUuid string) string {
	t.Helper()
	return addSessionAt(t, svcCtx, planUuid, 0)
}

func addSessionAt(t *testing.T, svcCtx *svc.ServiceContext, planUuid string, expected int) string {
	t.Helper()
	req := &types.AddRequest{
		SessionName: "新员工入职培训",
		PlanUuid:    planUuid,
		Lecturer:    "王老师",
		Location:    "总部一楼培训室",
		Hours:       14,
		StartTime:   "2026-11-20 09:00:00",
		EndTime:     "2026-11-21 17:00:00",
	}
	if _, err := NewAdd(context.Background(), svcCtx, nil).Add(req); err != nil {
		t.Fatalf("Add session: %v", err)
	}
	// list seam filtered by plan (seed data from migration 15 belongs to other plans)
	resp, err := NewList(context.Background(), svcCtx, nil).List(&types.ListRequest{
		PageRequest: types.PageRequest{Current: 1, Size: 10},
		PlanUuid:    planUuid,
	})
	if err != nil {
		t.Fatalf("List after Add: %v", err)
	}
	if len(resp.Records) != expected+1 {
		t.Fatalf("list records = %d, want %d", len(resp.Records), expected+1)
	}
	rec := resp.Records[0]
	if !strings.HasPrefix(rec.SessionNo, "PX-2026-") {
		t.Fatalf("sessionNo %q does not follow PX-年-流水 format", rec.SessionNo)
	}
	if rec.Status != "1" {
		t.Fatalf("new session status = %q, want 1 (未开始)", rec.Status)
	}
	if rec.PlanName != "测试培训计划" {
		t.Fatalf("planName snapshot = %q, want 测试培训计划", rec.PlanName)
	}
	return rec.Uuid
}

func TestAddRequiresPublishedPlan(t *testing.T) {
	env := svctest.New(t)
	svcCtx := testSvcCtx(t, env)

	// 草稿 (published=0) plan: rejected
	draftUuid := seedPublishedPlan(t, env, 0)
	req := &types.AddRequest{SessionName: "班次", PlanUuid: draftUuid, StartTime: "2026-11-20 09:00:00", EndTime: "2026-11-21 17:00:00"}
	if _, err := NewAdd(context.Background(), svcCtx, nil).Add(req); err == nil {
		t.Fatal("add session on unpublished (draft) plan should fail")
	}

	// 已发布 (published=1) plan: allowed, PX generated
	publishedUuid := seedPublishedPlan(t, env, 1)
	addSession(t, svcCtx, publishedUuid)
}

func TestAddValidationAndTimeOrder(t *testing.T) {
	env := svctest.New(t)
	svcCtx := testSvcCtx(t, env)
	planUuid := seedPublishedPlan(t, env, 1)

	cases := []struct {
		name   string
		mutate func(*types.AddRequest)
	}{
		{"missing name", func(r *types.AddRequest) { r.SessionName = "" }},
		{"missing plan", func(r *types.AddRequest) { r.PlanUuid = "" }},
		{"bad plan", func(r *types.AddRequest) { r.PlanUuid = "no-such-plan" }},
		{"end before start", func(r *types.AddRequest) { r.StartTime = "2026-11-21 17:00:00"; r.EndTime = "2026-11-20 09:00:00" }},
		{"bad time format", func(r *types.AddRequest) { r.StartTime = "20261120" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := &types.AddRequest{
				SessionName: "班次",
				PlanUuid:    planUuid,
				StartTime:   "2026-11-20 09:00:00",
				EndTime:     "2026-11-21 17:00:00",
			}
			tc.mutate(req)
			if _, err := NewAdd(context.Background(), svcCtx, nil).Add(req); err == nil {
				t.Fatal("expected validation error, got nil")
			}
		})
	}
}

func TestGenerateSessionNoMonotonicThroughAdd(t *testing.T) {
	env := svctest.New(t)
	svcCtx := testSvcCtx(t, env)
	planUuid := seedPublishedPlan(t, env, 1)

	first := addSessionAt(t, svcCtx, planUuid, 0)
	second := addSessionAt(t, svcCtx, planUuid, 1)

	// same adds → same year, 流水 must increase
	session1 := mustFindSession(t, env, first)
	session2 := mustFindSession(t, env, second)
	y1, s1, e1 := parseSessionNo(session1.SessionNo)
	y2, s2, e2 := parseSessionNo(session2.SessionNo)
	if e1 != nil || e2 != nil {
		t.Fatalf("parse session numbers: %v %v", e1, e2)
	}
	if y1 != y2 {
		t.Fatalf("back-to-back adds must share the year: %d vs %d", y1, y2)
	}
	if s2 != s1+1 {
		t.Fatalf("seq not monotonic: %d then %d", s1, s2)
	}
}

func TestFinishStatusTransition(t *testing.T) {
	env := svctest.New(t)
	svcCtx := testSvcCtx(t, env)
	planUuid := seedPublishedPlan(t, env, 1)
	sessionUuid := addSession(t, svcCtx, planUuid)

	// 未开始 → 已结束 (行操作「结束」)
	if _, err := NewFinish(context.Background(), svcCtx, nil).Finish(&types.FinishRequest{Uuid: sessionUuid}); err != nil {
		t.Fatalf("Finish: %v", err)
	}
	if row := mustFindSession(t, env, sessionUuid); row.Status != "3" {
		t.Fatalf("status after finish = %q, want 3 (已结束)", row.Status)
	}

	// 已结束 is terminal: finish again must be rejected
	if _, err := NewFinish(context.Background(), svcCtx, nil).Finish(&types.FinishRequest{Uuid: sessionUuid}); err == nil {
		t.Fatal("finish on 已结束 session should fail")
	}
}

func TestParticipantBindingFullReplace(t *testing.T) {
	env := svctest.New(t)
	svcCtx := testSvcCtx(t, env)
	planUuid := seedPublishedPlan(t, env, 1)
	sessionUuid := addSession(t, svcCtx, planUuid)

	u1 := seedUser(t, env, "user-a")
	u2 := seedUser(t, env, "user-b")

	// bind two
	if _, err := NewSetSessionParticipants(context.Background(), svcCtx, nil).SetSessionParticipants(
		&types.SetSessionParticipantsRequest{SessionUuid: sessionUuid, UserUuids: []string{u1, u2}}); err != nil {
		t.Fatalf("SetSessionParticipants: %v", err)
	}
	resp, err := NewGetSessionParticipantList(context.Background(), svcCtx, nil).GetSessionParticipantList(
		&types.GetSessionParticipantListRequest{SessionUuid: sessionUuid})
	if err != nil {
		t.Fatalf("GetSessionParticipantList: %v", err)
	}
	if len(resp.Records) != 2 {
		t.Fatalf("participants = %d, want 2", len(resp.Records))
	}

	// 全量覆盖: bind only u1 → u2 dropped
	if _, err := NewSetSessionParticipants(context.Background(), svcCtx, nil).SetSessionParticipants(
		&types.SetSessionParticipantsRequest{SessionUuid: sessionUuid, UserUuids: []string{u1}}); err != nil {
		t.Fatalf("SetSessionParticipants (replace): %v", err)
	}
	resp, err = NewGetSessionParticipantList(context.Background(), svcCtx, nil).GetSessionParticipantList(
		&types.GetSessionParticipantListRequest{SessionUuid: sessionUuid})
	if err != nil {
		t.Fatalf("GetSessionParticipantList (after replace): %v", err)
	}
	if len(resp.Records) != 1 {
		t.Fatalf("participants after replace = %d, want 1 (full replace)", len(resp.Records))
	}
	if resp.Records[0].UserUuid != u1 {
		t.Fatalf("remaining participant = %q, want %q", resp.Records[0].UserUuid, u1)
	}

	// 我的班次 seam: participant rows keyed by (session_uuid, user_uuid) must
	// surface through the shared table (HRM-31 getMySessionList reads them)
	participants, err := env.Model.HrmTrainingSessionParticipant.FindByCondition(context.Background(), nil,
		condition.NewChain().Equal(hrm_training_session_participant.SessionUuid, sessionUuid).Build()...)
	if err != nil {
		t.Fatalf("participant rows: %v", err)
	}
	if len(participants) != 1 || participants[0].UserUuid != u1 {
		t.Fatalf("participant row mismatch: %+v", participants)
	}
}
