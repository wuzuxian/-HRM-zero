package hrm_entry_bill_education

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmEntryBillEducationModel = (*customHrmEntryBillEducationModel)(nil)

type (
	// HrmEntryBillEducationModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmEntryBillEducationModel.
	HrmEntryBillEducationModel interface {
		hrmEntryBillEducationModel
	}

	customHrmEntryBillEducationModel struct {
		*defaultHrmEntryBillEducationModel
	}
)

// NewHrmEntryBillEducationModel returns a model for the database table.
func NewHrmEntryBillEducationModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmEntryBillEducationModel {
	return &customHrmEntryBillEducationModel{
		defaultHrmEntryBillEducationModel: newHrmEntryBillEducationModel(conn, op...),
	}
}
