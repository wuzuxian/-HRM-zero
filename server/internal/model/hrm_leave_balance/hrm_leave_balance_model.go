package hrm_leave_balance

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmLeaveBalanceModel = (*customHrmLeaveBalanceModel)(nil)

type (
	// HrmLeaveBalanceModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmLeaveBalanceModel.
	HrmLeaveBalanceModel interface {
		hrmLeaveBalanceModel
	}

	customHrmLeaveBalanceModel struct {
		*defaultHrmLeaveBalanceModel
	}
)

// NewHrmLeaveBalanceModel returns a model for the database table.
func NewHrmLeaveBalanceModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmLeaveBalanceModel {
	return &customHrmLeaveBalanceModel{
		defaultHrmLeaveBalanceModel: newHrmLeaveBalanceModel(conn, op...),
	}
}
