package regularization

import (
	"context"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	hrmemployee "github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
)

// updateEmployeeByUuid updates 员工档案 (HRM-03) fields inside the caller's
// transaction — the 转正回写 seam. Rows are matched by uuid first; when the
// bill was created without an employeeUuid link (仅快照工号), the 工号 falls
// back as the locator. No row matched is not an error (the snapshot contract
// tolerates archive rows deleted after bill creation).
func updateEmployeeByUuid(ctx context.Context, session sqlx.Session, employeeUuid string, fields map[string]any) (int64, error) {
	m := hrmemployee.NewHrmEmployeeModel(sqlx.NewSqlConnFromSession(session))

	chain := condition.NewChain().Equal(hrmemployee.Uuid, employeeUuid)
	if err := m.UpdateFieldsByCondition(ctx, session, fields, chain.Build()...); err != nil {
		return 0, err
	}

	count, err := m.CountByCondition(ctx, session, chain.Build()...)
	if err != nil {
		return 0, err
	}
	return count, nil
}
