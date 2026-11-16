package monthreport

// 月报自动锁定定时任务测试 (fix F2, spec 定时任务月报锁定, 锁定日可配置):
//   - 锁定日当天 (如 5 日) 自动锁定上一自然月;
//   - 锁定日前一天 (4 日) 不锁定;
//   - 已锁定月重跑幂等跳过 (双锁拒绝);
//   - lockDay 非法/越界不触发。

import (
	"context"
	"testing"
	"time"

	"github.com/jzero-io/jzero-admin/server/internal/svc"
)

func mustLockRecordAction(t *testing.T, svcCtx *svc.ServiceContext, yearMonth string) string {
	t.Helper()
	rows, err := svcCtx.Model.HrmAttendanceMonthLock.FindByCondition(context.Background(), nil)
	if err != nil {
		t.Fatalf("find lock records: %v", err)
	}
	for _, row := range rows {
		if row.YearMonth == yearMonth {
			return row.Action
		}
	}
	return ""
}

func TestAutoLockIfDue(t *testing.T) {
	svcCtx := newMrSvcCtx(t)
	ctx := context.Background()

	// 2026-04 是上一个自然月（now = 2026-05-05，锁定日 5）。
	// 先给 4 月造一行数据（员工有部门 → aggregateRows 有输入）。
	orgUuid := seedOrg(t, svcCtx, "公司A", "")
	seedEmployee(t, svcCtx, "张三", "E0001", orgUuid)

	// ---- 锁定日前一天 (4 日): 不触发 ----
	due, err := AutoLockIfDue(ctx, svcCtx, time.Date(2026, 5, 4, 10, 0, 0, 0, time.Local), 5)
	if err != nil {
		t.Fatalf("auto lock on 4th: %v", err)
	}
	if due {
		t.Fatal("must not lock before lockDay")
	}
	if act := mustLockRecordAction(t, svcCtx, "2026-04"); act != "" {
		t.Fatalf("expected no lock record on 4th, got %s", act)
	}

	// ---- 锁定日当天 (5 日): 锁定上一月 ----
	due, err = AutoLockIfDue(ctx, svcCtx, time.Date(2026, 5, 5, 10, 0, 0, 0, time.Local), 5)
	if err != nil {
		t.Fatalf("auto lock on 5th: %v", err)
	}
	if !due {
		t.Fatal("expected lock on lockDay")
	}
	if act := mustLockRecordAction(t, svcCtx, "2026-04"); act != ActionLock {
		t.Fatalf("expected lock record for 2026-04, got %q", act)
	}

	// ---- 幂等: 同月再次触发 (如重启) 静默跳过, 不追加第二条锁定流水 ----
	due, err = AutoLockIfDue(ctx, svcCtx, time.Date(2026, 5, 5, 22, 0, 0, 0, time.Local), 5)
	if err != nil {
		t.Fatalf("auto lock twice: %v", err)
	}
	if due {
		t.Fatal("already locked month must be skipped")
	}
	rows, err := svcCtx.Model.HrmAttendanceMonthLock.FindByCondition(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	lockCount := 0
	for _, row := range rows {
		if row.YearMonth == "2026-04" && row.Action == ActionLock {
			lockCount++
		}
	}
	if lockCount != 1 {
		t.Fatalf("expected exactly 1 lock record for 2026-04, got %d", lockCount)
	}

	// ---- 非法 lockDay 不触发 ----
	due, err = AutoLockIfDue(ctx, svcCtx, time.Date(2026, 5, 5, 10, 0, 0, 0, time.Local), 0)
	if err != nil || due {
		t.Fatalf("lockDay=0 must be a no-op, got due=%v err=%v", due, err)
	}
	due, err = AutoLockIfDue(ctx, svcCtx, time.Date(2026, 5, 5, 10, 0, 0, 0, time.Local), 32)
	if err != nil || due {
		t.Fatalf("lockDay=32 must be a no-op, got due=%v err=%v", due, err)
	}

	// ---- Job 构造: 注入非锁定日日期, Run 返回 nil 不锁定 ----
	job := NewMonthAutoLockJob(svcCtx, 5, func() time.Time {
		return time.Date(2026, 5, 4, 0, 0, 0, 0, time.Local)
	})
	if err := job.Run(context.Background()); err != nil {
		t.Fatalf("job run: %v", err)
	}
	if act := mustLockRecordAction(t, svcCtx, "2026-03"); act != "" {
		t.Fatalf("job with non-lockDay date must not lock, got %s", act)
	}
}
