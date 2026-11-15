package contract

import (
	"context"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee_contract"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
)

// shared helpers for the contract domain (List/Add/Edit/Delete/Renew).

// findContractByUuid 按主键 uuid 定位合同行；不存在 → ContractNotFoundCode。
func findContractByUuid(ctx context.Context, svcCtx *svc.ServiceContext, uuid string) (*hrm_employee_contract.HrmEmployeeContract, error) {
	row, err := svcCtx.Model.HrmEmployeeContract.FindOneByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_employee_contract.Uuid, uuid).Build()...)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.ContractNotFoundCode, "劳动合同不存在")
		}
		return nil, err
	}
	return row, nil
}

// resolveEmployee 按员工档案 uuid 带出快照（工号/姓名/部门）；不存在 → EmployeeNotFoundCode。
func resolveEmployee(ctx context.Context, svcCtx *svc.ServiceContext, employeeUuid string) (*hrm_employee.HrmEmployee, error) {
	row, err := svcCtx.Model.HrmEmployee.FindOneByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_employee.Uuid, employeeUuid).Build()...)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.EmployeeNotFoundCode, "员工不存在")
		}
		return nil, err
	}
	return row, nil
}

func validContractStatus(v string) bool {
	return v == ContractStatusActive || v == ContractStatusExpired || v == ContractStatusTerminated
}

// validDate 校验 YYYY-MM-DD 格式。
func validDate(v string) bool {
	_, err := timeParseDate(v)
	return err == nil
}
