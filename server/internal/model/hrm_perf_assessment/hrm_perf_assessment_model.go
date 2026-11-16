package hrm_perf_assessment

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmPerfAssessmentModel = (*customHrmPerfAssessmentModel)(nil)

type (
	// HrmPerfAssessmentModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmPerfAssessmentModel.
	HrmPerfAssessmentModel interface {
		hrmPerfAssessmentModel
	}

	customHrmPerfAssessmentModel struct {
		*defaultHrmPerfAssessmentModel
	}
)

// NewHrmPerfAssessmentModel returns a model for the database table.
func NewHrmPerfAssessmentModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmPerfAssessmentModel {
	return &customHrmPerfAssessmentModel{
		defaultHrmPerfAssessmentModel: newHrmPerfAssessmentModel(conn, op...),
	}
}
