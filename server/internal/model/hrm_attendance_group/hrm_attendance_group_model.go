package hrm_attendance_group

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmAttendanceGroupModel = (*customHrmAttendanceGroupModel)(nil)

type (
	// HrmAttendanceGroupModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmAttendanceGroupModel.
	HrmAttendanceGroupModel interface {
		hrmAttendanceGroupModel
	}

	customHrmAttendanceGroupModel struct {
		*defaultHrmAttendanceGroupModel
	}
)

// NewHrmAttendanceGroupModel returns a model for the database table.
func NewHrmAttendanceGroupModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmAttendanceGroupModel {
	return &customHrmAttendanceGroupModel{
		defaultHrmAttendanceGroupModel: newHrmAttendanceGroupModel(conn, op...),
	}
}
