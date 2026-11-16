package hrm_salary_payslip

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmSalaryPayslipModel = (*customHrmSalaryPayslipModel)(nil)

type (
	// HrmSalaryPayslipModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmSalaryPayslipModel.
	HrmSalaryPayslipModel interface {
		hrmSalaryPayslipModel
	}

	customHrmSalaryPayslipModel struct {
		*defaultHrmSalaryPayslipModel
	}
)

// NewHrmSalaryPayslipModel returns a model for the database table.
func NewHrmSalaryPayslipModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmSalaryPayslipModel {
	return &customHrmSalaryPayslipModel{
		defaultHrmSalaryPayslipModel: newHrmSalaryPayslipModel(conn, op...),
	}
}
