package transfer

import (
	"context"
	"errors"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
	model "github.com/jzero-io/jzero-admin/server/internal/model/hrm_transfer_bill"
	"github.com/jzero-io/jzero-admin/server/internal/scheduler"
)

// ScanPendingEffective is the daily cron scan behind 生效 write-backs:
// every 审批通过 bill with applied=0 that is due — 非立即生效 with 生效日期
// today or earlier, or 立即生效 whose write-back was skipped earlier (e.g.
// while the HRM-03 seam was not wired) — gets its write-back applied and its
// applied flag flipped. Each bill runs in its own transaction, so one bad
// bill never blocks the rest of the day's batch.
//
// Idempotent by construction (applied flag), so RunAtStart + restarts are
// safe. Bills whose HRM-03 write-back seam is not wired keep applied=0 and
// are retried by every subsequent scan (effect.go degrade path).
func ScanPendingEffective(ctx context.Context, conn sqlx.SqlConn, now time.Time) (int, error) {
	today := now.Format(time.DateOnly)
	rows, err := model.NewHrmTransferBillModel(conn).FindByCondition(ctx, nil,
		condition.NewChain().
			Equal(model.Status, bill.StatusApproved).
			Equal(model.Applied, 0).
			OrderByAsc(model.Id).
			Build()...)
	if err != nil {
		return 0, err
	}

	// due = 立即生效（回写仍待处理）OR 非立即生效且生效日期已到.
	due := make([]*model.HrmTransferBill, 0, len(rows))
	for _, row := range rows {
		if row.EffectiveImmediately == 1 || (row.EffectiveDate != "" && row.EffectiveDate <= today) {
			due = append(due, row)
		}
	}
	rows = due

	appliedCount := 0
	for _, row := range rows {
		err := conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
			// Re-read inside the tx: state may have changed since the scan.
			current, err := model.NewHrmTransferBillModel(sqlx.NewSqlConnFromSession(session)).
				FindOneByUuid(ctx, session, row.Uuid)
			if err != nil {
				return err
			}
			if current.Status != bill.StatusApproved || current.Applied != 0 {
				return nil // someone else got it; skip silently
			}
			if current.EffectiveImmediately != 1 && !(current.EffectiveDate != "" && current.EffectiveDate <= today) {
				return nil // 生效日未到（state changed since the scan query）
			}
			applied, err := applyEffect(ctx, session, effectFromRow(current))
			if err != nil {
				return err
			}
			if !applied {
				return nil // HRM-03 seam not wired; keep pending, retry next scan
			}
			return markApplied(ctx, session, current.Uuid)
		})
		if err != nil {
			if errors.Is(err, model.ErrNotFound) {
				continue // bill deleted between scan and apply
			}
			logx.WithContext(ctx).Errorf("transfer: 生效回写失败 bill=%s billNo=%s: %v", row.Uuid, row.BillNo, err)
			continue
		}
		appliedCount++
	}
	return appliedCount, nil
}

// NewEffectiveScanJob returns the reusable scheduler Job for the 调动生效日
// scan. Registered by internal/custom at server boot; HRM-10 (离职生效) and
// the attendance daily jobs should follow the same pattern with their own
// Job entries.
func NewEffectiveScanJob(conn sqlx.SqlConn, now func() time.Time) scheduler.Job {
	return scheduler.Job{
		Name:       "hrm-transfer-effective-scan",
		Interval:   24 * time.Hour,
		RunAtStart: true, // catch up on 生效日 missed while the server was down
		Run: func(ctx context.Context) error {
			if now == nil {
				now = time.Now
			}
			count, err := ScanPendingEffective(ctx, conn, now())
			if err != nil {
				return err
			}
			if count > 0 {
				logx.WithContext(ctx).Infof("transfer: 生效日扫描完成，回写 %d 张调动单", count)
			}
			return nil
		},
	}
}
