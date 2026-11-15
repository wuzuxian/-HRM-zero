package hrm_salary_scheme_item

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmSalarySchemeItemModel = (*customHrmSalarySchemeItemModel)(nil)

type (
	// HrmSalarySchemeItemModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmSalarySchemeItemModel.
	HrmSalarySchemeItemModel interface {
		hrmSalarySchemeItemModel
	}

	customHrmSalarySchemeItemModel struct {
		*defaultHrmSalarySchemeItemModel
	}
)

// NewHrmSalarySchemeItemModel returns a model for the database table.
func NewHrmSalarySchemeItemModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmSalarySchemeItemModel {
	return &customHrmSalarySchemeItemModel{
		defaultHrmSalarySchemeItemModel: newHrmSalarySchemeItemModel(conn, op...),
	}
}
