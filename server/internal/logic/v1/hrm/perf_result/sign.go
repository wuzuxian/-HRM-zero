package perf_result

import (
	"context"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	hrmuser "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/hrmuser"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_result"
)

// 签收（HRM-45）：员工对本人结果行签收（signed=1 + sign_time + sign_by），幂等：
// 已签收直接成功（demo 验收：重复点击不报错不重复记录）。
// 越权守卫：JWT 登录账号 → hrm_employee.user_uuid 反查员工 uuid，必须等于结果行
// employee_uuid（结果未发布/已作废不可签收 → ErrResultInvalid）。

// actorEmployeeUuid resolves the JWT 登录账号 to the bound 员工 uuid; "" when
// unbound or unauthenticated (logic-test path resolves identity explicitly).
func actorEmployeeUuid(ctx context.Context, conn sqlx.SqlConn) (string, error) {
	user := hrmuser.Current(ctx)
	if user.Uuid == "" {
		return "", nil
	}
	rows, err := hrm_employee.NewHrmEmployeeModel(conn).FindByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_employee.UserUuid, user.Uuid).Build()...)
	if err != nil {
		return "", err
	}
	if len(rows) == 0 {
		return "", nil
	}
	return rows[0].Uuid, nil
}

// SignResultRow signs one 结果行 by its employee (本人), idempotent.
func SignResultRow(ctx context.Context, conn sqlx.SqlConn, resultUuid, actorEmployeeUuid, actorName string) error {
	row, err := hrm_perf_result.NewHrmPerfResultModel(conn).FindOneByUuid(ctx, nil, resultUuid)
	if err != nil {
		if err == hrm_perf_result.ErrNotFound {
			return ErrResultNotFound
		}
		return err
	}
	if row.ResultStatus != ResultStatusNormal {
		return ErrResultInvalid
	}
	if row.Published != 1 {
		return ErrResultInvalid
	}
	if actorEmployeeUuid == "" || actorEmployeeUuid != row.EmployeeUuid {
		return ErrResultSignDenied
	}
	if row.Signed == 1 {
		return nil // 幂等：已签收
	}
	return hrm_perf_result.NewHrmPerfResultModel(conn).UpdateFieldsByCondition(ctx, nil, map[string]any{
		"signed":    1,
		"sign_time": nowStr(),
		"sign_by":   actorName,
	}, condition.NewChain().Equal(hrm_perf_result.Uuid, resultUuid).Build()...)
}
