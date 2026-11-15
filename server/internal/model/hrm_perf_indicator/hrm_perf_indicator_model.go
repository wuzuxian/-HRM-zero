package hrm_perf_indicator

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmPerfIndicatorModel = (*customHrmPerfIndicatorModel)(nil)

type (
	// HrmPerfIndicatorModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmPerfIndicatorModel.
	HrmPerfIndicatorModel interface {
		hrmPerfIndicatorModel
	}

	customHrmPerfIndicatorModel struct {
		*defaultHrmPerfIndicatorModel
	}
)

// NewHrmPerfIndicatorModel returns a model for the database table.
func NewHrmPerfIndicatorModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmPerfIndicatorModel {
	return &customHrmPerfIndicatorModel{
		defaultHrmPerfIndicatorModel: newHrmPerfIndicatorModel(conn, op...),
	}
}
