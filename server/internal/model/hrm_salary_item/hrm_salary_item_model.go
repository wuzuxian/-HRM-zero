package hrm_salary_item

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmSalaryItemModel = (*customHrmSalaryItemModel)(nil)

type (
	// HrmSalaryItemModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmSalaryItemModel.
	HrmSalaryItemModel interface {
		hrmSalaryItemModel
	}

	customHrmSalaryItemModel struct {
		*defaultHrmSalaryItemModel
	}
)

// NewHrmSalaryItemModel returns a model for the database table.
func NewHrmSalaryItemModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmSalaryItemModel {
	return &customHrmSalaryItemModel{
		defaultHrmSalaryItemModel: newHrmSalaryItemModel(conn, op...),
	}
}
