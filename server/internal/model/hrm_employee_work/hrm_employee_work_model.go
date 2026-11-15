package hrm_employee_work

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmEmployeeWorkModel = (*customHrmEmployeeWorkModel)(nil)

type (
	// HrmEmployeeWorkModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmEmployeeWorkModel.
	HrmEmployeeWorkModel interface {
		hrmEmployeeWorkModel
	}

	customHrmEmployeeWorkModel struct {
		*defaultHrmEmployeeWorkModel
	}
)

// NewHrmEmployeeWorkModel returns a model for the database table.
func NewHrmEmployeeWorkModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmEmployeeWorkModel {
	return &customHrmEmployeeWorkModel{
		defaultHrmEmployeeWorkModel: newHrmEmployeeWorkModel(conn, op...),
	}
}
