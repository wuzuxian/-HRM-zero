package hrm_card_replacement_bill

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmCardReplacementBillModel = (*customHrmCardReplacementBillModel)(nil)

type (
	// HrmCardReplacementBillModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmCardReplacementBillModel.
	HrmCardReplacementBillModel interface {
		hrmCardReplacementBillModel
	}

	customHrmCardReplacementBillModel struct {
		*defaultHrmCardReplacementBillModel
	}
)

// NewHrmCardReplacementBillModel returns a model for the database table.
func NewHrmCardReplacementBillModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmCardReplacementBillModel {
	return &customHrmCardReplacementBillModel{
		defaultHrmCardReplacementBillModel: newHrmCardReplacementBillModel(conn, op...),
	}
}
