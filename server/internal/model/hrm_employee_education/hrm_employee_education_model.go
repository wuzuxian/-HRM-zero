package hrm_employee_education

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmEmployeeEducationModel = (*customHrmEmployeeEducationModel)(nil)

type (
	// HrmEmployeeEducationModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmEmployeeEducationModel.
	HrmEmployeeEducationModel interface {
		hrmEmployeeEducationModel
	}

	customHrmEmployeeEducationModel struct {
		*defaultHrmEmployeeEducationModel
	}
)

// NewHrmEmployeeEducationModel returns a model for the database table.
func NewHrmEmployeeEducationModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmEmployeeEducationModel {
	return &customHrmEmployeeEducationModel{
		defaultHrmEmployeeEducationModel: newHrmEmployeeEducationModel(conn, op...),
	}
}
