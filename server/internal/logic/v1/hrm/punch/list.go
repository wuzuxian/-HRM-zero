package punch

// List 管理端「打卡记录」页（demo-inventory 4.4）：筛选 日期区间/姓名(模糊)/签到状态，
// 一人一天一行展示，附员工快照 + 日结结果列。
// 签到状态筛选: normal 正常 / field 外勤 / missing 缺卡（筛空状态行）。

import (
	"context"
	"net/http"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_org"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_punch_record"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/punch"
)

type List struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewList(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *List {
	return &List{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

func (l *List) List(req *types.ListRequest) (resp *types.ListResponse, err error) {
	chain := condition.NewChain().
		Page(req.Current, req.Size)

	// 签到状态筛选: normal/field 等值；missing 缺卡 = 状态为空。
	if req.ClockInStatus == SettleMissing {
		chain = chain.Equal(hrm_punch_record.ClockInStatus, "")
	} else {
		chain = chain.Equal(hrm_punch_record.ClockInStatus, req.ClockInStatus, condition.WithSkip(!isPunchStatus(req.ClockInStatus)))
	}

	// 日期区间
	if req.StartDate != "" {
		chain = chain.GreaterEqualThan(hrm_punch_record.PunchDate, req.StartDate)
	}
	if req.EndDate != "" {
		chain = chain.LessEqualThan(hrm_punch_record.PunchDate, req.EndDate)
	}

	// 姓名(模糊) → 员工 uuid 集合
	if req.Name != "" {
		employeeUuids, err := matchEmployeeUuids(l.ctx, l.svcCtx, req.Name)
		if err != nil {
			return nil, err
		}
		if len(employeeUuids) == 0 {
			return &types.ListResponse{
				PageResponse: types.PageResponse{Current: req.Current, Size: req.Size, Total: 0},
				Records:      []types.HrmPunchRecord{},
			}, nil
		}
		chain = chain.In(hrm_punch_record.EmployeeUuid, employeeUuids)
	}

	rows, total, err := l.svcCtx.Model.HrmPunchRecord.PageByCondition(l.ctx, nil,
		chain.OrderByDesc(hrm_punch_record.PunchDate).
			OrderByAsc(hrm_punch_record.EmployeeNo).
			Build()...)
	if err != nil {
		return nil, err
	}

	records := make([]types.HrmPunchRecord, 0, len(rows))
	for _, row := range rows {
		records = append(records, toRecord(l.ctx, l.svcCtx, row))
	}
	return &types.ListResponse{
		PageResponse: types.PageResponse{Current: req.Current, Size: req.Size, Total: total},
		Records:      records,
	}, nil
}

// isPunchStatus 判定是否为可直接等值筛选的签到状态。
func isPunchStatus(v string) bool {
	return v == StatusNormal || v == StatusField
}

// matchEmployeeUuids 姓名(模糊) → 员工 uuid 集合。
func matchEmployeeUuids(ctx context.Context, svcCtx *svc.ServiceContext, name string) ([]string, error) {
	rows, err := svcCtx.Model.HrmEmployee.FindByCondition(ctx, nil,
		condition.NewChain().Like(hrm_employee.Name, "%"+name+"%").Build()...)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.Uuid)
	}
	return out, nil
}

// toRecord 打卡行 → 列表行（附星期/部门名快照；空状态补成 missing 语义）。
func toRecord(ctx context.Context, svcCtx *svc.ServiceContext, row *hrm_punch_record.HrmPunchRecord) types.HrmPunchRecord {
	record := types.HrmPunchRecord{
		Uuid:               row.Uuid,
		PunchDate:          row.PunchDate,
		EmployeeUuid:       row.EmployeeUuid,
		UserUuid:           row.UserUuid,
		Username:           row.Username,
		EmployeeNo:         row.EmployeeNo,
		EmployeeName:       row.EmployeeName,
		Weekday:            weekdayName(row.PunchDate),
		ClockInTime:        formatNullTime(row.ClockInTime),
		ClockInStatus:      row.ClockInStatus,
		ClockInLocation:    row.ClockInLocation,
		ClockOutTime:       formatNullTime(row.ClockOutTime),
		ClockOutStatus:     row.ClockOutStatus,
		ClockOutLocation:   row.ClockOutLocation,
		Overtime:           row.Overtime,
		Source:             row.Source,
		RuleSource:         row.RuleSource,
		SettleStatus:       row.SettleStatus,
		SettleLateMinutes:  row.SettleLateMinutes,
		SettleEarlyMinutes: row.SettleEarlyMinutes,
		SettleWorkHours:    row.SettleWorkHours,
		SettleTime:         formatNullTime(row.SettleTime),
		CreateTime:         row.CreateTime.Format(time.DateTime),
		UpdateTime:         row.UpdateTime.Format(time.DateTime),
	}
	if row.EmployeeUuid != "" {
		employee, err := svcCtx.Model.HrmEmployee.FindOneByCondition(ctx, nil,
			condition.NewChain().Equal(hrm_employee.Uuid, row.EmployeeUuid).Build()...)
		if err == nil && employee != nil && employee.OrgUuid != "" {
			org, orgErr := svcCtx.Model.HrmOrg.FindOneByCondition(ctx, nil,
				condition.NewChain().Equal(hrm_org.Uuid, employee.OrgUuid).Build()...)
			if orgErr == nil && org != nil {
				record.OrgName = org.OrgName
			}
		}
	}
	if record.ClockInTime == "" && record.ClockInStatus == "" {
		record.ClockInStatus = SettleMissing
	}
	if record.ClockOutTime == "" && record.ClockOutStatus == "" {
		record.ClockOutStatus = SettleMissing
	}
	return record
}

// weekdayName yyyy-MM-dd → 周一..周日（非法日期返回空）。
func weekdayName(date string) string {
	t, err := time.ParseInLocation(DateLayout, date, time.Local)
	if err != nil {
		return ""
	}
	names := []string{"周日", "周一", "周二", "周三", "周四", "周五", "周六"}
	return names[int(t.Weekday())]
}
