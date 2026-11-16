package monthreport

// HRM-23 月报自动锁定定时任务 (spec: 定时任务…月报锁定, 锁定日可配置如次月 5 日):
// 每天 (含启动时) 检查 today.Day() == lockDay 时, 自动锁定上一个自然月
// ("yyyy-MM") 的月报 — 快照聚合写入复用 LockMonth (幂等: 已锁定月跳过)。
// lockDay 取 config hrm.attendanceLockDay (默认 5), 由 custom.Start 注入。

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/scheduler"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
)

// AutoLockSystemOperator 自动锁定操作人 (月报锁定流水 operator)。
const AutoLockSystemOperator = "系统"

// AutoLockIfDue 在 date 当天检查是否到达锁定日; 到达则自动锁定 previous month
// (date 所在月的前一个自然月), 已锁定月静默跳过 (幂等)。lockDay<=0 或 >31 视为
// 未启用。返回是否执行了锁定。
func AutoLockIfDue(ctx context.Context, svcCtx *svc.ServiceContext, now time.Time, lockDay int) (bool, error) {
	if lockDay <= 0 || lockDay > 31 || now.Day() != lockDay {
		return false, nil
	}
	// 上一个自然月: 本月 1 号减一天 → 取其 yyyy-MM。
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	prevMonth := firstOfMonth.AddDate(0, 0, -1).Format(YearMonthLayout)

	locked, err := isLocked(ctx, svcCtx, prevMonth)
	if err != nil {
		return false, err
	}
	if locked {
		return false, nil // 已锁定 (重跑/重启幂等): 跳过
	}
	if err := LockMonth(ctx, svcCtx, prevMonth, "", AutoLockSystemOperator,
		fmt.Sprintf("自动锁定 (锁定日 %d 日)", lockDay)); err != nil {
		return false, err
	}
	logx.WithContext(ctx).Infof("monthreport: 月报自动锁定完成 yearMonth=%s lockDay=%d", prevMonth, lockDay)
	return true, nil
}

// NewMonthAutoLockJob returns the scheduler Job for the monthly auto-lock.
// lockDay is read from hrm.attendanceLockDay at registration time (config
// listener changes take effect on restart; the job itself is a thin closure).
func NewMonthAutoLockJob(svcCtx *svc.ServiceContext, lockDay int, now func() time.Time) scheduler.Job {
	return scheduler.Job{
		Name:       "hrm-monthreport-auto-lock",
		Interval:   24 * time.Hour,
		RunAtStart: true, // 锁定日当天重启/补跑同样命中 (幂等跳过已锁月)
		Run: func(ctx context.Context) error {
			if now == nil {
				now = time.Now
			}
			_, err := AutoLockIfDue(ctx, svcCtx, now(), lockDay)
			return err
		},
	}
}
