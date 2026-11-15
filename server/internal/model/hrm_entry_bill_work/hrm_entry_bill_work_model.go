package hrm_entry_bill_work

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmEntryBillWorkModel = (*customHrmEntryBillWorkModel)(nil)

type (
	// HrmEntryBillWorkModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmEntryBillWorkModel.
	HrmEntryBillWorkModel interface {
		hrmEntryBillWorkModel
	}

	customHrmEntryBillWorkModel struct {
		*defaultHrmEntryBillWorkModel
	}
)

// NewHrmEntryBillWorkModel returns a model for the database table.
func NewHrmEntryBillWorkModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmEntryBillWorkModel {
	return &customHrmEntryBillWorkModel{
		defaultHrmEntryBillWorkModel: newHrmEntryBillWorkModel(conn, op...),
	}
}
