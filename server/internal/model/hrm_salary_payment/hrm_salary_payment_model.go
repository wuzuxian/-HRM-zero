package hrm_salary_payment

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmSalaryPaymentModel = (*customHrmSalaryPaymentModel)(nil)

type (
	// HrmSalaryPaymentModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmSalaryPaymentModel.
	HrmSalaryPaymentModel interface {
		hrmSalaryPaymentModel
	}

	customHrmSalaryPaymentModel struct {
		*defaultHrmSalaryPaymentModel
	}
)

// NewHrmSalaryPaymentModel returns a model for the database table.
func NewHrmSalaryPaymentModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmSalaryPaymentModel {
	return &customHrmSalaryPaymentModel{
		defaultHrmSalaryPaymentModel: newHrmSalaryPaymentModel(conn, op...),
	}
}
