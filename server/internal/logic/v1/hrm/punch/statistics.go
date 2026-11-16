package punch

// Statistics 考勤统计（HRM-22 考勤统计页，demo-inventory 4.4）：
//   - 统计卡：统计人数/异常人数/迟到次数/早退次数/缺卡次数/记录条数 —— 从 hrm_punch_record
//     settle_* 快照列按日期区间聚合（HRM-16 落库契约：日结结果直接落打卡行）。
//   - 按人明细：应出勤天（按生效规则 WorkDays 口径）/出勤天/缺勤天/迟到/早退/缺卡/
//     异常合计/异常摘要；筛选 日期区间/用户姓名/仅看异常。
//   - 行操作「查看明细」走 StatisticsDetail（复用打卡记录 toRecord 数据）。
//
// 口径（GLOSSARY「考勤统计」）：
//   - 异常 = settle_status ∈ {late, early, missing}；rest/no_rule 视为「未判定」，不计异常。
//   - 出勤天 = settle_status ∈ {normal, late, early} 的天数（迟到/早退也算出勤，同时计异常）。
//   - 缺勤天 = settle_status=missing 且上/下班卡均无打卡的天数（全天未出勤）；
//     只缺单侧卡的算「缺卡次数」不算缺勤天。
//   - 应出勤天 = 区间内工作日数（按三级回退生效规则 WorkDays；无规则兜底按周一至周五）。
//   - 日期区间缺省为当月（1 号至月末）；统计卡按筛选后的员工集合聚合。

import (
	"context"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/attendancegroup"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_org"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_punch_record"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/punch"
)

type Statistics struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewStatistics(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Statistics {
	return &Statistics{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// abnormalStatuses 异常状态集合（rest/no_rule 不算异常）。
var abnormalStatuses = map[string]bool{SettleLate: true, SettleEarly: true, SettleMissing: true}

// personAgg 单员工在区间内的打卡行聚合（Go 内存聚合，避免跨方言 SQL 聚合差异）。
type personAgg struct {
	statusCounts   map[string]int64 // settle_status → 天数
	absentDays     int64            // missing 且上/下班卡均无打卡
	totalWorkHours float64          // settle_work_hours 合计
}

func newPersonAgg() *personAgg {
	return &personAgg{statusCounts: map[string]int64{}}
}

func (l *Statistics) Statistics(req *types.StatisticsRequest) (resp *types.StatisticsResponse, err error) {
	startDate, endDate, rangeErr := resolveStatsRange(req.StartDate, req.EndDate)
	if rangeErr != nil {
		return nil, rangeErr
	}

	// 员工集合：姓名筛选（模糊）或全体有部门员工（与 Settle 默认口径一致）。
	var employees []*hrm_employee.HrmEmployee
	if req.Name != "" {
		employeeUuids, matchErr := matchEmployeeUuids(l.ctx, l.svcCtx, req.Name)
		if matchErr != nil {
			return nil, matchErr
		}
		if len(employeeUuids) == 0 {
			return emptyStatistics(req), nil
		}
		employees, err = l.svcCtx.Model.HrmEmployee.FindByCondition(l.ctx, nil,
			condition.NewChain().In(hrm_employee.Uuid, employeeUuids).Build()...)
		if err != nil {
			return nil, err
		}
	} else {
		employees, err = l.svcCtx.Model.HrmEmployee.FindByCondition(l.ctx, nil,
			condition.NewChain().NotEqual(hrm_employee.OrgUuid, "").Build()...)
		if err != nil {
			return nil, err
		}
	}

	// 区间内打卡行（settle_* 快照列已由 SettleDay/打卡时落库）。
	chain := condition.NewChain().
		GreaterEqualThan(hrm_punch_record.PunchDate, startDate).
		LessEqualThan(hrm_punch_record.PunchDate, endDate)
	if req.Name != "" {
		uuids := make([]string, 0, len(employees))
		for _, e := range employees {
			uuids = append(uuids, e.Uuid)
		}
		chain = chain.In(hrm_punch_record.EmployeeUuid, uuids)
	}
	rows, err := l.svcCtx.Model.HrmPunchRecord.FindByCondition(l.ctx, nil, chain.Build()...)
	if err != nil {
		return nil, err
	}

	// 按员工聚合。
	aggs := map[string]*personAgg{}
	for _, row := range rows {
		agg, ok := aggs[row.EmployeeUuid]
		if !ok {
			agg = newPersonAgg()
			aggs[row.EmployeeUuid] = agg
		}
		agg.statusCounts[row.SettleStatus]++
		if row.SettleStatus == SettleMissing && !row.ClockInTime.Valid && !row.ClockOutTime.Valid {
			agg.absentDays++
		}
		agg.totalWorkHours += row.SettleWorkHours
	}

	// 统计卡（筛选后的员工集合）。
	cards := types.HrmPunchStatsCard{}
	cards.StatEmployees = int64(len(aggs))
	abnormalEmployees := map[string]bool{}
	for _, row := range rows {
		switch row.SettleStatus {
		case SettleLate:
			cards.LateCount++
		case SettleEarly:
			cards.EarlyCount++
		case SettleMissing:
			cards.MissingCount++
		}
		if abnormalStatuses[row.SettleStatus] {
			abnormalEmployees[row.EmployeeUuid] = true
		}
	}
	cards.AbnormalEmployees = int64(len(abnormalEmployees))
	cards.RecordCount = int64(len(rows))

	// 按人明细。
	persons := make([]types.HrmPunchPersonStats, 0, len(employees))
	for _, employee := range employees {
		persons = append(persons, l.buildPersonStats(employee, startDate, endDate, aggs[employee.Uuid]))
	}

	// 仅看异常（onlyAbnormal=1）。
	if req.OnlyAbnormal == "1" || req.OnlyAbnormal == "true" {
		filtered := make([]types.HrmPunchPersonStats, 0, len(persons))
		for _, p := range persons {
			if p.AbnormalTotal > 0 {
				filtered = append(filtered, p)
			}
		}
		persons = filtered
	}

	// 排序 + 内存分页（按工号升序）。
	sort.Slice(persons, func(i, j int) bool { return persons[i].EmployeeNo < persons[j].EmployeeNo })
	total := int64(len(persons))
	start, size := pageBounds(int64(req.Current), int64(req.Size), total)
	return &types.StatisticsResponse{
		PageResponse: types.PageResponse{Current: req.Current, Size: req.Size, Total: total},
		Cards:        cards,
		Records:      persons[start:size],
	}, nil
}

// buildPersonStats 单员工按人明细行（应出勤天按生效规则 WorkDays 逐日判定）。
func (l *Statistics) buildPersonStats(employee *hrm_employee.HrmEmployee, startDate, endDate string, agg *personAgg) types.HrmPunchPersonStats {
	person := types.HrmPunchPersonStats{
		EmployeeUuid: employee.Uuid,
		EmployeeNo:   employee.EmployeeNo,
		EmployeeName: employee.Name,
		OrgName:      orgNameOf(l.ctx, l.svcCtx, employee.OrgUuid),
	}

	// 应出勤天：生效规则 WorkDays；无规则兜底按周一至周五（仓库默认口径）。
	workDays := map[int64]bool{1: true, 2: true, 3: true, 4: true, 5: true}
	rule, err := attendancegroup.ResolveEffectiveRule(l.ctx, l.svcCtx, employee)
	if err != nil {
		l.Errorf("resolve rule for %s: %v", employee.Uuid, err)
	} else if rule != nil && rule.Found {
		workDays = map[int64]bool{}
		for _, d := range rule.WorkDays {
			workDays[d] = true
		}
	}
	for date := startDate; date != ""; {
		if workDays[weekdayOf(date)] {
			person.ExpectedDays++
		}
		if date == endDate {
			break
		}
		next, nextErr := addOneDay(date)
		if nextErr != nil {
			break
		}
		date = next
	}

	if agg != nil {
		person.AttendDays = agg.statusCounts[SettleNormal] + agg.statusCounts[SettleLate] + agg.statusCounts[SettleEarly]
		person.AbsentDays = agg.absentDays
		person.LateCount = agg.statusCounts[SettleLate]
		person.EarlyCount = agg.statusCounts[SettleEarly]
		person.MissingCount = agg.statusCounts[SettleMissing]
		person.AbnormalTotal = person.LateCount + person.EarlyCount + person.MissingCount
		person.TotalWorkHours = agg.totalWorkHours
		person.AbnormalSummary = abnormalSummary(person.LateCount, person.EarlyCount, person.MissingCount)
	}
	return person
}

// resolveStatsRange 统计日期区间：两端可缺省（缺省当月 1 号至月末），非法日期报 10630。
func resolveStatsRange(startDate, endDate string) (string, string, error) {
	now := time.Now()
	if startDate == "" && endDate == "" {
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
		end := start.AddDate(0, 1, -1)
		return start.Format(DateLayout), end.Format(DateLayout), nil
	}
	if startDate == "" {
		startDate = endDate
	}
	if endDate == "" {
		endDate = startDate
	}
	start, err := time.ParseInLocation(DateLayout, startDate, time.Local)
	if err != nil {
		return "", "", status.ErrorMessage(errcodes.PunchStatisticsInvalidCode, "统计开始日期必须为 yyyy-MM-dd: "+startDate)
	}
	end, err := time.ParseInLocation(DateLayout, endDate, time.Local)
	if err != nil {
		return "", "", status.ErrorMessage(errcodes.PunchStatisticsInvalidCode, "统计结束日期必须为 yyyy-MM-dd: "+endDate)
	}
	if end.Before(start) {
		return "", "", status.ErrorMessage(errcodes.PunchStatisticsInvalidCode, "统计结束日期不能早于开始日期")
	}
	if end.Sub(start).Hours() > 24*366 {
		return "", "", status.ErrorMessage(errcodes.PunchStatisticsInvalidCode, "统计日期区间不能超过一年")
	}
	return start.Format(DateLayout), end.Format(DateLayout), nil
}

// addOneDay yyyy-MM-dd + 1 天。
func addOneDay(date string) (string, error) {
	t, err := time.ParseInLocation(DateLayout, date, time.Local)
	if err != nil {
		return "", err
	}
	return t.AddDate(0, 0, 1).Format(DateLayout), nil
}

// abnormalSummary 异常摘要文本（如「迟到3次,缺卡1次」；无异常返回空）。
func abnormalSummary(late, early, missing int64) string {
	summary := ""
	if late > 0 {
		summary += "迟到" + strconv.FormatInt(late, 10) + "次"
	}
	if early > 0 {
		if summary != "" {
			summary += ","
		}
		summary += "早退" + strconv.FormatInt(early, 10) + "次"
	}
	if missing > 0 {
		if summary != "" {
			summary += ","
		}
		summary += "缺卡" + strconv.FormatInt(missing, 10) + "次"
	}
	return summary
}

// pageBounds 内存分页边界（Current/Size 防御性钳制）。
func pageBounds(current, size, total int64) (int64, int64) {
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 10
	}
	start := (current - 1) * size
	if start > total {
		start = total
	}
	end := start + size
	if end > total {
		end = total
	}
	return start, end
}

// emptyStatistics 无员工命中的空响应（统计卡全 0）。
func emptyStatistics(req *types.StatisticsRequest) *types.StatisticsResponse {
	return &types.StatisticsResponse{
		PageResponse: types.PageResponse{Current: req.Current, Size: req.Size, Total: 0},
		Cards:        types.HrmPunchStatsCard{},
		Records:      []types.HrmPunchPersonStats{},
	}
}

// orgNameOf 员工所属组织名（空/查无返回空串）。
func orgNameOf(ctx context.Context, svcCtx *svc.ServiceContext, orgUuid string) string {
	if orgUuid == "" {
		return ""
	}
	org, err := svcCtx.Model.HrmOrg.FindOneByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_org.Uuid, orgUuid).Build()...)
	if err != nil || org == nil {
		return ""
	}
	return org.OrgName
}
