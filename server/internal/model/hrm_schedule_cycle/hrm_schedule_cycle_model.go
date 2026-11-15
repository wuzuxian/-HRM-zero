package hrm_schedule_cycle

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmScheduleCycleModel = (*customHrmScheduleCycleModel)(nil)

type (
	// HrmScheduleCycleModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmScheduleCycleModel.
	HrmScheduleCycleModel interface {
		hrmScheduleCycleModel
	}

	customHrmScheduleCycleModel struct {
		*defaultHrmScheduleCycleModel
	}
)

// NewHrmScheduleCycleModel returns a model for the database table.
func NewHrmScheduleCycleModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmScheduleCycleModel {
	return &customHrmScheduleCycleModel{
		defaultHrmScheduleCycleModel: newHrmScheduleCycleModel(conn, op...),
	}
}
