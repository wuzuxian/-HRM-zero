package monthreport

// Lock/Unlock/Recalc 月报锁定域操作（demo-inventory 4.4 按钮）。

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_attendance_month_lock"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_attendance_month_report"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
)

// lockAction 锁定记录动作。
const (
	ActionLock   = "lock"
	ActionUnlock = "unlock"
	ActionRecalc = "recalc"
)

// statusError 业务错误码包装。
func statusError(code int, err error) error {
	return status.ErrorMessage(status.Code(code), err.Error())
}

// jstatusError 业务错误码包装（字符串消息版）。
func jstatusError(code int, msg string) error {
	return status.ErrorMessage(status.Code(code), msg)
}

// newUuid uuid v4。
func newUuid() string {
	return uuid.New().String()
}

// insertLockRecord 锁定流水（谁何时 lock/unlock/recalc）。
func insertLockRecord(ctx context.Context, session sqlx.Session, svcCtx *svc.ServiceContext,
	yearMonth, action, operatorUuid, operatorName, remark string) error {
	_, err := svcCtx.Model.HrmAttendanceMonthLock.Insert(ctx, session, &hrm_attendance_month_lock.HrmAttendanceMonthLock{
		Uuid:         newUuid(),
		YearMonth:    yearMonth,
		Action:       action,
		OperatorUuid: operatorUuid,
		OperatorName: operatorName,
		Remark:       remark,
	})
	return err
}

// upsertSnapshotRow 插入或覆盖单行快照（按 (year_month, employee_uuid) 唯一键）。
func upsertSnapshotRow(ctx context.Context, session sqlx.Session, svcCtx *svc.ServiceContext, row MonthReportRow, hash string) error {
	model := svcCtx.Model.HrmAttendanceMonthReport
	existing, err := model.FindOneByCondition(ctx, session,
		condition.NewChain().
			Equal(hrm_attendance_month_report.YearMonth, row.YearMonth).
			Equal(hrm_attendance_month_report.EmployeeUuid, row.EmployeeUuid).Build()...)
	if err != nil {
		// ErrNotFound → 插入
		_, insErr := model.Insert(ctx, session, &hrm_attendance_month_report.HrmAttendanceMonthReport{
			Uuid:                 newUuid(),
			YearMonth:            row.YearMonth,
			EmployeeUuid:         row.EmployeeUuid,
			EmployeeNo:           row.EmployeeNo,
			EmployeeName:         row.EmployeeName,
			OrgUuid:              row.OrgUuid,
			OrgName:              row.OrgName,
			AttGroupName:         row.AttGroupName,
			ExpectedDays:         row.ExpectedDays,
			AttendDays:           row.AttendDays,
			LateCount:            row.LateCount,
			EarlyCount:           row.EarlyCount,
			MissingCount:         row.MissingCount,
			AbsentDays:           row.AbsentDays,
			WorkdayOvertimeHours: row.WorkdayOtH,
			RestdayOvertimeHours: row.RestdayOtH,
			HolidayOvertimeHours: row.HolidayOtH,
			CompAccrueHours:      row.CompAccrueH,
			RowHash:              hash,
			Changed:              0,
		})
		return insErr
	}
	// 覆盖（重算缓存刷新；锁定 diff 不走此路径）
	return model.UpdateFieldsByCondition(ctx, session, map[string]any{
		"employee_no":             row.EmployeeNo,
		"employee_name":           row.EmployeeName,
		"org_uuid":                row.OrgUuid,
		"org_name":                row.OrgName,
		"att_group_name":          row.AttGroupName,
		"expected_days":           row.ExpectedDays,
		"attend_days":             row.AttendDays,
		"late_count":              row.LateCount,
		"early_count":             row.EarlyCount,
		"missing_count":           row.MissingCount,
		"absent_days":             row.AbsentDays,
		"workday_overtime_hours":  row.WorkdayOtH,
		"restday_overtime_hours":  row.RestdayOtH,
		"holiday_overtime_hours":  row.HolidayOtH,
		"comp_accrue_hours":       row.CompAccrueH,
		"row_hash":                hash,
	}, condition.NewChain().Equal(hrm_attendance_month_report.Uuid, existing.Uuid).Build()...)
}

// LockMonth 锁定月报：实时聚合写入快照（幂等覆盖）+ 流水。锁定后月报走快照口径。
func LockMonth(ctx context.Context, svcCtx *svc.ServiceContext, yearMonth, operatorUuid, operatorName, remark string) error {
	yearMonth, err := validateYearMonth(yearMonth)
	if err != nil {
		return jstatusError(errcodes.MonthReportInvalidCode, err.Error())
	}
	employees, err := listMonthEmployees(ctx, svcCtx, nil, nil)
	if err != nil {
		return err
	}
	rows, err := aggregateRows(ctx, svcCtx, yearMonth, employees)
	if err != nil {
		return err
	}
	err = withTx(ctx, svcCtx, func(session sqlx.Session) error {
		for _, row := range rows {
			if err := upsertSnapshotRow(ctx, session, svcCtx, row, rowHashOf(row)); err != nil {
				return err
			}
		}
		return insertLockRecord(ctx, session, svcCtx, yearMonth, ActionLock, operatorUuid, operatorName, remark)
	})
	if err != nil {
		return err
	}
	logx.WithContext(ctx).Infof("month report locked %s rows=%d", yearMonth, len(rows))
	return nil
}

// UnlockMonth 解锁：删除该月快照行（月报恢复实时聚合）+ 流水。
func UnlockMonth(ctx context.Context, svcCtx *svc.ServiceContext, yearMonth, operatorUuid, operatorName, remark string) error {
	yearMonth, err := validateYearMonth(yearMonth)
	if err != nil {
		return jstatusError(errcodes.MonthReportInvalidCode, err.Error())
	}
	locked, err := isLocked(ctx, svcCtx, yearMonth)
	if err != nil {
		return err
	}
	if !locked {
		return jstatusError(errcodes.MonthReportNotFoundCode, fmt.Sprintf("月份 %s 未锁定", yearMonth))
	}
	return withTx(ctx, svcCtx, func(session sqlx.Session) error {
		if err := svcCtx.Model.HrmAttendanceMonthReport.DeleteByCondition(ctx, session,
			condition.NewChain().Equal(hrm_attendance_month_report.YearMonth, yearMonth).Build()...); err != nil {
			return err
		}
		return insertLockRecord(ctx, session, svcCtx, yearMonth, ActionUnlock, operatorUuid, operatorName, remark)
	})
}

// RecalcMonth 重新计算：重跑聚合。
//   - 未锁定月 → 覆盖快照行（缓存刷新），changed 清零；
//   - 锁定月 → 快照保留（保锁定口径），与快照 hash diff，不一致行标 changed=1。
func RecalcMonth(ctx context.Context, svcCtx *svc.ServiceContext, yearMonth string) (bool, int64, error) {
	yearMonth, err := validateYearMonth(yearMonth)
	if err != nil {
		return false, 0, jstatusError(errcodes.MonthReportInvalidCode, err.Error())
	}
	locked, err := isLocked(ctx, svcCtx, yearMonth)
	if err != nil {
		return false, 0, err
	}
	employees, err := listMonthEmployees(ctx, svcCtx, nil, nil)
	if err != nil {
		return locked, 0, err
	}
	rows, err := aggregateRows(ctx, svcCtx, yearMonth, employees)
	if err != nil {
		return locked, 0, err
	}
	var changedCount int64
	err = withTx(ctx, svcCtx, func(session sqlx.Session) error {
		changedCount = 0
		for _, row := range rows {
			hash := rowHashOf(row)
			if locked {
				// 锁定月：查快照行 diff。
				existing, findErr := svcCtx.Model.HrmAttendanceMonthReport.FindOneByCondition(ctx, session,
					condition.NewChain().
						Equal(hrm_attendance_month_report.YearMonth, row.YearMonth).
						Equal(hrm_attendance_month_report.EmployeeUuid, row.EmployeeUuid).Build()...)
				if findErr != nil {
					// 无快照行（锁定后新入职/首行）→ 记 changed=1 快照行
					if insErr := insertChangedSnapshot(ctx, session, svcCtx, row, hash); insErr != nil {
						return insErr
					}
					changedCount++
					continue
				}
				newChanged := int64(0)
				if existing.RowHash != hash {
					newChanged = 1
					changedCount++
				}
				if err := svcCtx.Model.HrmAttendanceMonthReport.UpdateFieldsByCondition(ctx, session,
					map[string]any{"changed": newChanged},
					condition.NewChain().Equal(hrm_attendance_month_report.Uuid, existing.Uuid).Build()...); err != nil {
					return err
				}
			} else {
				// 未锁定月：覆盖快照（缓存刷新），changed 清零。
				if err := upsertSnapshotRow(ctx, session, svcCtx, row, hash); err != nil {
					return err
				}
			}
		}
		return insertLockRecord(ctx, session, svcCtx, yearMonth, ActionRecalc, "", "系统", "")
	})
	return locked, changedCount, err
}

// insertChangedSnapshot 锁定月重算时出现的新行（快照无此员工）→ 落 changed=1 行。
func insertChangedSnapshot(ctx context.Context, session sqlx.Session, svcCtx *svc.ServiceContext, row MonthReportRow, hash string) error {
	_, err := svcCtx.Model.HrmAttendanceMonthReport.Insert(ctx, session, &hrm_attendance_month_report.HrmAttendanceMonthReport{
		Uuid:                 newUuid(),
		YearMonth:            row.YearMonth,
		EmployeeUuid:         row.EmployeeUuid,
		EmployeeNo:           row.EmployeeNo,
		EmployeeName:         row.EmployeeName,
		OrgUuid:              row.OrgUuid,
		OrgName:              row.OrgName,
		AttGroupName:         row.AttGroupName,
		ExpectedDays:         row.ExpectedDays,
		AttendDays:           row.AttendDays,
		LateCount:            row.LateCount,
		EarlyCount:           row.EarlyCount,
		MissingCount:         row.MissingCount,
		AbsentDays:           row.AbsentDays,
		WorkdayOvertimeHours: row.WorkdayOtH,
		RestdayOvertimeHours: row.RestdayOtH,
		HolidayOvertimeHours: row.HolidayOtH,
		CompAccrueHours:      row.CompAccrueH,
		RowHash:              hash,
		Changed:              1,
	})
	return err
}

// withTx 事务包装。
func withTx(ctx context.Context, svcCtx *svc.ServiceContext, fn func(session sqlx.Session) error) error {
	return svcCtx.SqlxConn.TransactCtx(ctx, func(_ context.Context, session sqlx.Session) error {
		return fn(session)
	})
}

// timeNow 测试注入点保留（当前直接 time.Now）。
func timeNow() time.Time { return time.Now() }
