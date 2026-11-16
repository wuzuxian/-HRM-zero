package hrm_attendance_month_report

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmAttendanceMonthReportModel = (*customHrmAttendanceMonthReportModel)(nil)

type (
	// HrmAttendanceMonthReportModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmAttendanceMonthReportModel.
	HrmAttendanceMonthReportModel interface {
		hrmAttendanceMonthReportModel
	}

	customHrmAttendanceMonthReportModel struct {
		*defaultHrmAttendanceMonthReportModel
	}
)

// NewHrmAttendanceMonthReportModel returns a model for the database table.
func NewHrmAttendanceMonthReportModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmAttendanceMonthReportModel {
	return &customHrmAttendanceMonthReportModel{
		defaultHrmAttendanceMonthReportModel: newHrmAttendanceMonthReportModel(conn, op...),
	}
}
