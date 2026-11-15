package transfer

import (
	"context"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
	model "github.com/jzero-io/jzero-admin/server/internal/model/hrm_transfer_bill"
)

// BillTypeCode is the 人事调动 类型码 (HR203, demo-inventory §0).
const BillTypeCode = "203"

func init() {
	bill.RegisterType(BillTypeCode, bill.Handler{
		Load: func(ctx context.Context, session sqlx.Session, billUuid string) (string, string, error) {
			row, err := model.NewHrmTransferBillModel(sqlx.NewSqlConnFromSession(session)).
				FindOneByUuid(ctx, session, billUuid)
			if err != nil {
				return "", "", err
			}
			return row.BillNo, row.Status, nil
		},
		SetStatus: func(ctx context.Context, session sqlx.Session, billUuid, status string) error {
			return model.NewHrmTransferBillModel(sqlx.NewSqlConnFromSession(session)).
				UpdateFieldsByCondition(ctx, session,
					map[string]any{"status": status},
					condition.NewChain().Equal(model.Uuid, billUuid).Build()...)
		},
		OnApproved: onApproved,
	})
}

// onApproved is the 审批通过 write-back (ADR-0001): 立即生效 bills take effect
// right inside the approval transaction; delayed bills wait for the daily
// cron scan (cron.go). When the HRM-03 seam is not wired yet the write-back
// degrades to a log+skip and the bill stays applied=0 (effect.go).
func onApproved(ctx context.Context, session sqlx.Session, billUuid string) error {
	m := model.NewHrmTransferBillModel(sqlx.NewSqlConnFromSession(session))
	row, err := m.FindOneByUuid(ctx, session, billUuid)
	if err != nil {
		return err
	}
	if row.EffectiveImmediately != 1 {
		// 非立即生效：由定时任务在生效日处理，这里不动 applied。
		return nil
	}
	applied, err := applyEffect(ctx, session, effectFromRow(row))
	if err != nil {
		return err
	}
	if !applied {
		return nil
	}
	return markApplied(ctx, session, billUuid)
}

// effectFromRow snapshots the write-back payload off the bill row.
func effectFromRow(row *model.HrmTransferBill) Effect {
	return Effect{
		BillUuid:      row.Uuid,
		EmployeeUuid:  row.EmployeeUuid,
		EmployeeNo:    row.EmployeeNo,
		EmployeeName:  row.EmployeeName,
		ChangeType:    row.ChangeType,
		NewCompany:    row.NewCompany,
		NewOrg:        row.NewOrg,
		NewPosition:   row.NewPosition,
		NewPost:       row.NewPost,
		NewSalary:     row.NewSalary,
		EffectiveDate: row.EffectiveDate,
	}
}

// markApplied sets the idempotency flag so neither the approval path nor the
// cron scan applies the same bill twice.
func markApplied(ctx context.Context, session sqlx.Session, billUuid string) error {
	return model.NewHrmTransferBillModel(sqlx.NewSqlConnFromSession(session)).
		UpdateFieldsByCondition(ctx, session,
			map[string]any{"applied": 1},
			condition.NewChain().Equal(model.Uuid, billUuid).Build()...)
}
