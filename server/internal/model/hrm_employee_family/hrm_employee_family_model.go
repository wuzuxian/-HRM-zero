package hrm_employee_family

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmEmployeeFamilyModel = (*customHrmEmployeeFamilyModel)(nil)

type (
	// HrmEmployeeFamilyModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmEmployeeFamilyModel.
	HrmEmployeeFamilyModel interface {
		hrmEmployeeFamilyModel
	}

	customHrmEmployeeFamilyModel struct {
		*defaultHrmEmployeeFamilyModel
	}
)

// NewHrmEmployeeFamilyModel returns a model for the database table.
func NewHrmEmployeeFamilyModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmEmployeeFamilyModel {
	return &customHrmEmployeeFamilyModel{
		defaultHrmEmployeeFamilyModel: newHrmEmployeeFamilyModel(conn, op...),
	}
}
