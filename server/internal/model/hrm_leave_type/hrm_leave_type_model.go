package hrm_leave_type

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmLeaveTypeModel = (*customHrmLeaveTypeModel)(nil)

type (
	// HrmLeaveTypeModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmLeaveTypeModel.
	HrmLeaveTypeModel interface {
		hrmLeaveTypeModel
	}

	customHrmLeaveTypeModel struct {
		*defaultHrmLeaveTypeModel
	}
)

// NewHrmLeaveTypeModel returns a model for the database table.
func NewHrmLeaveTypeModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmLeaveTypeModel {
	return &customHrmLeaveTypeModel{
		defaultHrmLeaveTypeModel: newHrmLeaveTypeModel(conn, op...),
	}
}
