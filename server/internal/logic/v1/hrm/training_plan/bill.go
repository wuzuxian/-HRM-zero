package training_plan

import (
	"context"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
	model "github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_plan"
)

// BillTypeCode is the 培训计划 类型码 (HR216, demo-inventory §0).
const BillTypeCode = "216"

func init() {
	bill.RegisterType(BillTypeCode, bill.Handler{
		Load: func(ctx context.Context, session sqlx.Session, billUuid string) (string, string, error) {
			row, err := model.NewHrmTrainingPlanModel(sqlx.NewSqlConnFromSession(session)).
				FindOneByUuid(ctx, session, billUuid)
			if err != nil {
				return "", "", err
			}
			return row.BillNo, row.Status, nil
		},
		SetStatus: func(ctx context.Context, session sqlx.Session, billUuid, status string) error {
			return model.NewHrmTrainingPlanModel(sqlx.NewSqlConnFromSession(session)).
				UpdateFieldsByCondition(ctx, session,
					map[string]any{"status": status},
					condition.NewChain().Equal(model.Uuid, billUuid).Build()...)
		},
		// OnApproved is the 发布 write-back (ADR-0001): 审批通过即发布 —
		// published 草稿(0) → 已发布(1) inside the approval transaction.
		// Idempotent: re-approving is impossible (审批通过 terminal), and the
		// flag write is a no-op once set.
		OnApproved: func(ctx context.Context, session sqlx.Session, billUuid string) error {
			return model.NewHrmTrainingPlanModel(sqlx.NewSqlConnFromSession(session)).
				UpdateFieldsByCondition(ctx, session,
					map[string]any{"published": 1},
					condition.NewChain().Equal(model.Uuid, billUuid).Build()...)
		},
	})
}
