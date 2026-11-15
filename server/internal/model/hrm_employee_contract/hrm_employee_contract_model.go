package hrm_employee_contract

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmEmployeeContractModel = (*customHrmEmployeeContractModel)(nil)

type (
	// HrmEmployeeContractModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmEmployeeContractModel.
	HrmEmployeeContractModel interface {
		hrmEmployeeContractModel
	}

	customHrmEmployeeContractModel struct {
		*defaultHrmEmployeeContractModel
	}
)

// NewHrmEmployeeContractModel returns a model for the database table.
func NewHrmEmployeeContractModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmEmployeeContractModel {
	return &customHrmEmployeeContractModel{
		defaultHrmEmployeeContractModel: newHrmEmployeeContractModel(conn, op...),
	}
}
