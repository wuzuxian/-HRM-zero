package hrm_salary_batch_item

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmSalaryBatchItemModel = (*customHrmSalaryBatchItemModel)(nil)

type (
	// HrmSalaryBatchItemModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmSalaryBatchItemModel.
	HrmSalaryBatchItemModel interface {
		hrmSalaryBatchItemModel
	}

	customHrmSalaryBatchItemModel struct {
		*defaultHrmSalaryBatchItemModel
	}
)

// NewHrmSalaryBatchItemModel returns a model for the database table.
func NewHrmSalaryBatchItemModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmSalaryBatchItemModel {
	return &customHrmSalaryBatchItemModel{
		defaultHrmSalaryBatchItemModel: newHrmSalaryBatchItemModel(conn, op...),
	}
}
