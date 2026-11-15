package resignation

import (
	"context"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
	model "github.com/jzero-io/jzero-admin/server/internal/model/hrm_resignation_bill"
)

// billStatusApproved mirrors bill.StatusApproved; kept local so this file has a
// single dependency surface for the scan.
const billStatusApproved = bill.StatusApproved

// BillTypeCode lives in validate.go; this file registers the 单据 handler
// against the shared approval base (HRM-06). 审批通过只记账（applied=0）：
// 员工置离职统一由定时任务在离职日处理（cron.go），即时审批不做档案写回.
func init() {
	bill.RegisterType(BillTypeCode, bill.Handler{
		Load: func(ctx context.Context, session sqlx.Session, billUuid string) (string, string, error) {
			row, err := model.NewHrmResignationBillModel(sqlx.NewSqlConnFromSession(session)).
				FindOneByUuid(ctx, session, billUuid)
			if err != nil {
				return "", "", err
			}
			return row.BillNo, row.Status, nil
		},
		SetStatus: func(ctx context.Context, session sqlx.Session, billUuid, status string) error {
			return model.NewHrmResignationBillModel(sqlx.NewSqlConnFromSession(session)).
				UpdateFieldsByCondition(ctx, session,
					map[string]any{"status": status},
					condition.NewChain().Equal(model.Uuid, billUuid).Build()...)
		},
		// OnApproved intentionally nil-ish: 审批通过仅记录，不即时改档案
		//（spec 用户故事 22：定时任务在离职日当天统一置离职）.
		OnApproved: nil,
	})
}
