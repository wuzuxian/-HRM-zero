package hrm_offer_bill

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmOfferBillModel = (*customHrmOfferBillModel)(nil)

type (
	// HrmOfferBillModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmOfferBillModel.
	HrmOfferBillModel interface {
		hrmOfferBillModel
	}

	customHrmOfferBillModel struct {
		*defaultHrmOfferBillModel
	}
)

// NewHrmOfferBillModel returns a model for the database table.
func NewHrmOfferBillModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmOfferBillModel {
	return &customHrmOfferBillModel{
		defaultHrmOfferBillModel: newHrmOfferBillModel(conn, op...),
	}
}
