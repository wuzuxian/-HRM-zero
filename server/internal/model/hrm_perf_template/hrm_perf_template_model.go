package hrm_perf_template

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmPerfTemplateModel = (*customHrmPerfTemplateModel)(nil)

type (
	// HrmPerfTemplateModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmPerfTemplateModel.
	HrmPerfTemplateModel interface {
		hrmPerfTemplateModel
	}

	customHrmPerfTemplateModel struct {
		*defaultHrmPerfTemplateModel
	}
)

// NewHrmPerfTemplateModel returns a model for the database table.
func NewHrmPerfTemplateModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmPerfTemplateModel {
	return &customHrmPerfTemplateModel{
		defaultHrmPerfTemplateModel: newHrmPerfTemplateModel(conn, op...),
	}
}
