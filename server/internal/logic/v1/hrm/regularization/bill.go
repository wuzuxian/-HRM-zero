package regularization

import (
	"context"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
	model "github.com/jzero-io/jzero-admin/server/internal/model/hrm_regularization_bill"
)

// BillTypeCode lives in validate.go; this file registers the 单据 handler
// against the shared approval base (HRM-06) and owns the 审批通过 write-back.

func init() {
	bill.RegisterType(BillTypeCode, bill.Handler{
		Load: func(ctx context.Context, session sqlx.Session, billUuid string) (string, string, error) {
			row, err := model.NewHrmRegularizationBillModel(sqlx.NewSqlConnFromSession(session)).
				FindOneByUuid(ctx, session, billUuid)
			if err != nil {
				return "", "", err
			}
			return row.BillNo, row.Status, nil
		},
		SetStatus: func(ctx context.Context, session sqlx.Session, billUuid, status string) error {
			return model.NewHrmRegularizationBillModel(sqlx.NewSqlConnFromSession(session)).
				UpdateFieldsByCondition(ctx, session,
					map[string]any{"status": status},
					condition.NewChain().Equal(model.Uuid, billUuid).Build()...)
		},
		OnApproved: onApproved,
	})
}

// onApproved is the 审批通过 write-back (spec 用户故事 20 / demo-inventory §3.2):
// inside the approval transaction, flip the employee's 人员状态 to 正式
// (person_status='1') and write the 转正日期 — preferring the bill's
// 实际转正日期, falling back to the 预计转正日期 snapshot. Runs as ONE
// transaction with the status flip and 审批记录 insert (ADR-0001): a failing
// write-back rolls back the whole approval. Idempotent via the applied flag.
func onApproved(ctx context.Context, session sqlx.Session, billUuid string) error {
	m := model.NewHrmRegularizationBillModel(sqlx.NewSqlConnFromSession(session))
	row, err := m.FindOneByUuid(ctx, session, billUuid)
	if err != nil {
		return err
	}
	return applyRegularization(ctx, session, row)
}

// applyRegularization performs the archive write-back for one approved bill.
// Kept separate from onApproved so tests (and any future re-drive) can replay
// it safely.
func applyRegularization(ctx context.Context, session sqlx.Session, row *model.HrmRegularizationBill) error {
	if row.Applied == 1 {
		return nil // idempotent: already written back
	}
	formalDate := row.FormalDate
	if formalDate == "" {
		formalDate = row.ExpectedFormalDate // 优先单据转正日期，其次预计转正日期
	}
	fields := map[string]any{
		"person_status": "1", // 正式 (HRM-03 enum)
	}
	if formalDate != "" {
		fields["formal_date"] = formalDate
	}
	if _, err := updateEmployeeByUuid(ctx, session, row.EmployeeUuid, fields); err != nil {
		return err
	}
	return model.NewHrmRegularizationBillModel(sqlx.NewSqlConnFromSession(session)).
		UpdateFieldsByCondition(ctx, session,
			map[string]any{"applied": int64(1)},
			condition.NewChain().Equal(model.Uuid, row.Uuid).Build()...)
}
