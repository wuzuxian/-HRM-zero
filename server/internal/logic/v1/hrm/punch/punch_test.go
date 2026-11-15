package punch

import (
	"context"
	"testing"
	"time"

	"github.com/jzero-io/jzero/core/status"

	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/punch"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
)

// TestPunchClockInIdempotent 打卡幂等: 同日重复上班卡拒绝（10523），下班卡可正常打。
func TestPunchClockInIdempotent(t *testing.T) {
	svcCtx := newSvcCtx(t)
	ctx := context.Background()
	s := seedOrgAndEmployee(t, svcCtx)
	saveOrgConfig(t, svcCtx, s.orgTech, "09:00", "18:00", []int64{1, 2, 3, 4, 5})

	// A: 上班卡（外勤，带地点）
	resp, err := NewPunch(ctx, svcCtx, nil).Punch(&types.PunchRequest{
		Type: PunchTypeClockIn, Location: "客户现场-外勤", EmployeeUuid: s.empZhang,
	})
	if err != nil {
		t.Fatalf("clockIn: %v", err)
	}
	if resp.PunchDate != today() {
		t.Fatalf("punchDate = %q, want %q", resp.PunchDate, today())
	}
	if resp.ClockInTime == "" {
		t.Fatalf("clockInTime empty")
	}

	// B: 同日重复上班卡 → 10523 拒绝
	_, err = NewPunch(ctx, svcCtx, nil).Punch(&types.PunchRequest{
		Type: PunchTypeClockIn, Location: "公司", EmployeeUuid: s.empZhang,
	})
	if err == nil {
		t.Fatalf("duplicate clockIn should fail")
	}
	if sc := status.FromError(err); sc == nil || sc.Code() != errcodes.PunchDuplicateCode {
		t.Fatalf("duplicate clockIn error = %v, want code 10523", err)
	}

	// C: 下班卡正常
	resp, err = NewPunch(ctx, svcCtx, nil).Punch(&types.PunchRequest{
		Type: PunchTypeClockOut, EmployeeUuid: s.empZhang,
	})
	if err != nil {
		t.Fatalf("clockOut: %v", err)
	}
	if resp.ClockOutTime == "" {
		t.Fatalf("clockOutTime empty")
	}
	if resp.SettleStatus == "" {
		t.Fatalf("settleStatus empty after clockOut")
	}

	// D: 重复下班卡 → 拒绝
	_, err = NewPunch(ctx, svcCtx, nil).Punch(&types.PunchRequest{Type: PunchTypeClockOut, EmployeeUuid: s.empZhang})
	if sc := status.FromError(err); sc == nil || sc.Code() != errcodes.PunchDuplicateCode {
		t.Fatalf("duplicate clockOut error = %v, want code 10523", err)
	}

	// E: 参数校验 — 非法 type / 员工不存在
	_, err = NewPunch(ctx, svcCtx, nil).Punch(&types.PunchRequest{Type: "lunch", EmployeeUuid: s.empZhang})
	if sc := status.FromError(err); sc == nil || sc.Code() != errcodes.PunchInvalidCode {
		t.Fatalf("bad type error = %v, want code 10521", err)
	}
	_, err = NewPunch(ctx, svcCtx, nil).Punch(&types.PunchRequest{Type: PunchTypeClockIn, EmployeeUuid: "no-such"})
	if sc := status.FromError(err); sc == nil || sc.Code() != errcodes.PunchEmployeeNotFoundCode {
		t.Fatalf("bad employee error = %v, want code 10522", err)
	}
}

// TestPunchCrossDayOwnership 跨天班下班卡归属上班日: 22:00-06:00 夜班，次日 06:10 打下班卡
// 落在上班日的行上（punch_date = 上班日）。
func TestPunchCrossDayOwnership(t *testing.T) {
	svcCtx := newSvcCtx(t)
	ctx := context.Background()
	s := seedOrgAndEmployee(t, svcCtx)
	addGroup(t, svcCtx, "夜班组", "22:00", "06:00", []int64{1, 2, 3, 4, 5, 6, 7}, []string{s.empZhang})

	// 用固定日期注入上班卡：直接调用内部 clockInAt（支持测试注入时间）。
	inAt := time.Date(2026, 11, 16, 21, 55, 0, 0, time.Local)
	outAt := time.Date(2026, 11, 17, 6, 10, 0, 0, time.Local)

	rec, err := punchAt(ctx, svcCtx, s.empZhang, PunchTypeClockIn, "", inAt)
	if err != nil {
		t.Fatalf("cross-day clockIn: %v", err)
	}
	if rec.PunchDate != "2026-11-16" {
		t.Fatalf("punchDate = %q, want 2026-11-16", rec.PunchDate)
	}

	rec, err = punchAt(ctx, svcCtx, s.empZhang, PunchTypeClockOut, "", outAt)
	if err != nil {
		t.Fatalf("cross-day clockOut: %v", err)
	}
	if rec.PunchDate != "2026-11-16" {
		t.Fatalf("cross-day clockOut punchDate = %q, want 2026-11-16 (归属上班日)", rec.PunchDate)
	}
	if !rec.ClockOutTime.Valid {
		t.Fatalf("clockOutTime not written")
	}

	// 日结: 跨天班 21:55→06:10 → normal，工时 8.25
	res, err := SettleDay(ctx, svcCtx, s.empZhang, "2026-11-16")
	if err != nil {
		t.Fatalf("settle: %v", err)
	}
	if res.Status != SettleNormal {
		t.Fatalf("status = %q, want normal", res.Status)
	}
	if res.WorkHours != 8.25 {
		t.Fatalf("workHours = %v, want 8.25", res.WorkHours)
	}
}

func today() string {
	return time.Now().Format(DateLayout)
}
