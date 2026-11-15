package punch

// Punch 端点（员工上/下班打卡，HRM-16 issue 正文）：
//   - 写入 hrm_punch_record（一人一天一行）：签到/签退时间、签到状态 正常/外勤、签到地点（GPS 地址文本）、打卡日期；
//   - 幂等策略（拒绝）：同日同类型已有卡 → 10543 拒绝；同日另一类型未打则补写对应列；
//   - 跨天班（下班时间 ≤ 上班时间）下班卡归属上班日（punch_date = 上班日）；
//   - 打卡后即时日结（复用 SettleDay），返回当前日结状态。

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/jzero-io/jzero/core/status"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/attendancegroup"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/hrmuser"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_punch_record"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/punch"
)

type Punch struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewPunch(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Punch {
	return &Punch{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

func (l *Punch) Punch(req *types.PunchRequest) (resp *types.PunchResponse, err error) {
	// 打卡人: 管理端代打卡传 employeeUuid，员工端缺省取当前登录人对应的员工档案。
	employeeUuid := req.EmployeeUuid
	if employeeUuid == "" {
		user := hrmuser.Current(l.ctx)
		if user.Uuid == "" {
			return nil, status.ErrorMessage(errcodes.PunchEmployeeNotFoundCode, "当前登录人未关联员工档案")
		}
		employee, findErr := l.svcCtx.Model.HrmEmployee.FindOneByCondition(l.ctx, nil,
			condition.NewChain().Equal(hrm_employee.UserUuid, user.Uuid).Build()...)
		if findErr != nil {
			if findErr == hrm_employee.ErrNotFound {
				return nil, status.ErrorMessage(errcodes.PunchEmployeeNotFoundCode, "当前登录人未关联员工档案")
			}
			return nil, findErr
		}
		employeeUuid = employee.Uuid
	}

	rec, err := punchAt(l.ctx, l.svcCtx, employeeUuid, req.Type, req.Location, time.Now())
	if err != nil {
		return nil, err
	}

	result, err := SettleDay(l.ctx, l.svcCtx, employeeUuid, rec.PunchDate)
	if err != nil {
		return nil, err
	}

	return &types.PunchResponse{
		PunchDate:    rec.PunchDate,
		ClockInTime:  formatNullTime(rec.ClockInTime),
		ClockOutTime: formatNullTime(rec.ClockOutTime),
		SettleStatus: result.Status,
	}, nil
}

// punchAt 打卡核心（可注入时间，测试与端点共用）：
// 解析打卡归属日（跨天班下班卡归属上班日）→ 幂等检查 → 写入/更新打卡行。
func punchAt(ctx context.Context, svcCtx *svc.ServiceContext, employeeUuid, punchType, location string, at time.Time) (*hrm_punch_record.HrmPunchRecord, error) {
	if punchType != PunchTypeClockIn && punchType != PunchTypeClockOut {
		return nil, status.ErrorMessage(errcodes.PunchInvalidCode, "打卡类型必须为 clockIn（上班卡）或 clockOut（下班卡）")
	}

	employee, err := svcCtx.Model.HrmEmployee.FindOneByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_employee.Uuid, employeeUuid).Build()...)
	if err != nil {
		if err == hrm_employee.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.PunchEmployeeNotFoundCode, "员工不存在: "+employeeUuid)
		}
		return nil, err
	}

	punchDate, err := resolvePunchDate(ctx, svcCtx, employee, punchType, at)
	if err != nil {
		return nil, err
	}

	rec, err := findPunchRow(ctx, svcCtx, employeeUuid, punchDate)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		rec = &hrm_punch_record.HrmPunchRecord{
			Uuid:         newUuid(),
			PunchDate:    punchDate,
			EmployeeUuid: employeeUuid,
			UserUuid:     employee.UserUuid,
			Username:     usernameOf(ctx, svcCtx, employee.UserUuid),
			EmployeeNo:   employee.EmployeeNo,
			EmployeeName: employee.Name,
			Source:       SourceApp,
		}
		if _, err := svcCtx.Model.HrmPunchRecord.Insert(ctx, nil, rec); err != nil {
			return nil, err
		}
	}

	fields := map[string]any{}
	if punchType == PunchTypeClockIn {
		if rec.ClockInTime.Valid || rec.ClockInStatus != "" {
			return nil, status.ErrorMessage(errcodes.PunchDuplicateCode, "今日上班卡已打过（幂等：同日重复上班卡拒绝）")
		}
		fields["clock_in_time"] = at
		fields["clock_in_status"] = punchStatus(location)
		fields["clock_in_location"] = location
	} else {
		if rec.ClockOutTime.Valid || rec.ClockOutStatus != "" {
			return nil, status.ErrorMessage(errcodes.PunchDuplicateCode, "今日下班卡已打过（幂等：同日重复下班卡拒绝）")
		}
		fields["clock_out_time"] = at
		fields["clock_out_status"] = punchStatus(location)
		fields["clock_out_location"] = location
	}
	if err := svcCtx.Model.HrmPunchRecord.UpdateFieldsByCondition(ctx, nil, fields,
		condition.NewChain().Equal(hrm_punch_record.Uuid, rec.Uuid).Build()...); err != nil {
		return nil, err
	}
	return findPunchRow(ctx, svcCtx, employeeUuid, punchDate)
}

// resolvePunchDate 打卡归属日: 跨天班（下班时间 ≤ 上班时间）的下班卡归属上班日。
func resolvePunchDate(ctx context.Context, svcCtx *svc.ServiceContext, employee *hrm_employee.HrmEmployee, punchType string, at time.Time) (string, error) {
	date := at.Format(DateLayout)
	if punchType != PunchTypeClockOut {
		return date, nil
	}
	rule, err := attendancegroup.ResolveEffectiveRule(ctx, svcCtx, employee)
	if err != nil {
		return "", err
	}
	if rule == nil || !rule.Found || rule.StartTime == "" || rule.EndTime == "" {
		return date, nil
	}
	startMinutes, err := parseHHmm(rule.StartTime)
	if err != nil {
		return date, nil
	}
	endMinutes, err := parseHHmm(rule.EndTime)
	if err != nil {
		return date, nil
	}
	if endMinutes <= startMinutes {
		// 跨天班: 下班卡在次日凌晨 → 归属上班日。判定: 打卡时刻相对打卡日零点的
		// 分钟数 < 规则上班分钟（06:10=370 < 22:00=1320）说明还在班次凌晨段，
		// 下班卡回退到前一天；深夜 23:59（1439 > 1320）保持当天。
		minutesOfDay := int64(at.Hour()*60 + at.Minute())
		if minutesOfDay < startMinutes {
			date = at.AddDate(0, 0, -1).Format(DateLayout)
		}
	}
	return date, nil
}

// punchStatus 签到状态: 带地点视为外勤打卡，否则正常。
func punchStatus(location string) string {
	if location != "" {
		return StatusField
	}
	return StatusNormal
}

// formatNullTime datetime 列 → "yyyy-MM-dd HH:mm:ss"（NULL → ""）。
func formatNullTime(nt sql.NullTime) string {
	if !nt.Valid {
		return ""
	}
	return nt.Time.Format(time.DateTime)
}
