package hrm_salary_subject

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmSalarySubjectModel = (*customHrmSalarySubjectModel)(nil)

type (
	// HrmSalarySubjectModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmSalarySubjectModel.
	HrmSalarySubjectModel interface {
		hrmSalarySubjectModel
	}

	customHrmSalarySubjectModel struct {
		*defaultHrmSalarySubjectModel
	}
)

// NewHrmSalarySubjectModel returns a model for the database table.
func NewHrmSalarySubjectModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmSalarySubjectModel {
	return &customHrmSalarySubjectModel{
		defaultHrmSalarySubjectModel: newHrmSalarySubjectModel(conn, op...),
	}
}
