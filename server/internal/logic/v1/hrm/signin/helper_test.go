package signin

import (
	"bytes"
	"context"
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_org"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_plan"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_signin_token"
	"github.com/jzero-io/jzero-admin/server/internal/model/manage_user"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/signin"
)

// testOnly helpers split out of signin_test.go to keep seams at top.

func timeNow() time.Time { return time.Now() }

func seedTokenRow(t *testing.T, env *svctest.Env, sessionUuid, token string, expire time.Time) string {
	t.Helper()
	u := uuid.New().String()
	err := hrm_training_signin_token.NewHrmTrainingSigninTokenModel(env.SqlConn).InsertV2(context.Background(), nil,
		&hrm_training_signin_token.HrmTrainingSigninToken{
			Uuid:        u,
			SessionUuid: sessionUuid,
			Token:       token,
			ExpireTime:  expire,
			CreateBy:    "admin-user-uuid",
		})
	if err != nil {
		t.Fatalf("seed token row: %v", err)
	}
	return u
}

func mustRefresh(t *testing.T, svcCtx *svc.ServiceContext, sessionUuid string) string {
	t.Helper()
	req := &signinRefreshReq{SessionUuid: sessionUuid}
	resp, err := NewRefreshSigninToken(adminCtx(), svcCtx, nil).RefreshSigninToken(req.toTypes())
	if err != nil {
		t.Fatalf("mustRefresh: %v", err)
	}
	return resp.Token
}

type signinRefreshReq struct{ SessionUuid string }

func (r *signinRefreshReq) toTypes() *types.RefreshSigninTokenRequest {
	return &types.RefreshSigninTokenRequest{SessionUuid: r.SessionUuid}
}

// seedEmployee links a manage_user to an hrm_employee row (工号/部门 snapshot source).
func seedEmployee(t *testing.T, env *svctest.Env, userUuid, employeeNo, name, orgUuid string) string {
	t.Helper()
	u := uuid.New().String()
	err := env.Model.HrmEmployee.InsertV2(context.Background(), nil, &hrm_employee.HrmEmployee{
		Uuid:       u,
		EmployeeNo: employeeNo,
		Name:       name,
		OrgUuid:    orgUuid,
		UserUuid:   userUuid,
	})
	if err != nil {
		t.Fatalf("seed employee: %v", err)
	}
	return u
}

var orgUuid = func() string {
	// deterministic demo org uuid; seeded lazily by ensureOrg in each test env
	return "org-test-0000-0000-0000-000000000001"
}()

func ensureOrg(t *testing.T, env *svctest.Env) {
	t.Helper()
	err := env.Model.HrmOrg.InsertV2(context.Background(), nil, &hrm_org.HrmOrg{
		Uuid:    orgUuid,
		OrgName: "测试部门",
		Status:  "1",
	})
	if err != nil && !strings.Contains(err.Error(), "UNIQUE") {
		t.Fatalf("seed org: %v", err)
	}
}

// seedPublishedPlanT is the plan fixture shared by addSessionWithRoster.
func seedPublishedPlanT(t *testing.T, svcCtx *svc.ServiceContext) string {
	t.Helper()
	u := uuid.New().String()
	err := hrm_training_plan.NewHrmTrainingPlanModel(svcCtx.SqlxConn).InsertV2(context.Background(), nil,
		&hrm_training_plan.HrmTrainingPlan{
			Uuid:      u,
			BillNo:    "HR216-" + u[:13],
			Status:    "approved",
			PlanName:  "测试培训计划",
			PlanYear:  "2026",
			PlanType:  "special",
			StartDate: "2026-01-01",
			EndDate:   "2026-12-31",
			Published: 1,
		})
	if err != nil {
		t.Fatalf("seed plan: %v", err)
	}
	return u
}

func manage_user_fixture(account, uuid string) manage_user.ManageUser {
	return manage_user.ManageUser{
		Uuid:     uuid,
		Username: account,
		Nickname: account + "姓名",
		Password: "x",
		Gender:   "1",
		Status:   "1",
	}
}

// verifyExportedXlsx decodes the base64 xlsx and asserts data rows by name →
// (header → expected cell) pairs (round-trip against the export contract).
func verifyExportedXlsx(t *testing.T, b64 string, wantRows []map[string]string) {
	t.Helper()
	raw, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		t.Fatalf("decode xlsx: %v", err)
	}
	f, err := excelize.OpenReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("open xlsx: %v", err)
	}
	defer f.Close()
	sheet := f.GetSheetName(0)
	rows, err := f.GetRows(sheet)
	if err != nil {
		t.Fatalf("read xlsx rows: %v", err)
	}
	if len(rows) < len(wantRows)+1 {
		t.Fatalf("xlsx rows = %d, want >= %d", len(rows), len(wantRows)+1)
	}
	header := rows[0]
	if len(header) < 8 || header[0] != "姓名" || header[4] != "签到时间" || header[5] != "签到方式" || header[7] != "补签原因" {
		t.Fatalf("xlsx header = %v", header)
	}
	found := 0
	matched := make(map[string]bool, len(wantRows))
	for _, row := range rows[1:] {
		if len(row) == 0 {
			continue
		}
		if cells, ok := wantNameRows(wantRows)[row[0]]; ok && !matched[row[0]] {
			for hdr, want := range cells {
				col := columnOf(header, hdr)
				if rowCell(row, col) != want {
					t.Fatalf("xlsx row %q: %s = %q, want %q", row[0], hdr, rowCell(row, col), want)
				}
			}
			matched[row[0]] = true
			found++
		}
	}
	if found != len(wantRows) {
		t.Fatalf("matched %d of %d expected rows", found, len(wantRows))
	}
}

// wantNameRows converts []map["姓名"→name + header→cell] into name→cells.
func wantNameRows(rows []map[string]string) map[string]map[string]string {
	out := make(map[string]map[string]string, len(rows))
	for _, r := range rows {
		name := r["姓名"]
		cells := make(map[string]string, len(r)-1)
		for k, v := range r {
			if k != "姓名" {
				cells[k] = v
			}
		}
		out[name] = cells
	}
	return out
}

func columnOf(header []string, name string) int {
	for i, h := range header {
		if h == name {
			return i
		}
	}
	return -1
}

func rowCell(row []string, i int) string {
	if i < len(row) {
		return row[i]
	}
	return ""
}
