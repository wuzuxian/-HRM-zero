package hrm_salary_scheme

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmSalarySchemeModel = (*customHrmSalarySchemeModel)(nil)

type (
	// HrmSalarySchemeModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmSalarySchemeModel.
	HrmSalarySchemeModel interface {
		hrmSalarySchemeModel
	}

	customHrmSalarySchemeModel struct {
		*defaultHrmSalarySchemeModel
	}
)

// NewHrmSalarySchemeModel returns a model for the database table.
func NewHrmSalarySchemeModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmSalarySchemeModel {
	return &customHrmSalarySchemeModel{
		defaultHrmSalarySchemeModel: newHrmSalarySchemeModel(conn, op...),
	}
}
