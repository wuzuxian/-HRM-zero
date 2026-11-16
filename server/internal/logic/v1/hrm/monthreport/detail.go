package monthreport

// Detail 行操作「日明细」：某员工当月逐日 settle 详情（复用打卡记录数据，HRM-22 明细模式）。

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/punch"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_punch_record"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/monthreport"
)

type Detail struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewDetail(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Detail {
	return &Detail{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

func (l *Detail) Detail(req *types.DetailRequest) (resp *types.DetailResponse, err error) {
	yearMonth, err := validateYearMonth(req.YearMonth)
	if err != nil {
		return nil, statusError(errcodes.MonthReportInvalidCode, err)
	}
	startDate := firstDayOfMonth(yearMonth)
	endDate := lastDayOfMonth(yearMonth)

	rows, err := l.svcCtx.Model.HrmPunchRecord.FindByCondition(l.ctx, nil,
		condition.NewChain().
			Equal(hrm_punch_record.EmployeeUuid, req.EmployeeUuid).
			GreaterEqualThan(hrm_punch_record.PunchDate, startDate).
			LessEqualThan(hrm_punch_record.PunchDate, endDate).
			OrderByAsc(hrm_punch_record.PunchDate).Build()...)
	if err != nil {
		return nil, err
	}

	records := make([]types.HrmPunchDetailDay, 0, len(rows))
	for _, row := range rows {
		records = append(records, types.HrmPunchDetailDay{
			PunchDate:          row.PunchDate,
			Weekday:            weekdayCN(row.PunchDate),
			SettleStatus:       row.SettleStatus,
			ClockInTime:        formatNullTime(row.ClockInTime),
			ClockOutTime:       formatNullTime(row.ClockOutTime),
			SettleLateMinutes:  row.SettleLateMinutes,
			SettleEarlyMinutes: row.SettleEarlyMinutes,
			SettleWorkHours:    row.SettleWorkHours,
		})
	}
	return &types.DetailResponse{Records: records}, nil
}

// weekdayCN yyyy-MM-dd → 周一..周日（与 punch.weekdayName 同口径；未导出故本地实现）。
func weekdayCN(date string) string {
	t, err := time.ParseInLocation(punch.DateLayout, date, time.Local)
	if err != nil {
		return ""
	}
	names := []string{"周日", "周一", "周二", "周三", "周四", "周五", "周六"}
	return names[int(t.Weekday())]
}

// formatNullTime datetime 列 → "yyyy-MM-dd HH:mm:ss"（NULL → ""；punch.formatNullTime 未导出）。
func formatNullTime(nt sql.NullTime) string {
	if !nt.Valid {
		return ""
	}
	return nt.Time.Format(time.DateTime)
}
