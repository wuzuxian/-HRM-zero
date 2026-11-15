package hrm_transfer_bill

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmTransferBillModel = (*customHrmTransferBillModel)(nil)

type (
	// HrmTransferBillModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmTransferBillModel.
	HrmTransferBillModel interface {
		hrmTransferBillModel
	}

	customHrmTransferBillModel struct {
		*defaultHrmTransferBillModel
	}
)

// NewHrmTransferBillModel returns a model for the database table.
func NewHrmTransferBillModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmTransferBillModel {
	return &customHrmTransferBillModel{
		defaultHrmTransferBillModel: newHrmTransferBillModel(conn, op...),
	}
}
