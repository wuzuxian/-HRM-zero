package hrm_perf_scheme

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmPerfSchemeModel = (*customHrmPerfSchemeModel)(nil)

type (
	// HrmPerfSchemeModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmPerfSchemeModel.
	HrmPerfSchemeModel interface {
		hrmPerfSchemeModel
	}

	customHrmPerfSchemeModel struct {
		*defaultHrmPerfSchemeModel
	}
)

// NewHrmPerfSchemeModel returns a model for the database table.
func NewHrmPerfSchemeModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmPerfSchemeModel {
	return &customHrmPerfSchemeModel{
		defaultHrmPerfSchemeModel: newHrmPerfSchemeModel(conn, op...),
	}
}
