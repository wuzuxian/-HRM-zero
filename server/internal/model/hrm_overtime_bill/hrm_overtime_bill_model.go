package hrm_overtime_bill

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmOvertimeBillModel = (*customHrmOvertimeBillModel)(nil)

type (
	// HrmOvertimeBillModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmOvertimeBillModel.
	HrmOvertimeBillModel interface {
		hrmOvertimeBillModel
	}

	customHrmOvertimeBillModel struct {
		*defaultHrmOvertimeBillModel
	}
)

// NewHrmOvertimeBillModel returns a model for the database table.
func NewHrmOvertimeBillModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmOvertimeBillModel {
	return &customHrmOvertimeBillModel{
		defaultHrmOvertimeBillModel: newHrmOvertimeBillModel(conn, op...),
	}
}
