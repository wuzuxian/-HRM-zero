package leave

import (
	"context"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/leavebalance"
	model "github.com/jzero-io/jzero-admin/server/internal/model/hrm_leave_bill"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_leave_type"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
)

// svcCtxKey carries the minimal *svc.ServiceContext (tests wire a real one via
// WithSvcCtx) into the approval OnApproved closure, which the bill base calls
// with only (ctx, session, billUuid).
type svcCtxKey struct{}

// WithSvcCtx attaches the minimal ServiceContext used by balance settlement.
func WithSvcCtx(ctx context.Context, svcCtx *svc.ServiceContext) context.Context {
	return context.WithValue(ctx, svcCtxKey{}, svcCtx)
}

func svcCtxOf(ctx context.Context) *svc.ServiceContext {
	if v, ok := ctx.Value(svcCtxKey{}).(*svc.ServiceContext); ok {
		return v
	}
	return nil
}

// onApprovedConsumeBalance 审批通过事务内的业务回写：仅余额管理类假期扣减余额
// （leavebalance.ConsumeLeave，可用不足 → 10513 整体回滚）；
// 非余额管理类型直接通过不扣减（demo 口径）；类型行已删除时同样跳过扣减.
// 年度取预计开始时间所在自然年（销假差额也按该年度结算）.
func onApprovedConsumeBalance(ctx context.Context, session sqlx.Session, billUuid string) error {
	row, err := model.NewHrmLeaveBillModel(sqlx.NewSqlConnFromSession(session)).FindOneByUuid(ctx, session, billUuid)
	if err != nil {
		return err
	}
	lt, err := leaveTypeOf(ctx, session, row.LeaveTypeCode)
	if err != nil || lt == nil {
		return err
	}
	if lt.BalanceEnabled != 1 {
		return nil // 非余额管理类型直接通过不扣减
	}
	// 审批通过即全额扣减预计天数；差额多退少补发生在销假（Cancel）事务.
	return leavebalance.WithYearDo(ctx, planYearOf(row.PlanStartTime), func(yctx context.Context) error {
		return leavebalance.ConsumeLeave(yctx, session, svcCtxOf(ctx), row.EmployeeUuid, row.LeaveTypeCode, row.PlanDays, row.Uuid)
	})
}

// settleBalance 销假差额多退少补（Cancel 事务内调用）：
// diff = 实际天数 - 预计天数；diff < 0 退回（RefundLeave），diff > 0 再扣（ConsumeLeave，可用不足 → 10513 回滚）。
// 仅余额管理类假期结算.
func settleBalance(ctx context.Context, session sqlx.Session, svcCtx *svc.ServiceContext,
	row *model.HrmLeaveBill, actualDays float64) error {
	lt, err := leaveTypeOf(ctx, session, row.LeaveTypeCode)
	if err != nil || lt == nil {
		return err
	}
	if lt.BalanceEnabled != 1 {
		return nil
	}
	return settleLeaveBalance(ctx, session, svcCtx, row, actualDays)
}

// settleLeaveBalance 按 (row 的预计天数口径, 实际天数) 差额多退少补.
// svcCtx 仅作 leavebalance.modelOf 的 nil-session 回退使用；本文件所有调用都在
// 事务内（session != nil），svcCtx 允许为 nil（OnApproved 闭包无 svcCtx 可透传）.
func settleLeaveBalance(ctx context.Context, session sqlx.Session, svcCtx *svc.ServiceContext,
	row *model.HrmLeaveBill, actualDays float64) error {
	diff := roundHalfStep(actualDays - row.PlanDays)
	year := planYearOf(row.PlanStartTime)
	switch {
	case diff < 0:
		// 实际 < 预计：退回差额（used -= |diff|，refund 流水）
		return leavebalance.WithYearDo(ctx, year, func(yctx context.Context) error {
			return leavebalance.RefundLeave(yctx, session, svcCtx, row.EmployeeUuid, row.LeaveTypeCode, -diff, row.Uuid)
		})
	case diff > 0:
		// 实际 > 预计：再扣差额（可用不足 → 10513）
		return leavebalance.WithYearDo(ctx, year, func(yctx context.Context) error {
			return leavebalance.ConsumeLeave(yctx, session, svcCtx, row.EmployeeUuid, row.LeaveTypeCode, diff, row.Uuid)
		})
	}
	return nil
}

// leaveTypeOf 事务内取假期类型行；不存在返回 (nil, nil)（类型已删除 → 跳过余额结算）.
func leaveTypeOf(ctx context.Context, session sqlx.Session, code string) (*hrm_leave_type.HrmLeaveType, error) {
	lt, err := hrm_leave_type.NewHrmLeaveTypeModel(sqlx.NewSqlConnFromSession(session)).
		FindOneByCondition(ctx, session, conditionByTypeCode(code).Build()...)
	if err != nil {
		if err == hrm_leave_type.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	return lt, nil
}

// planYearOf 从 YYYY-MM-DD HH:MM:SS 取自然年（非法时间回退当前年）。
func planYearOf(billTime string) int64 {
	parts := strings.Fields(billTime)
	if len(parts) == 0 {
		return int64(time.Now().Year())
	}
	t, err := time.ParseInLocation("2006-01-02", parts[0], time.Local)
	if err != nil {
		return int64(time.Now().Year())
	}
	return int64(t.Year())
}

// roundHalfStep 差额对齐 0.5 天步进（浮点漂移防御，与余额 0.5 口径一致）.
func roundHalfStep(v float64) float64 {
	return RoundHalfStep(v)
}
