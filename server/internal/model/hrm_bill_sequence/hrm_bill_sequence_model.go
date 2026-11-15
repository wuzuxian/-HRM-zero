package hrm_bill_sequence

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmBillSequenceModel = (*customHrmBillSequenceModel)(nil)

type (
	// HrmBillSequenceModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmBillSequenceModel.
	HrmBillSequenceModel interface {
		hrmBillSequenceModel
	}

	customHrmBillSequenceModel struct {
		*defaultHrmBillSequenceModel
	}
)

// NewHrmBillSequenceModel returns a model for the database table.
func NewHrmBillSequenceModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmBillSequenceModel {
	return &customHrmBillSequenceModel{
		defaultHrmBillSequenceModel: newHrmBillSequenceModel(conn, op...),
	}
}
