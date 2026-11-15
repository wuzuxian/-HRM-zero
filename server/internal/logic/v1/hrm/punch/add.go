package punch

// Add 管理端手动新增打卡记录（demo-inventory 4.4 工具栏「新增」，补录外部数据单行）：
// (punch_date, employee_uuid) 唯一 → 重复新增 10523；新增后即时日结。

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/jzero-io/jzero/core/status"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_punch_record"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/punch"
)

type Add struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewAdd(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Add {
	return &Add{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

func (l *Add) Add(req *types.AddRequest) (resp *types.AddResponse, err error) {
	if req.EmployeeUuid == "" || req.PunchDate == "" {
		return nil, status.ErrorMessage(errcodes.PunchInvalidCode, "员工与打卡日期必填")
	}
	if _, err := time.ParseInLocation(DateLayout, req.PunchDate, time.Local); err != nil {
		return nil, status.ErrorMessage(errcodes.PunchInvalidCode, "打卡日期必须为 yyyy-MM-dd")
	}
	if err := validatePunchTimes(req.ClockInTime, req.ClockOutTime); err != nil {
		return nil, err
	}

	employee, err := l.svcCtx.Model.HrmEmployee.FindOneByCondition(l.ctx, nil,
		condition.NewChain().Equal(hrm_employee.Uuid, req.EmployeeUuid).Build()...)
	if err != nil {
		if err == hrm_employee.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.PunchEmployeeNotFoundCode, "员工不存在: "+req.EmployeeUuid)
		}
		return nil, err
	}

	if existing, findErr := findPunchRow(l.ctx, l.svcCtx, req.EmployeeUuid, req.PunchDate); findErr != nil {
		return nil, findErr
	} else if existing != nil {
		return nil, status.ErrorMessage(errcodes.PunchDuplicateCode, "该员工当日已有打卡记录（一人一天一行）")
	}

	row := &hrm_punch_record.HrmPunchRecord{
		Uuid:         newUuid(),
		PunchDate:    req.PunchDate,
		EmployeeUuid: req.EmployeeUuid,
		UserUuid:     employee.UserUuid,
		Username:     usernameOf(l.ctx, l.svcCtx, employee.UserUuid),
		EmployeeNo:   employee.EmployeeNo,
		EmployeeName: employee.Name,
		ClockInStatus:  req.ClockInStatus,
		ClockInLocation: req.ClockInLocation,
		ClockOutStatus: req.ClockOutStatus,
		ClockOutLocation: req.ClockOutLocation,
		Overtime:   normalizeOvertime(req.Overtime),
		Source:     SourceAdmin,
	}
	if req.ClockInTime != "" {
		tm, parseErr := time.ParseInLocation(time.DateTime, req.ClockInTime, time.Local)
		if parseErr != nil {
			return nil, status.ErrorMessage(errcodes.PunchInvalidCode, "签到时间必须为 yyyy-MM-dd HH:mm:ss")
		}
		row.ClockInTime = sql.NullTime{Time: tm, Valid: true}
	}
	if req.ClockOutTime != "" {
		tm, parseErr := time.ParseInLocation(time.DateTime, req.ClockOutTime, time.Local)
		if parseErr != nil {
			return nil, status.ErrorMessage(errcodes.PunchInvalidCode, "签退时间必须为 yyyy-MM-dd HH:mm:ss")
		}
		row.ClockOutTime = sql.NullTime{Time: tm, Valid: true}
	}
	if _, err := l.svcCtx.Model.HrmPunchRecord.Insert(l.ctx, nil, row); err != nil {
		return nil, err
	}

	// 新增后即时日结（状态快照固化）
	if _, err := SettleDay(l.ctx, l.svcCtx, req.EmployeeUuid, req.PunchDate); err != nil {
		return nil, err
	}
	return &types.AddResponse{}, nil
}

// validatePunchTimes 管理端时间格式校验（空=缺卡合法）。
func validatePunchTimes(clockIn, clockOut string) error {
	for _, v := range []string{clockIn, clockOut} {
		if v == "" {
			continue
		}
		if _, err := time.ParseInLocation(time.DateTime, v, time.Local); err != nil {
			return status.ErrorMessage(errcodes.PunchInvalidCode, "打卡时间必须为 yyyy-MM-dd HH:mm:ss: "+v)
		}
	}
	return nil
}

func normalizeOvertime(v string) string {
	if v == OvertimeYes {
		return OvertimeYes
	}
	return OvertimeNo
}
