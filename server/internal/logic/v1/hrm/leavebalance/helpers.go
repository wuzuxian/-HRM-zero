package leavebalance

import (
	"context"
	"strconv"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_leave_balance"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_leave_balance_log"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_leave_type"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
)

// yearCtxKey 测试/调用方覆盖入账年度的 context key（正常取当前自然年）。
type yearCtxKey struct{}

// WithYear 返回携带指定年度的 context（仅 ConsumeLeave/AccrueCompLeave 内部解析）。
func WithYear(ctx context.Context, year int64) context.Context {
	return context.WithValue(ctx, yearCtxKey{}, year)
}

// yearOfCtx 解析余额所属年度：context 有显式年度用之，否则当前自然年。
func yearOfCtx(ctx context.Context) int64 {
	if v, ok := ctx.Value(yearCtxKey{}).(int64); ok && v > 0 {
		return v
	}
	return int64(time.Now().Year())
}

// HoursPerDay 调休入账折算口径：1 天 = 8 小时。
const HoursPerDay = 8.0

// hoursToDays 加班小时 → 调休天数（1 位小数）。
func hoursToDays(hours float64) float64 {
	return round1(hours / HoursPerDay)
}

// modelOf / modelLogOf: 事务内构造 model（sqlx.NewSqlConnFromSession 与 reward.onApproved 同款）。
func modelOf(svcCtx *svc.ServiceContext, session sqlx.Session) hrm_leave_balance.HrmLeaveBalanceModel {
	if session == nil {
		return svcCtx.Model.HrmLeaveBalance
	}
	return hrm_leave_balance.NewHrmLeaveBalanceModel(sqlx.NewSqlConnFromSession(session))
}

func modelLogOf(svcCtx *svc.ServiceContext, session sqlx.Session) hrm_leave_balance_log.HrmLeaveBalanceLogModel {
	if session == nil {
		return svcCtx.Model.HrmLeaveBalanceLog
	}
	return hrm_leave_balance_log.NewHrmLeaveBalanceLogModel(sqlx.NewSqlConnFromSession(session))
}

func balanceCondition(year int64, employeeUuid, leaveTypeCode string) condition.Chain {
	return condition.NewChain().
		Equal(hrm_leave_balance.Year, year).
		Equal(hrm_leave_balance.EmployeeUuid, employeeUuid).
		Equal(hrm_leave_balance.LeaveTypeCode, leaveTypeCode)
}

func balanceByUuidCondition(balanceUuid string) condition.Chain {
	return condition.NewChain().Equal(hrm_leave_balance.Uuid, balanceUuid)
}

func leaveTypeCodeCondition(code string) condition.Chain {
	return condition.NewChain().Equal(hrm_leave_type.TypeCode, code)
}

// AtoiYear 年度字符串 → int64（"" 或非法 → 0 表示不过滤）。
func AtoiYear(s string) int64 {
	if s == "" {
		return 0
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return v
}
