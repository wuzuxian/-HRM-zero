// Package leavebalance 假期余额核心服务 (HRM-18)。
// 余额行: (year, employee_uuid, leave_type_code) 唯一; 应发 granted = 初始 + 调整;
// 可用 available = granted - used - frozen。
// 每次初始化/调整/请假扣减/调休入账都写一条 hrm_leave_balance_log 流水。
package leavebalance

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_leave_balance"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_leave_balance_log"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_leave_type"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero/core/status"
)

// 流水变更类型 (GLOSSARY「假期余额」/ demo-inventory 4.4).
const (
	ChangeTypeInit    = "init"    // 初始化
	ChangeTypeAdjust  = "adjust"  // 手工调整
	ChangeTypeConsume = "consume" // 请假扣减 (HRM-19 消费)
	ChangeTypeAccrue  = "accrue"  // 调休入账 (HRM-21 消费)
)

// CompLeaveCode 调休假类型编码 — 加班调休入账 (HRM-21) 固定入 comp_leave.
const CompLeaveCode = "comp_leave"

// round1 保留一位小数（余额数量统一 0.5 天步进口径，避免浮点漂移；负数走对称舍入）。
func round1(v float64) float64 {
	if v < 0 {
		return -round1(-v)
	}
	return float64(int64(v*10+0.5)) / 10
}

// findOrCreateBalance 在事务内取余额行，不存在则建空行（granted/used/frozen 全 0）。
// 余额行以 (year, employee_uuid, leave_type_code) 唯一；并发下依赖唯一索引兜底。
func findOrCreateBalance(ctx context.Context, session sqlx.Session, svcCtx *svc.ServiceContext,
	year int64, employeeUuid, leaveTypeCode string) (*hrm_leave_balance.HrmLeaveBalance, error) {
	m := modelOf(svcCtx, session)
	row, err := m.FindOneByCondition(ctx, session, balanceCondition(year, employeeUuid, leaveTypeCode).Build()...)
	if errors.Is(err, hrm_leave_balance.ErrNotFound) {
		row = &hrm_leave_balance.HrmLeaveBalance{
			Uuid:          uuid.New().String(),
			Year:          year,
			EmployeeUuid:  employeeUuid,
			LeaveTypeCode: leaveTypeCode,
		}
		if _, err = m.Insert(ctx, session, row); err != nil {
			return nil, err
		}
		return row, nil
	}
	return row, err
}

// writeLog 事务内写一条余额流水（变更后三值快照）。
func writeLog(ctx context.Context, session sqlx.Session, svcCtx *svc.ServiceContext,
	row *hrm_leave_balance.HrmLeaveBalance, changeType string, days float64,
	reason, billUuid, operatorUuid, operatorName string) error {
	if _, err := modelLogOf(svcCtx, session).Insert(ctx, session, &hrm_leave_balance_log.HrmLeaveBalanceLog{
		Uuid:          uuid.New().String(),
		BalanceUuid:   row.Uuid,
		Year:          row.Year,
		EmployeeUuid:  row.EmployeeUuid,
		LeaveTypeCode: row.LeaveTypeCode,
		ChangeType:    changeType,
		Days:          round1(days),
		GrantedAfter:  row.Granted,
		UsedAfter:     row.Used,
		FrozenAfter:   row.Frozen,
		Reason:        reason,
		BillUuid:      billUuid,
		OperatorUuid:  operatorUuid,
		OperatorName:  operatorName,
	}); err != nil {
		return err
	}
	return nil
}

// updateBalanceFields 事务内按 uuid 更新余额行字段（granted/used/frozen）。
func updateBalanceFields(ctx context.Context, session sqlx.Session, svcCtx *svc.ServiceContext,
	row *hrm_leave_balance.HrmLeaveBalance, fields map[string]any) error {
	return modelOf(svcCtx, session).UpdateFieldsByCondition(ctx, session, fields,
		balanceByUuidCondition(row.Uuid).Build()...)
}

// fetchLeaveType 校验假期类型存在且启用余额管理（流水/调整的前提）。
func fetchLeaveType(ctx context.Context, svcCtx *svc.ServiceContext, leaveTypeCode string) (*hrm_leave_type.HrmLeaveType, error) {
	lt, err := svcCtx.Model.HrmLeaveType.FindOneByCondition(ctx, nil,
		leaveTypeCodeCondition(leaveTypeCode).Build()...)
	if err != nil {
		if errors.Is(err, hrm_leave_type.ErrNotFound) {
			return nil, status.ErrorMessage(errcodes.LeaveBalanceTypeNotFoundCode, "假期类型不存在: "+leaveTypeCode)
		}
		return nil, err
	}
	if lt.BalanceEnabled != 1 {
		return nil, status.ErrorMessage(errcodes.LeaveBalanceTypeNotManagedCode, "假期类型未启用余额管理: "+lt.TypeName)
	}
	return lt, nil
}

// ConsumeLeave 请假扣减 (HRM-19 请假单消费契约): 审批通过/销假生效后调用，
// 在调用方事务内扣减余额并记 consume 流水。days > 0。
// 契约（写入 /home/wuzuxian/hrm-notes/hrm-18-notes.md）:
//   ConsumeLeave(ctx, tx, employeeUuid, leaveTypeCode, days, billUuid) error
//   - tx 为调用方 sqlx.Session（审批事务），原子性由调用方事务保证
//   - 余额行不存在时自动创建空行再扣减；可用 = granted-used-frozen < days → 10513 拒绝
//   - billUuid 传入请假单 uuid，落入流水的关联单据列
func ConsumeLeave(ctx context.Context, tx sqlx.Session, svcCtx *svc.ServiceContext,
	employeeUuid, leaveTypeCode string, days float64, billUuid string) error {
	if days <= 0 {
		return status.ErrorMessage(errcodes.LeaveBalanceInvalidCode, "扣减天数必须大于 0")
	}
	row, err := findOrCreateBalance(ctx, tx, svcCtx, yearOfCtx(ctx), employeeUuid, leaveTypeCode)
	if err != nil {
		return err
	}
	if round1(row.Granted-row.Used-row.Frozen) < round1(days) {
		return status.ErrorMessage(errcodes.LeaveBalanceInsufficientCode, "可用余额不足")
	}
	newUsed := round1(row.Used + days)
	if err = updateBalanceFields(ctx, tx, svcCtx, row, map[string]any{"used": newUsed}); err != nil {
		return err
	}
	row.Used = newUsed
	return writeLog(ctx, tx, svcCtx, row, ChangeTypeConsume, -round1(days),
		"请假扣减", billUuid, "", "")
}

// AccrueCompLeave 加班调休入账 (HRM-21 加班单消费契约): 加班审批通过后按调休倍数
// 折算小时入账 comp_leave 余额。hours > 0（小时），入账按 1 天 = 8 小时折算天数口径记账，
// 余额数值单位为天，流水 days 字段记折算后的天数（正数）。
// 契约:
//   AccrueCompLeave(ctx, tx, employeeUuid, hours, billUuid) error
//   - tx 为调用方 sqlx.Session；year 取当前自然年（context 带 x-hrm-balance-year 时优先，测试用）
//   - comp_leave 必须启用余额管理，否则 10514
//   - 入账增加 granted，记 accrue 流水
func AccrueCompLeave(ctx context.Context, tx sqlx.Session, svcCtx *svc.ServiceContext,
	employeeUuid string, hours float64, billUuid string) error {
	if hours <= 0 {
		return status.ErrorMessage(errcodes.LeaveBalanceInvalidCode, "调休入账小时数必须大于 0")
	}
	if _, err := fetchLeaveType(ctx, svcCtx, CompLeaveCode); err != nil {
		return err
	}
	year := yearOfCtx(ctx)
	row, err := findOrCreateBalance(ctx, tx, svcCtx, year, employeeUuid, CompLeaveCode)
	if err != nil {
		return err
	}
	newGranted := round1(row.Granted + hoursToDays(hours))
	if err = updateBalanceFields(ctx, tx, svcCtx, row, map[string]any{"granted": newGranted}); err != nil {
		return err
	}
	row.Granted = newGranted
	return writeLog(ctx, tx, svcCtx, row, ChangeTypeAccrue, hoursToDays(hours),
		"加班调休入账", billUuid, "", "")
}
