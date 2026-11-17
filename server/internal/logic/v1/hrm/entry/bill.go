package entry

import (
	"context"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
	model "github.com/jzero-io/jzero-admin/server/internal/model/hrm_entry_bill"
)

// BillTypeCode is the 员工入职单 类型码 (HR201, GLOSSARY 单据编号).
const BillTypeCode = "201"

func init() {
	bill.RegisterType(BillTypeCode, bill.Handler{
		Load: func(ctx context.Context, session sqlx.Session, billUuid string) (string, string, error) {
			row, err := model.NewHrmEntryBillModel(sqlx.NewSqlConnFromSession(session)).
				FindOneByUuid(ctx, session, billUuid)
			if err != nil {
				return "", "", err
			}
			return row.BillNo, row.Status, nil
		},
		SetStatus: func(ctx context.Context, session sqlx.Session, billUuid, status string) error {
			return model.NewHrmEntryBillModel(sqlx.NewSqlConnFromSession(session)).
				UpdateFieldsByCondition(ctx, session,
					map[string]any{"status": status},
					condition.NewChain().Equal(model.Uuid, billUuid).Build()...)
		},
		OnApproved: onApproved,
	})
}

// onApproved is the 审批通过 write-back (ADR-0001): 自动建档 (demo-inventory 3.1
// 开源源码逻辑 — 审批通过后自动建员工档案). Runs inside the approval transaction;
// idempotent via the employee_uuid 回写 marker (重复审批不重复建档).
// HRM-50 建档即开号: createEmployeeArchive 内部在同一事务接唯一开号接缝
// ProvisionAccount, 开号失败随审批整体回滚 (approval state 不落库)。
// svcCtx 经 WithSvcCtx 从 ctx 取回 (offer/leave/cardreplacement 同款约定)。
func onApproved(ctx context.Context, session sqlx.Session, billUuid string) error {
	m := model.NewHrmEntryBillModel(sqlx.NewSqlConnFromSession(session))
	row, err := m.FindOneByUuid(ctx, session, billUuid)
	if err != nil {
		return err
	}
	if row.EmployeeUuid != "" {
		return nil // idempotent: 员工档案 already created for this 入职单
	}
	employeeUuid, err := createEmployeeArchive(ctx, svcCtxOf(ctx), session, row)
	if err != nil {
		return err
	}
	// 回写 employee_uuid inside the same tx — doubles as the idempotency marker.
	return m.UpdateFieldsByCondition(ctx, session,
		map[string]any{"employee_uuid": employeeUuid},
		condition.NewChain().Equal(model.Uuid, row.Uuid).Build()...)
}
