package hrm_perf_period

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmPerfPeriodModel = (*customHrmPerfPeriodModel)(nil)

type (
	// HrmPerfPeriodModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmPerfPeriodModel.
	HrmPerfPeriodModel interface {
		hrmPerfPeriodModel
	}

	customHrmPerfPeriodModel struct {
		*defaultHrmPerfPeriodModel
	}
)

// NewHrmPerfPeriodModel returns a model for the database table.
func NewHrmPerfPeriodModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmPerfPeriodModel {
	return &customHrmPerfPeriodModel{
		defaultHrmPerfPeriodModel: newHrmPerfPeriodModel(conn, op...),
	}
}
