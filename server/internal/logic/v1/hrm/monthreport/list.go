package monthreport

// List 考勤月报列表（demo-inventory 4.4）：按月按人聚合行。
//   - 锁定月 → 月报表快照行（changed 标记保留）；未锁定月 → 实时聚合。
//   - 筛选 月份/部门含下级/考勤组名(模糊)/姓名(模糊) + 只看异常 + 只看锁定后变更。

import (
	"context"
	"net/http"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/monthreport"
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

// abnormalRow 行是否异常（late/early/missing/absent 任一 >0）。
func abnormalRow(row MonthReportRow) bool {
	return row.LateCount > 0 || row.EarlyCount > 0 || row.MissingCount > 0 || row.AbsentDays > 0
}

func (l *List) List(req *types.ListRequest) (resp *types.ListResponse, err error) {
	yearMonth, err := validateYearMonth(req.YearMonth)
	if err != nil {
		return nil, statusError(errcodes.MonthReportInvalidCode, err)
	}

	// 部门含下级展开。
	var orgUuids []string
	if req.OrgUuid != "" {
		orgUuids, err = subOrgUuids(l.ctx, l.svcCtx, req.OrgUuid)
		if err != nil {
			return nil, err
		}
	}
	// 姓名（模糊）→ 员工 uuid 集合。
	var employeeUuids []string
	if req.Name != "" {
		employeeUuids, err = employeeUuidsByName(l.ctx, l.svcCtx, req.Name)
		if err != nil {
			return nil, err
		}
	}

	locked, err := isLocked(l.ctx, l.svcCtx, yearMonth)
	if err != nil {
		return nil, err
	}

	rows := make([]MonthReportRow, 0)
	changedFlags := map[string]bool{}
	if locked {
		snapshots, snapErr := snapshotRowsForList(l.ctx, l.svcCtx, yearMonth, orgUuids, employeeUuids)
		if snapErr != nil {
			return nil, snapErr
		}
		for _, s := range snapshots {
			rows = append(rows, s.row)
			if s.changed {
				changedFlags[s.row.EmployeeUuid] = true
			}
		}
	} else {
		employees, empErr := listMonthEmployees(l.ctx, l.svcCtx, orgUuids, employeeUuids)
		if empErr != nil {
			return nil, empErr
		}
		rows, err = aggregateRows(l.ctx, l.svcCtx, yearMonth, employees)
		if err != nil {
			return nil, err
		}
	}

	// 考勤组名（模糊）过滤（快照行有 att_group_name；实时行按生效规则解析过）。
	if req.AttGroupName != "" {
		filtered := make([]MonthReportRow, 0, len(rows))
		for _, row := range rows {
			if strings.Contains(row.AttGroupName, req.AttGroupName) {
				filtered = append(filtered, row)
			}
		}
		rows = filtered
	}
	// 只看异常。
	if req.OnlyAbnormal == "1" || req.OnlyAbnormal == "true" {
		filtered := make([]MonthReportRow, 0, len(rows))
		for _, row := range rows {
			if abnormalRow(row) {
				filtered = append(filtered, row)
			}
		}
		rows = filtered
	}
	// 只看锁定后变更。
	if req.OnlyChanged == "1" || req.OnlyChanged == "true" {
		filtered := make([]MonthReportRow, 0, len(rows))
		for _, row := range rows {
			if changedFlags[row.EmployeeUuid] {
				filtered = append(filtered, row)
			}
		}
		rows = filtered
	}

	records := make([]types.HrmMonthReportRow, 0, len(rows))
	for _, row := range rows {
		records = append(records, toTypeRow(row, changedFlags[row.EmployeeUuid]))
	}

	// 内存分页（工号升序，与 HRM-22 一致）。
	total := int64(len(records))
	start, size := pageBounds(int64(req.Current), int64(req.Size), total)
	return &types.ListResponse{
		PageResponse: types.PageResponse{Current: req.Current, Size: req.Size, Total: total},
		Locked:       locked,
		LockHint:     lockHint(yearMonth),
		Records:      records[start:size],
	}, nil
}

// toTypeRow 领域行 → API 行（附异常合计）。
func toTypeRow(row MonthReportRow, changed bool) types.HrmMonthReportRow {
	return types.HrmMonthReportRow{
		Uuid:          "",
		YearMonth:     row.YearMonth,
		EmployeeUuid:  row.EmployeeUuid,
		EmployeeNo:    row.EmployeeNo,
		EmployeeName:  row.EmployeeName,
		OrgUuid:       row.OrgUuid,
		OrgName:       row.OrgName,
		AttGroupName:  row.AttGroupName,
		ExpectedDays:  row.ExpectedDays,
		AttendDays:    row.AttendDays,
		LateCount:     row.LateCount,
		EarlyCount:    row.EarlyCount,
		MissingCount:  row.MissingCount,
		AbsentDays:    row.AbsentDays,
		WorkdayOtH:    row.WorkdayOtH,
		RestdayOtH:    row.RestdayOtH,
		HolidayOtH:    row.HolidayOtH,
		CompAccrueH:   row.CompAccrueH,
		AbnormalTotal: row.LateCount + row.EarlyCount + row.MissingCount + row.AbsentDays,
		Changed:       changed,
	}
}
