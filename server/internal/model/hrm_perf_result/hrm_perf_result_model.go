package hrm_perf_result

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmPerfResultModel = (*customHrmPerfResultModel)(nil)

type (
	// HrmPerfResultModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmPerfResultModel.
	HrmPerfResultModel interface {
		hrmPerfResultModel
	}

	customHrmPerfResultModel struct {
		*defaultHrmPerfResultModel
	}
)

// NewHrmPerfResultModel returns a model for the database table.
func NewHrmPerfResultModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmPerfResultModel {
	return &customHrmPerfResultModel{
		defaultHrmPerfResultModel: newHrmPerfResultModel(conn, op...),
	}
}
