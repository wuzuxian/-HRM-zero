package leave

import (
	"context"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
	model "github.com/jzero-io/jzero-admin/server/internal/model/hrm_leave_bill"
)

// StatusApproved is the 单据状态 required before 销假.
const StatusApproved = bill.StatusApproved

// BillTypeCode lives in validate.go; this file registers the 单据 handler
// against the shared approval base (HRM-06).
// OnApproved（审批通过事务内）：仅余额管理类假期扣减余额（leavebalance.ConsumeLeave，
// 可用不足 → 10513 整体回滚）；非余额管理类型直接通过不扣减。
// 销假的差额多退少补在 Cancel 里做（不在审批事务）.
func init() {
	bill.RegisterType(BillTypeCode, bill.Handler{
		Load: func(ctx context.Context, session sqlx.Session, billUuid string) (string, string, error) {
			row, err := model.NewHrmLeaveBillModel(sqlx.NewSqlConnFromSession(session)).
				FindOneByUuid(ctx, session, billUuid)
			if err != nil {
				return "", "", err
			}
			return row.BillNo, row.Status, nil
		},
		SetStatus: func(ctx context.Context, session sqlx.Session, billUuid, status string) error {
			return model.NewHrmLeaveBillModel(sqlx.NewSqlConnFromSession(session)).
				UpdateFieldsByCondition(ctx, session,
					map[string]any{"status": status},
					condition.NewChain().Equal(model.Uuid, billUuid).Build()...)
		},
		OnApproved: onApprovedConsumeBalance,
	})
}
