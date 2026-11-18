package resignation

import (
	"context"
	"errors"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	employeelogic "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/employee"
	hrmemployee "github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	model "github.com/jzero-io/jzero-admin/server/internal/model/hrm_resignation_bill"
	"github.com/jzero-io/jzero-admin/server/internal/scheduler"
)

// 离职生效回写（spec 用户故事 22 / demo-inventory §3.4 / HRM-03 人员状态枚举）:
// 审批通过仅记录（applied=0 不写档案），定时任务在离职日当天统一将员工
// person_status→'5'(离职)。与 transfer 不同，这里没有接缝解耦 —— 回写目标
// (hrm_employee.person_status) 是本仓库已合入的 HRM-03 档案列，直接更新；
// 档案行不存在（仅快照单/档案被删）按幂等跳过，不阻塞批量。
const personStatusResigned = "5" // 离职 (HRM-03 enum: 1 正式/2 试用期/3 实习生/4 临时工/5 离职/6 退休)

// applyResignation writes back one 审批通过 bill whose 离职日 has come (or whose
// write-back was retried): flip the employee's 人员状态 to 离职('5') and mark the
// bill applied=1. Idempotent via the applied flag; runs inside the caller's
// transaction (cron tx). Missing archive rows are tolerated (snapshot contract)
// but still marked applied so the scan does not retry them forever.
// HRM-53 (ADR 0007 生命周期一致性): 置离职的同一事务内同步禁用关联账号。
func applyResignation(ctx context.Context, session sqlx.Session, row *model.HrmResignationBill) error {
	if row.Applied == 1 {
		return nil // idempotent: already written back
	}
	m := hrmemployee.NewHrmEmployeeModel(sqlx.NewSqlConnFromSession(session))

	chain := condition.NewChain().Equal(hrmemployee.Uuid, row.EmployeeUuid)
	if row.EmployeeUuid == "" {
		// 仅快照工号（档案未接入时建的旧单）：工号兜底定位
		chain = condition.NewChain().Equal(hrmemployee.EmployeeNo, row.EmployeeNo)
	}
	if err := m.UpdateFieldsByCondition(ctx, session,
		map[string]any{"person_status": personStatusResigned}, chain.Build()...); err != nil {
		return err
	}
	count, err := m.CountByCondition(ctx, session, chain.Build()...)
	if err != nil {
		return err
	}
	if count == 0 {
		logx.WithContext(ctx).Infof(
			"resignation: 离职回写未命中员工档案（bill=%s employee=%s/%s uuid=%q），仅标记 applied",
			row.BillNo, row.EmployeeNo, row.EmployeeName, row.EmployeeUuid)
	}
	// HRM-53: 离职生效同一事务内禁用关联账号 (登录状态校验随之拒绝该员工)。
	// 规则的唯一实现放在员工包 (编辑档案翻离职同走此函数), 本处只调用;
	// 容错与档案回写一致: 档案未命中/未关联账号/账号行缺失都幂等跳过。
	if err := employeelogic.DisableLinkedAccount(ctx, session, row.EmployeeUuid, row.EmployeeNo); err != nil {
		return err
	}
	return model.NewHrmResignationBillModel(sqlx.NewSqlConnFromSession(session)).
		UpdateFieldsByCondition(ctx, session,
			map[string]any{"applied": int64(1)},
			condition.NewChain().Equal(model.Uuid, row.Uuid).Build()...)
}

// ScanPendingResigned is the daily cron scan behind 离职生效: every 审批通过 bill
// with applied=0 whose 离职日期 is today or earlier gets the employee's
// person_status flipped to '5'(离职) and its applied flag set, each bill in its
// own transaction so one bad bill never blocks the batch.
//
// Idempotent by construction (applied flag), so RunAtStart + restarts are safe.
// Bills not yet approved keep applied=0 but are filtered by status; bills
// approved but dated in the future wait until their 离职日.
func ScanPendingResigned(ctx context.Context, conn sqlx.SqlConn, now time.Time) (int, error) {
	today := now.Format(time.DateOnly)
	rows, err := model.NewHrmResignationBillModel(conn).FindByCondition(ctx, nil,
		condition.NewChain().
			Equal(model.Status, billStatusApproved).
			Equal(model.Applied, 0).
			OrderByAsc(model.Id).
			Build()...)
	if err != nil {
		return 0, err
	}

	appliedCount := 0
	for _, row := range rows {
		if row.ResignDate == "" || row.ResignDate > today {
			continue // 离职日未到
		}
		err := conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
			// Re-read inside the tx: state may have changed since the scan.
			current, err := model.NewHrmResignationBillModel(sqlx.NewSqlConnFromSession(session)).
				FindOneByUuid(ctx, session, row.Uuid)
			if err != nil {
				return err
			}
			if current.Status != billStatusApproved || current.Applied != 0 {
				return nil // someone else got it; skip silently
			}
			if current.ResignDate == "" || current.ResignDate > today {
				return nil // 离职日未到（state changed since the scan query）
			}
			return applyResignation(ctx, session, current)
		})
		if err != nil {
			if errors.Is(err, model.ErrNotFound) {
				continue // bill deleted between scan and apply
			}
			logx.WithContext(ctx).Errorf("resignation: 离职生效回写失败 bill=%s billNo=%s: %v", row.Uuid, row.BillNo, err)
			continue
		}
		appliedCount++
	}
	return appliedCount, nil
}

// NewResignedScanJob returns the scheduler Job for the 离职日 scan. Registered
// by internal/custom at server boot (see transfer.NewEffectiveScanJob for the
// reference pattern).
func NewResignedScanJob(conn sqlx.SqlConn, now func() time.Time) scheduler.Job {
	return scheduler.Job{
		Name:       "hrm-resignation-effective-scan",
		Interval:   24 * time.Hour,
		RunAtStart: true, // catch up on 离职日 missed while the server was down
		Run: func(ctx context.Context) error {
			if now == nil {
				now = time.Now
			}
			count, err := ScanPendingResigned(ctx, conn, now())
			if err != nil {
				return err
			}
			if count > 0 {
				logx.WithContext(ctx).Infof("resignation: 离职日扫描完成，置离职 %d 张离职单对应员工", count)
			}
			return nil
		},
	}
}
