package hrm_perf_appeal

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmPerfAppealModel = (*customHrmPerfAppealModel)(nil)

type (
	// HrmPerfAppealModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmPerfAppealModel.
	HrmPerfAppealModel interface {
		hrmPerfAppealModel
	}

	customHrmPerfAppealModel struct {
		*defaultHrmPerfAppealModel
	}
)

// NewHrmPerfAppealModel returns a model for the database table.
func NewHrmPerfAppealModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmPerfAppealModel {
	return &customHrmPerfAppealModel{
		defaultHrmPerfAppealModel: newHrmPerfAppealModel(conn, op...),
	}
}
