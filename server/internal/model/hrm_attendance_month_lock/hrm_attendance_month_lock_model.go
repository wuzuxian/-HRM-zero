package hrm_attendance_month_lock

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmAttendanceMonthLockModel = (*customHrmAttendanceMonthLockModel)(nil)

type (
	// HrmAttendanceMonthLockModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmAttendanceMonthLockModel.
	HrmAttendanceMonthLockModel interface {
		hrmAttendanceMonthLockModel
	}

	customHrmAttendanceMonthLockModel struct {
		*defaultHrmAttendanceMonthLockModel
	}
)

// NewHrmAttendanceMonthLockModel returns a model for the database table.
func NewHrmAttendanceMonthLockModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmAttendanceMonthLockModel {
	return &customHrmAttendanceMonthLockModel{
		defaultHrmAttendanceMonthLockModel: newHrmAttendanceMonthLockModel(conn, op...),
	}
}
