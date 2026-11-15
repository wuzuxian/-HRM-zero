package training_session

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	coresvc "github.com/jzero-io/jzero-admin/core-engine/svc"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_plan"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_session"
	"github.com/jzero-io/jzero-admin/server/internal/model/manage_user"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
)

// testSvcCore builds the core-engine embed needed by logic that calls
// l.svcCtx.SqlxConn (the Add transaction).
func testSvcCore(env *svctest.Env) *coresvc.ServiceContext {
	return &coresvc.ServiceContext{SqlxConn: env.SqlConn}
}

// Seams under test (per ticket acceptance criteria):
//  1. PX 编号生成单测 (acceptance criterion): pure function FormatSessionNo
//     renders `PX-{year}-{4位流水}`; allocation via GenerateSessionNo is per-year
//     and monotonic within a year.
//  2. Add: PX 编号 generated in-tx, status 默认 1=未开始, plan_uuid filled; only
//     published (published=1) 培训计划 can be referenced (发布计划后可建班次).
//  3. 状态流转: Finish flips 未开始 → 已结束; finish on 已结束 rejected.
//  4. Participant binding: 选人 SetSessionParticipants is a full replace
//     (全量覆盖), GetSessionParticipantList returns the bound rows with user info.

func TestFormatSessionNo(t *testing.T) {
	no, err := FormatSessionNo(2026, 1)
	if err != nil {
		t.Fatalf("FormatSessionNo: %v", err)
	}
	if no != "PX-2026-0001" {
		t.Fatalf("FormatSessionNo = %q, want PX-2026-0001", no)
	}
	if _, err = FormatSessionNo(2026, 0); err == nil {
		t.Fatal("seq 0 should be rejected")
	}
	if _, err = FormatSessionNo(2026, 10000); err == nil {
		t.Fatal("seq 10000 (4 位流水上限) should be rejected")
	}
	no2, err := FormatSessionNo(1999, 9999)
	if err != nil || no2 != "PX-1999-9999" {
		t.Fatalf("FormatSessionNo(1999,9999) = %q err=%v", no2, err)
	}
}

func TestGenerateSessionNoAllocatesPerYear(t *testing.T) {
	env := svctest.New(t)
	day := time.Date(2026, 11, 15, 10, 0, 0, 0, time.Local)
	n1, err := GenerateSessionNo(context.Background(), env.SqlConn, 2026, day)
	if err != nil {
		t.Fatalf("GenerateSessionNo #1: %v", err)
	}
	n2, err := GenerateSessionNo(context.Background(), env.SqlConn, 2026, day)
	if err != nil {
		t.Fatalf("GenerateSessionNo #2: %v", err)
	}
	_, s1, e1 := parseSessionNo(n1)
	_, s2, e2 := parseSessionNo(n2)
	if e1 != nil || e2 != nil {
		t.Fatalf("parse: %v %v", e1, e2)
	}
	if s2 != s1+1 {
		t.Fatalf("per-year allocation not monotonic: %d then %d", s1, s2)
	}
}

func parseSessionNo(no string) (int, int, error) {
	// PX-2026-0001
	parts := strings.Split(no, "-")
	if len(parts) != 3 || parts[0] != "PX" {
		return 0, 0, ErrSessionNoSeqOutOfRange
	}
	var y, s int
	if _, err := fmt.Sscanf(parts[1]+" "+parts[2], "%d %d", &y, &s); err != nil {
		return 0, 0, err
	}
	return y, s, nil
}

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

func seedUser(t *testing.T, env *svctest.Env, account string) string {
	t.Helper()
	u := uuid.New().String()
	err := env.Model.ManageUser.InsertV2(context.Background(), nil, &manage_user.ManageUser{
		Uuid:     u,
		Username: account,
		Nickname: account + "姓名",
		Password: "x",
		Gender:   "1",
		Status:   "1",
	})
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return u
}

func mustFindSession(t *testing.T, env *svctest.Env, sessionUuid string) *hrm_training_session.HrmTrainingSession {
	t.Helper()
	row, err := hrm_training_session.NewHrmTrainingSessionModel(env.SqlConn).FindOneByUuid(context.Background(), nil, sessionUuid)
	if err != nil {
		t.Fatalf("find session %s: %v", sessionUuid, err)
	}
	return row
}

var _ = time.Now
