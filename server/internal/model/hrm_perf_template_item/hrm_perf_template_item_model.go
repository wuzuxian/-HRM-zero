package hrm_perf_template_item

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmPerfTemplateItemModel = (*customHrmPerfTemplateItemModel)(nil)

type (
	// HrmPerfTemplateItemModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmPerfTemplateItemModel.
	HrmPerfTemplateItemModel interface {
		hrmPerfTemplateItemModel
	}

	customHrmPerfTemplateItemModel struct {
		*defaultHrmPerfTemplateItemModel
	}
)

// NewHrmPerfTemplateItemModel returns a model for the database table.
func NewHrmPerfTemplateItemModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmPerfTemplateItemModel {
	return &customHrmPerfTemplateItemModel{
		defaultHrmPerfTemplateItemModel: newHrmPerfTemplateItemModel(conn, op...),
	}
}
