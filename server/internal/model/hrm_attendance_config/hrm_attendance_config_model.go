package hrm_attendance_config

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmAttendanceConfigModel = (*customHrmAttendanceConfigModel)(nil)

type (
	// HrmAttendanceConfigModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmAttendanceConfigModel.
	HrmAttendanceConfigModel interface {
		hrmAttendanceConfigModel
	}

	customHrmAttendanceConfigModel struct {
		*defaultHrmAttendanceConfigModel
	}
)

// NewHrmAttendanceConfigModel returns a model for the database table.
func NewHrmAttendanceConfigModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmAttendanceConfigModel {
	return &customHrmAttendanceConfigModel{
		defaultHrmAttendanceConfigModel: newHrmAttendanceConfigModel(conn, op...),
	}
}
