package hrm_entry_bill

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmEntryBillModel = (*customHrmEntryBillModel)(nil)

type (
	// HrmEntryBillModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmEntryBillModel.
	HrmEntryBillModel interface {
		hrmEntryBillModel
	}

	customHrmEntryBillModel struct {
		*defaultHrmEntryBillModel
	}
)

// NewHrmEntryBillModel returns a model for the database table.
func NewHrmEntryBillModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmEntryBillModel {
	return &customHrmEntryBillModel{
		defaultHrmEntryBillModel: newHrmEntryBillModel(conn, op...),
	}
}
