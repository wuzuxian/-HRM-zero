package employee

import (
	"context"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/manage_user"
)

// DisableLinkedAccount 离职禁用规则 (HRM-53 / ADR 0007 生命周期一致性) 的唯一实现:
// 员工档案进入离职态时, 其关联系统账号 manage_user.status 置禁用。两个触发点共用:
//   - 编辑档案把人员状态翻成离职 (hrm/employee edit.go)
//   - 离职单生效定时回写 (hrm/resignation cron.go)
//
// 容错与离职回写一致: 档案未命中 (仅快照单/档案被删)、未关联账号 (建档即开号之前
// 的存量档案)、账号行缺失 (悬空指针) 都幂等跳过, 不报错。在调用方事务内执行。
func DisableLinkedAccount(ctx context.Context, session sqlx.Session, employeeUuid, employeeNo string) error {
	em := hrm_employee.NewHrmEmployeeModel(sqlx.NewSqlConnFromSession(session))
	var emp *hrm_employee.HrmEmployee
	var err error
	if employeeUuid != "" {
		emp, err = em.FindOneByUuid(ctx, session, employeeUuid)
	} else {
		// 仅快照工号（档案未接入时建的旧单）：工号兜底定位
		emp, err = em.FindOneByCondition(ctx, session,
			condition.NewChain().Equal(hrm_employee.EmployeeNo, employeeNo).Build()...)
	}
	if err == sqlx.ErrNotFound {
		return nil // 档案不存在 (仅快照单), 无账号可禁用
	}
	if err != nil {
		return err
	}
	if emp.UserUuid == "" {
		return nil // 未关联账号, 无账号可禁用
	}
	return manage_user.NewManageUserModel(sqlx.NewSqlConnFromSession(session)).
		UpdateFieldsByCondition(ctx, session,
			map[string]any{"status": manage_user.StatusDisabled},
			condition.NewChain().Equal(manage_user.Uuid, emp.UserUuid).Build()...)
}
