package hrm_leave_balance_log

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmLeaveBalanceLogModel = (*customHrmLeaveBalanceLogModel)(nil)

type (
	// HrmLeaveBalanceLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmLeaveBalanceLogModel.
	HrmLeaveBalanceLogModel interface {
		hrmLeaveBalanceLogModel
	}

	customHrmLeaveBalanceLogModel struct {
		*defaultHrmLeaveBalanceLogModel
	}
)

// NewHrmLeaveBalanceLogModel returns a model for the database table.
func NewHrmLeaveBalanceLogModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmLeaveBalanceLogModel {
	return &customHrmLeaveBalanceLogModel{
		defaultHrmLeaveBalanceLogModel: newHrmLeaveBalanceLogModel(conn, op...),
	}
}
