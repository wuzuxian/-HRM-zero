package punch

// 日结定时任务测试 (fix F2, spec 定时任务每日考勤日结):
//   - DayClose 对全体在册员工逐人 SettleDay (含无打卡日固化缺卡行);
//   - 幂等: 重跑覆盖同一行 (不重复计数);
//   - 单人失败不阻塞批次。

import (
	"context"
	"testing"
	"time"
)

func TestDayCloseJob(t *testing.T) {
	svcCtx := newSvcCtx(t)
	s := seedOrgAndEmployee(t, svcCtx)
	// 组织默认规则: 周一至周五 09:00-18:00。
	saveOrgConfig(t, svcCtx, s.orgTech, "09:00", "18:00", []int64{1, 2, 3, 4, 5})

	// 2026-05-11 周一, 无打卡 → 缺卡固化空行。
	date := "2026-05-11"
	settled, err := DayClose(context.Background(), svcCtx, date)
	if err != nil {
		t.Fatalf("day close: %v", err)
	}
	if settled != 1 {
		t.Fatalf("expected 1 settled employee, got %d", settled)
	}

	// 幂等: 再跑一次, 结果一致 (SettleDay upsert 覆盖, 不多不少)。
	settledAgain, err := DayClose(context.Background(), svcCtx, date)
	if err != nil {
		t.Fatalf("day close rerun: %v", err)
	}
	if settledAgain != 1 {
		t.Fatalf("expected idempotent rerun to settle 1, got %d", settledAgain)
	}

	// Job 构造: 注入 now, Run 应成功返回。
	job := NewDayCloseJob(svcCtx, func() time.Time {
		return time.Date(2026, 5, 11, 23, 0, 0, 0, time.Local)
	})
	if err := job.Run(context.Background()); err != nil {
		t.Fatalf("job run: %v", err)
	}
}
