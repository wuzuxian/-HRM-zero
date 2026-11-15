package leavebalance

import (
	"context"

	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_leave_balance"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
)

// ChangeTypeRefund 销假退回流水变更类型（HRM-19 销假差额「多退」侧）。
const ChangeTypeRefund = "refund"

// RefundLeave 销假退回 (HRM-19 请假销假差额「多退」侧消费契约):
// 在调用方事务内把已扣减余额退回 days 天（used -= days）并记 refund 流水。
// days > 0（退回的天数，正数传入）；流水 days 记 正数（与 consume 的负数对称，GLOSSARY 流水口径 正=增 负=减 中退回即「增可用」）。
// 契约（写入 /home/wuzuxian/hrm-notes/hrm-19-notes.md，供未来销假流程复用）:
//   RefundLeave(ctx, tx, svcCtx, employeeUuid, leaveTypeCode, days, billUuid) error
//   - tx 为调用方 sqlx.Session（销假事务），原子性由调用方事务保证
//   - 余额行不存在（该员工该类型从未扣减过）→ 10540 幂等保护拒绝
//   - 退回后 used < 0 → 10540 拒绝（重复销假/超额退回）
//   - billUuid 传入请假销假单 uuid，落入流水的关联单据列
// 与 ConsumeLeave 配对使用：实际>预计 由调用方再调 ConsumeLeave 补扣差额（「少补」侧）。
func RefundLeave(ctx context.Context, tx sqlx.Session, svcCtx *svc.ServiceContext,
	employeeUuid, leaveTypeCode string, days float64, billUuid string) error {
	if days <= 0 {
		return status.ErrorMessage(errcodes.LeaveBalanceInvalidCode, "退回天数必须大于 0")
	}
	row, err := findOrCreateBalance(ctx, tx, svcCtx, yearOfCtx(ctx), employeeUuid, leaveTypeCode)
	if err != nil {
		return err
	}
	if round1(row.Used-days) < 0 {
		return status.ErrorMessage(errcodes.LeaveRefundOverflowCode, "退回天数超过已扣减余额")
	}
	newUsed := round1(row.Used - days)
	if err = updateBalanceFields(ctx, tx, svcCtx, row, map[string]any{"used": newUsed}); err != nil {
		return err
	}
	row.Used = newUsed
	return writeLog(ctx, tx, svcCtx, row, ChangeTypeRefund, round1(days),
		"销假退回", billUuid, "", "")
}

// WithYearDo 返回携带指定年度的 context 执行 fn（ConsumeLeave/RefundLeave/AccrueCompLeave
// 在 fn 内解析年度；调用方如请假单按预计开始时间所在自然年结算时使用）。
func WithYearDo(ctx context.Context, year int64, fn func(yctx context.Context) error) error {
	return fn(WithYear(ctx, year))
}

var _ = hrm_leave_balance.ErrNotFound
