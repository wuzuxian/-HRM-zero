package hrm_entry_bill_family

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmEntryBillFamilyModel = (*customHrmEntryBillFamilyModel)(nil)

type (
	// HrmEntryBillFamilyModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmEntryBillFamilyModel.
	HrmEntryBillFamilyModel interface {
		hrmEntryBillFamilyModel
	}

	customHrmEntryBillFamilyModel struct {
		*defaultHrmEntryBillFamilyModel
	}
)

// NewHrmEntryBillFamilyModel returns a model for the database table.
func NewHrmEntryBillFamilyModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmEntryBillFamilyModel {
	return &customHrmEntryBillFamilyModel{
		defaultHrmEntryBillFamilyModel: newHrmEntryBillFamilyModel(conn, op...),
	}
}
