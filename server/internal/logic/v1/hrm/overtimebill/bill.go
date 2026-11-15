package overtimebill

import (
	"context"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/attendancegroup"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/leavebalance"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/overtimerule"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	model "github.com/jzero-io/jzero-admin/server/internal/model/hrm_overtime_bill"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero/core/status"
)

// svcCtxKey carries the *svc.ServiceContext (tests wire a real one via WithSvcCtx)
// into the approval OnApproved closure, which the bill base calls with only
// (ctx, session, billUuid).
type svcCtxKey struct{}

// WithSvcCtx attaches the ServiceContext used by 调休入账 settlement.
func WithSvcCtx(ctx context.Context, svcCtx *svc.ServiceContext) context.Context {
	return context.WithValue(ctx, svcCtxKey{}, svcCtx)
}

func svcCtxOf(ctx context.Context) *svc.ServiceContext {
	if v, ok := ctx.Value(svcCtxKey{}).(*svc.ServiceContext); ok {
		return v
	}
	return nil
}

// BillTypeCode lives in hours.go; this file registers the 单据 handler
// against the shared approval base (HRM-06, typeCode 217).
//
// 免审批口径（issue 正文「免审批→直接核定：仍走底座但自动通过」）:
//   - 提交（SubmitOvertimeBill）时按 HRM-13 ResolveOvertimeRule 判定当天策略并快照
//     rule_source/comp_multiplier；ApprovalRequired=false 时在同一事务内
//     SubmitBill + ApproveBill 自动通过——审批记录完整（submit+approve），
//     单据不落核定（confirm_hours=0），折算基准=申请时长。
//   - ApprovalRequired=true 时正常走 底座生命周期；审批人先 Confirm（核定时长+补偿方式）
//     再 ApproveBill；未核定时折算基准回退申请时长、补偿方式回退调休.
//
// OnApproved（审批通过事务内）: 按基准×CompMultiplier 折算调休小时
// （基准=核定时长>0?核定:申请，均按最小计算单位向下取整），月上限守卫
// （当月已折算小时合计 + 本次 ≤ MonthlyLimitHours，0=不限），
// 补偿方式=调休 时 leavebalance.AccrueCompLeave 入账 comp_leave（流水 accrue），
// 月上限超限 → 10574 整体回滚；补偿方式=其他 时仅落 accrue_hours 记录不入余额.
func init() {
	bill.RegisterType(BillTypeCode, bill.Handler{
		Load: func(ctx context.Context, session sqlx.Session, billUuid string) (string, string, error) {
			row, err := model.NewHrmOvertimeBillModel(sqlx.NewSqlConnFromSession(session)).
				FindOneByUuid(ctx, session, billUuid)
			if err != nil {
				return "", "", err
			}
			return row.BillNo, row.Status, nil
		},
		SetStatus: func(ctx context.Context, session sqlx.Session, billUuid, statusCode string) error {
			return model.NewHrmOvertimeBillModel(sqlx.NewSqlConnFromSession(session)).
				UpdateFieldsByCondition(ctx, session,
					map[string]any{"status": statusCode},
					condition.NewChain().Equal(model.Uuid, billUuid).Build()...)
		},
		OnApproved: onApprovedAccrueCompLeave,
	})
}

// SubmitOvertimeBill 提交加班单：按 ResolveOvertimeRule 判定 需审批/免审批。
// 免审批 → 同一事务内 SubmitBill + ApproveBill（底座自动通过）；
// 需审批 → SubmitBill 正常进入 审批中。返回最终单据状态.
func SubmitOvertimeBill(ctx context.Context, conn sqlx.SqlConn, svcCtx *svc.ServiceContext,
	billUuid string, op bill.Operator, opinion string) (string, error) {
	// 先判规则（快照 rule_source/comp_multiplier 到单据行）
	if err := snapshotRule(ctx, conn, svcCtx, billUuid); err != nil {
		return "", err
	}
	if err := bill.SubmitBill(ctx, conn, BillTypeCode, billUuid, op, opinion); err != nil {
		return "", err
	}
	row, err := modelOf(svcCtx, conn).FindOneByUuid(ctx, conn, billUuid)
	if err != nil {
		return "", err
	}
	if !submissionRequiresApproval(ctx, svcCtx, row) {
		if err := bill.ApproveBill(ctx, conn, BillTypeCode, billUuid, op, "免审批自动通过"); err != nil {
			return "", err
		}
		return StatusApproved, nil
	}
	return StatusApproving, nil
}

// submissionRequiresApproval 复算提交时的当天策略（与 snapshotRule 同源规则）.
func submissionRequiresApproval(ctx context.Context, svcCtx *svc.ServiceContext, row *model.HrmOvertimeBill) bool {
	groupUuid, orgUuid := attendanceScopeOfUuid(ctx, svcCtx, row.EmployeeUuid)
	resolved, err := overtimerule.ResolveOvertimeRule(ctx, svcCtx, groupUuid, orgUuid, row.DayType)
	if err != nil || !resolved.Found {
		return true // 规则未命中兜底：需审批（HRM-13 notes 建议按 需审批+倍数1 处理）
	}
	return resolved.ApprovalRequired
}

// snapshotRule 提交时快照 命中规则来源 + 调休倍数（未命中规则保持倍数 1）.
func snapshotRule(ctx context.Context, conn sqlx.SqlConn, svcCtx *svc.ServiceContext, billUuid string) error {
	m := modelOf(svcCtx, conn)
	row, err := m.FindOneByUuid(ctx, conn, billUuid)
	if err != nil {
		if err == model.ErrNotFound {
			return status.ErrorMessage(errcodes.OvertimeNotFoundCode, "加班单不存在")
		}
		return err
	}
	if row.Status != StatusUnsubmitted && row.Status != StatusRejected && row.Status != StatusRevoked {
		return status.ErrorMessage(errcodes.OvertimeNotEditableCode, "当前单据状态不允许提交")
	}
	groupUuid, orgUuid := attendanceScopeOfUuid(ctx, svcCtx, row.EmployeeUuid)
	resolved, err := overtimerule.ResolveOvertimeRule(ctx, svcCtx, groupUuid, orgUuid, row.DayType)
	if err != nil {
		return err
	}
	fields := map[string]any{"rule_source": "", "comp_multiplier": 1.0}
	if resolved.Found {
		fields["rule_source"] = resolved.Source
		fields["comp_multiplier"] = resolved.CompMultiplier
	}
	return m.UpdateFieldsByCondition(ctx, conn, fields,
		condition.NewChain().Equal(model.Uuid, billUuid).Build()...)
}

// attendanceScopeOfUuid 按员工档案 uuid 取考勤组/组织上下文；员工不存在返回空.
func attendanceScopeOfUuid(ctx context.Context, svcCtx *svc.ServiceContext, employeeUuid string) (groupUuid, orgUuid string) {
	if employeeUuid == "" || svcCtx == nil {
		return "", ""
	}
	emp, err := svcCtx.Model.HrmEmployee.FindOneByUuid(ctx, nil, employeeUuid)
	if err != nil {
		return "", ""
	}
	return attendanceScopeOf(ctx, svcCtx, emp)
}

// attendanceScopeOf 取员工考勤组/组织上下文（HRM-12 ResolveEffectiveRule 命中 group 时
// 把它的 SourceUuid 作为 groupUuid；员工 OrgUuid 始终作为 orgUuid 供组织默认回退）.
func attendanceScopeOf(ctx context.Context, svcCtx *svc.ServiceContext,
	emp *hrm_employee.HrmEmployee) (groupUuid, orgUuid string) {
	resolved, err := attendancegroup.ResolveEffectiveRule(ctx, svcCtx, emp)
	if err == nil && resolved.Found && resolved.Source == attendancegroup.RuleSourceGroup {
		return resolved.SourceUuid, emp.OrgUuid
	}
	return "", emp.OrgUuid
}

// onApprovedAccrueCompLeave 审批通过事务内的业务回写：按核定（未核定用申请）×倍数折算
// 调休小时，月上限守卫，补偿方式=调休 时入账 comp_leave 余额；幂等（settled=1 跳过）.
func onApprovedAccrueCompLeave(ctx context.Context, session sqlx.Session, billUuid string) error {
	svcCtx := svcCtxOf(ctx)
	row, err := modelOf(svcCtx, session).FindOneByUuid(ctx, session, billUuid)
	if err != nil {
		return err
	}
	if row.Settled == 1 {
		return nil // 幂等：已入账不再重复折算
	}
	if row.CompType == "" {
		// 未核定的免审批单/未填补偿方式的单默认调休（issue 口径：补偿方式 调休/其他，默认调休）
		row.CompType = CompTypeLeave
	}

	// 折算基准：核定时长 > 0 用核定，否则申请；按最小计算单位向下取整
	minUnit := resolveMinUnitMinutes(ctx, svcCtx, row.EmployeeUuid, row.DayType)
	baseHours := row.ConfirmHours
	if baseHours <= 0 {
		baseHours = row.ApplyHours
	}
	baseHours = RoundHoursToUnit(baseHours, minUnit)
	accrueHours := ConvertAccrueHours(baseHours, row.CompMultiplier)

	fields := map[string]any{
		"accrue_hours": accrueHours,
		"settled":      1,
	}
	if row.CompType == CompTypeLeave && accrueHours > 0 {
		// 月上限守卫：当月已折算小时合计 + 本次 ≤ MonthlyLimitHours（0=不限），超限 → 10574 回滚
		if err := guardMonthlyLimit(ctx, session, svcCtx, row, accrueHours); err != nil {
			return err
		}
		if err := leavebalance.WithYearDo(ctx, overtimeYearOf(row.OvertimeDate), func(yctx context.Context) error {
			return leavebalance.AccrueCompLeave(yctx, session, svcCtx, row.EmployeeUuid, accrueHours, row.Uuid)
		}); err != nil {
			return err
		}
	}
	return modelOf(svcCtx, session).UpdateFieldsByCondition(ctx, session, fields,
		condition.NewChain().Equal(model.Uuid, billUuid).Build()...)
}

// guardMonthlyLimit 当月上限守卫：同一员工同月已审批通过的加班单折算小时合计 + 本次
// ≤ 规则 MonthlyLimitHours（0=不限）→ 否则 10574（审批整体回滚）.
func guardMonthlyLimit(ctx context.Context, session sqlx.Session, svcCtx *svc.ServiceContext,
	row *model.HrmOvertimeBill, accrueHours float64) error {
	groupUuid, orgUuid := attendanceScopeOfUuid(ctx, svcCtx, row.EmployeeUuid)
	resolved, err := overtimerule.ResolveOvertimeRule(ctx, svcCtx, groupUuid, orgUuid, row.DayType)
	if err != nil {
		return err
	}
	if !resolved.Found || resolved.MonthlyLimitHours <= 0 {
		return nil // 不限
	}
	// 当月已折算合计 = 该员工同月（overtime_date 前缀 YYYY-MM）已审批通过单据的 accrue_hours 之和
	month := overtimeMonthPrefix(row.OvertimeDate)
	rows, err := modelOf(svcCtx, session).FindByCondition(ctx, session,
		condition.NewChain().
			Equal(model.EmployeeUuid, row.EmployeeUuid).
			Equal(model.Status, StatusApproved).
			Equal(model.Settled, 1).
			Like(model.OvertimeDate, month+"%").
			Build()...)
	if err != nil {
		return err
	}
	var used float64
	for _, r := range rows {
		if r.Uuid == row.Uuid {
			continue
		}
		used += r.AccrueHours
	}
	if used+accrueHours > resolved.MonthlyLimitHours+1e-9 {
		return status.ErrorMessage(errcodes.OvertimeMonthlyLimitExceededCode,
			"折算调休小时超出该员工当月加班月上限")
	}
	return nil
}

// overtimeYearOf 从 YYYY-MM-DD 取自然年（非法回退当前年）——余额入账年度口径.
func overtimeYearOf(date string) int64 {
	t, err := time.ParseInLocation("2006-01-02", date, time.Local)
	if err != nil {
		return int64(time.Now().Year())
	}
	return int64(t.Year())
}

// overtimeMonthPrefix 取 YYYY-MM 前缀.
func overtimeMonthPrefix(date string) string {
	if len(date) >= 7 {
		return date[:7]
	}
	return date
}
