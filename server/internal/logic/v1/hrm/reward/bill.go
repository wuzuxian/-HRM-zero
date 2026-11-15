package reward

import (
	"context"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
	model "github.com/jzero-io/jzero-admin/server/internal/model/hrm_reward_bill"
)

// BillTypeCode is the 奖惩单 类型码 (HR215, GLOSSARY 单据编号).
const BillTypeCode = "215"

func init() {
	bill.RegisterType(BillTypeCode, bill.Handler{
		Load: func(ctx context.Context, session sqlx.Session, billUuid string) (string, string, error) {
			row, err := model.NewHrmRewardBillModel(sqlx.NewSqlConnFromSession(session)).
				FindOneByUuid(ctx, session, billUuid)
			if err != nil {
				return "", "", err
			}
			return row.BillNo, row.Status, nil
		},
		SetStatus: func(ctx context.Context, session sqlx.Session, billUuid, status string) error {
			return model.NewHrmRewardBillModel(sqlx.NewSqlConnFromSession(session)).
				UpdateFieldsByCondition(ctx, session,
					map[string]any{"status": status},
					condition.NewChain().Equal(model.Uuid, billUuid).Build()...)
		},
		OnApproved: onApproved,
	})
}

// onApproved is the 审批通过 write-back (ADR-0001) and the 薪资数据源同步
// (spec 用户故事 53): the approved 奖惩单 becomes the data source for the
// 奖惩奖励/奖惩扣款 salary items — HRM-34 薪资项目 / HRM-37 试算 read approved
// bills by (employee, salary_month). The sync is a marker flip (synced=1)
// inside the approval transaction, so it is synchronous, atomic and
// idempotent; no separate salary table is needed yet.
func onApproved(ctx context.Context, session sqlx.Session, billUuid string) error {
	m := model.NewHrmRewardBillModel(sqlx.NewSqlConnFromSession(session))
	row, err := m.FindOneByUuid(ctx, session, billUuid)
	if err != nil {
		return err
	}
	return markSynced(ctx, session, row)
}

// markSynced flips the 薪资数据源 marker. salary_month defaults to the
// 发生日期's month when left empty at creation, so every approved bill is
// addressable by (employee_uuid, salary_month).
func markSynced(ctx context.Context, session sqlx.Session, row *model.HrmRewardBill) error {
	if row.Synced == 1 {
		return nil // idempotent: already a salary data source
	}
	fields := map[string]any{"synced": int64(1)}
	if row.SalaryMonth == "" {
		fields["salary_month"] = monthOf(row.OccurDate)
	}
	return model.NewHrmRewardBillModel(sqlx.NewSqlConnFromSession(session)).
		UpdateFieldsByCondition(ctx, session, fields,
			condition.NewChain().Equal(model.Uuid, row.Uuid).Build()...)
}
