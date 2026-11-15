package leavebalance

import (
	"context"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_leave_balance"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_leave_type"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
)

// employee 包人员状态枚举复用（HRM-03 seam）: 离职/退休 不参与初始化。
var (
	personStatusResigned = employee.PersonStatusResigned
	personStatusRetired  = employee.PersonStatusRetired
)

// managedLeaveTypes 查询启用余额管理 (balance_enabled=1) 且启用的假期类型，按 order_num 排序。
func managedLeaveTypes(ctx context.Context, svcCtx *svc.ServiceContext) ([]*hrm_leave_type.HrmLeaveType, error) {
	return svcCtx.Model.HrmLeaveType.FindByCondition(ctx, nil,
		condition.NewChain().
			Equal(hrm_leave_type.BalanceEnabled, int64(1)).
			Equal(hrm_leave_type.Status, "1").
			OrderByAsc(hrm_leave_type.OrderNum).
			OrderByAsc(hrm_leave_type.Id).
			Build()...)
}

// resolveInitEmployees 解析初始化目标员工：指定 uuids → 逐个校验存在；
// 空 → 全体在册员工（person_status NOT IN ('5','6')，HRM-03 在册口径）。
func resolveInitEmployees(ctx context.Context, svcCtx *svc.ServiceContext, employeeUuids []string) ([]string, map[string]string, error) {
	names := make(map[string]string)
	if len(employeeUuids) > 0 {
		rows, err := svcCtx.Model.HrmEmployee.FindByCondition(ctx, nil,
			condition.NewChain().In(hrm_employee.Uuid, employeeUuids).Build()...)
		if err != nil {
			return nil, nil, err
		}
		found := make(map[string]bool, len(rows))
		for _, row := range rows {
			found[row.Uuid] = true
			names[row.Uuid] = row.Name
		}
		for _, u := range employeeUuids {
			if !found[u] {
				return nil, nil, errEmployeeNotFound(u)
			}
		}
		return employeeUuids, names, nil
	}

	// 全体在册员工
	rows, err := svcCtx.Model.HrmEmployee.FindByCondition(ctx, nil,
		condition.NewChain().
			NotIn(hrm_employee.PersonStatus, []string{personStatusResigned, personStatusRetired}).
			OrderByAsc(hrm_employee.Id).
			Build()...)
	if err != nil {
		return nil, nil, err
	}
	out := make([]string, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.Uuid)
		names[row.Uuid] = row.Name
	}
	return out, names, nil
}

// initOneBalance 事务内初始化一行余额：granted = 类型默认天数，记 init 流水。
func initOneBalance(ctx context.Context, session sqlx.Session, svcCtx *svc.ServiceContext,
	year int64, employeeUuid, employeeName string, lt *hrm_leave_type.HrmLeaveType) error {
	m := modelOf(svcCtx, session)
	row := &hrm_leave_balance.HrmLeaveBalance{
		Uuid:          uuid.New().String(),
		Year:          year,
		EmployeeUuid:  employeeUuid,
		EmployeeNo:    employeeNoOf(ctx, svcCtx, session, employeeUuid),
		EmployeeName:  employeeName,
		LeaveTypeCode: lt.TypeCode,
		Granted:       round1(lt.DefaultDays),
	}
	if _, err := m.Insert(ctx, session, row); err != nil {
		return err
	}
	return writeLog(ctx, session, svcCtx, row, ChangeTypeInit, row.Granted,
		"初始化本年度额度", "", "", "")
}

func employeeNoOf(ctx context.Context, svcCtx *svc.ServiceContext, session sqlx.Session, employeeUuid string) string {
	row, err := svcCtx.Model.HrmEmployee.FindOneByCondition(ctx, session,
		condition.NewChain().Equal(hrm_employee.Uuid, employeeUuid).Build()...)
	if err != nil || row == nil {
		return ""
	}
	return row.EmployeeNo
}
