package hrm_salary_archive

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmSalaryArchiveModel = (*customHrmSalaryArchiveModel)(nil)

type (
	// HrmSalaryArchiveModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmSalaryArchiveModel.
	HrmSalaryArchiveModel interface {
		hrmSalaryArchiveModel
	}

	customHrmSalaryArchiveModel struct {
		*defaultHrmSalaryArchiveModel
	}
)

// NewHrmSalaryArchiveModel returns a model for the database table.
func NewHrmSalaryArchiveModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmSalaryArchiveModel {
	return &customHrmSalaryArchiveModel{
		defaultHrmSalaryArchiveModel: newHrmSalaryArchiveModel(conn, op...),
	}
}
