package punch

// HRM-16 日结定时任务 (spec: 定时任务每日执行考勤日结): 全体在册员工当日
// 逐人调 SettleDay (复用三级回退 ResolveEffectiveRule)。SettleDay 幂等
// (一人一天一行 upsert 语义, 重跑覆盖同列), RunAtStart + 重启安全。

import (
	"context"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/scheduler"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
)

// DayClose 日结扫描: 对全体员工 (与手动日结 Settle.resolveEmployees 同口径,
// 跳过无部门员工) 执行 date 当日 SettleDay。返回处理人数。逐员工独立错误容忍:
// 单人失败记日志继续, 不阻塞当日批次 (与 transfer.ScanPendingEffective 同策略)。
func DayClose(ctx context.Context, svcCtx *svc.ServiceContext, date string) (int, error) {
	rows, err := svcCtx.Model.HrmEmployee.FindByCondition(ctx, nil,
		condition.NewChain().NotEqual(hrm_employee.OrgUuid, "").Build()...)
	if err != nil {
		return 0, err
	}
	settled := 0
	for _, employee := range rows {
		if _, err := SettleDay(ctx, svcCtx, employee.Uuid, date); err != nil {
			logx.WithContext(ctx).Errorf("punch: 日结失败 employee=%s date=%s: %v", employee.Uuid, date, err)
			continue
		}
		settled++
	}
	return settled, nil
}

// NewDayCloseJob returns the scheduler Job for the daily 日结. Registered by
// internal/custom at server boot (follows transfer.NewEffectiveScanJob).
func NewDayCloseJob(svcCtx *svc.ServiceContext, now func() time.Time) scheduler.Job {
	return scheduler.Job{
		Name:       "hrm-punch-day-close",
		Interval:   24 * time.Hour,
		RunAtStart: true, // 补日结服务器宕机期间错过的当日
		Run: func(ctx context.Context) error {
			if now == nil {
				now = time.Now
			}
			n, err := DayClose(ctx, svcCtx, now().Format(DateLayout))
			if err != nil {
				return err
			}
			logx.WithContext(ctx).Infof("punch: 日结完成 date=%s settled=%d", now().Format(DateLayout), n)
			return nil
		},
	}
}
