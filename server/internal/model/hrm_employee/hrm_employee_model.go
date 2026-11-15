package hrm_employee

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmEmployeeModel = (*customHrmEmployeeModel)(nil)

type (
	// HrmEmployeeModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmEmployeeModel.
	HrmEmployeeModel interface {
		hrmEmployeeModel
	}

	customHrmEmployeeModel struct {
		*defaultHrmEmployeeModel
	}
)

// NewHrmEmployeeModel returns a model for the database table.
func NewHrmEmployeeModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmEmployeeModel {
	return &customHrmEmployeeModel{
		defaultHrmEmployeeModel: newHrmEmployeeModel(conn, op...),
	}
}
