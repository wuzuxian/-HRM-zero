package hrm_salary_confirm_bill

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmSalaryConfirmBillModel = (*customHrmSalaryConfirmBillModel)(nil)

type (
	// HrmSalaryConfirmBillModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmSalaryConfirmBillModel.
	HrmSalaryConfirmBillModel interface {
		hrmSalaryConfirmBillModel
	}

	customHrmSalaryConfirmBillModel struct {
		*defaultHrmSalaryConfirmBillModel
	}
)

// NewHrmSalaryConfirmBillModel returns a model for the database table.
func NewHrmSalaryConfirmBillModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmSalaryConfirmBillModel {
	return &customHrmSalaryConfirmBillModel{
		defaultHrmSalaryConfirmBillModel: newHrmSalaryConfirmBillModel(conn, op...),
	}
}
