package hrm_salary_batch

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmSalaryBatchModel = (*customHrmSalaryBatchModel)(nil)

type (
	// HrmSalaryBatchModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmSalaryBatchModel.
	HrmSalaryBatchModel interface {
		hrmSalaryBatchModel
	}

	customHrmSalaryBatchModel struct {
		*defaultHrmSalaryBatchModel
	}
)

// NewHrmSalaryBatchModel returns a model for the database table.
func NewHrmSalaryBatchModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmSalaryBatchModel {
	return &customHrmSalaryBatchModel{
		defaultHrmSalaryBatchModel: newHrmSalaryBatchModel(conn, op...),
	}
}
