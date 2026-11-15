package hrm_regularization_bill

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmRegularizationBillModel = (*customHrmRegularizationBillModel)(nil)

type (
	// HrmRegularizationBillModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmRegularizationBillModel.
	HrmRegularizationBillModel interface {
		hrmRegularizationBillModel
	}

	customHrmRegularizationBillModel struct {
		*defaultHrmRegularizationBillModel
	}
)

// NewHrmRegularizationBillModel returns a model for the database table.
func NewHrmRegularizationBillModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmRegularizationBillModel {
	return &customHrmRegularizationBillModel{
		defaultHrmRegularizationBillModel: newHrmRegularizationBillModel(conn, op...),
	}
}
