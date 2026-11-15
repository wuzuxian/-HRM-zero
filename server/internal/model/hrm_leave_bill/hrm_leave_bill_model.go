package hrm_leave_bill

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmLeaveBillModel = (*customHrmLeaveBillModel)(nil)

type (
	// HrmLeaveBillModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmLeaveBillModel.
	HrmLeaveBillModel interface {
		hrmLeaveBillModel
	}

	customHrmLeaveBillModel struct {
		*defaultHrmLeaveBillModel
	}
)

// NewHrmLeaveBillModel returns a model for the database table.
func NewHrmLeaveBillModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmLeaveBillModel {
	return &customHrmLeaveBillModel{
		defaultHrmLeaveBillModel: newHrmLeaveBillModel(conn, op...),
	}
}
