package hrm_schedule_entry

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmScheduleEntryModel = (*customHrmScheduleEntryModel)(nil)

type (
	// HrmScheduleEntryModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmScheduleEntryModel.
	HrmScheduleEntryModel interface {
		hrmScheduleEntryModel
	}

	customHrmScheduleEntryModel struct {
		*defaultHrmScheduleEntryModel
	}
)

// NewHrmScheduleEntryModel returns a model for the database table.
func NewHrmScheduleEntryModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmScheduleEntryModel {
	return &customHrmScheduleEntryModel{
		defaultHrmScheduleEntryModel: newHrmScheduleEntryModel(conn, op...),
	}
}
