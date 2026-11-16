package monthreport

// Export 导出考勤月报 xlsx（demo-inventory 4.4 按钮）：同列表筛选条件，base64（excelize 惯例）。

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/xuri/excelize/v2"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/monthreport"
)

type Export struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewExport(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Export {
	return &Export{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

var monthReportHeader = []string{
	"部门", "考勤组", "姓名", "工号", "应出勤", "实出勤", "迟到", "早退", "缺卡", "旷工",
	"工作日加班(h)", "休息日加班(h)", "节假日加班(h)", "调休入账(h)", "锁定后变更",
}

func (l *Export) Export(req *types.ExportRequest) (resp *types.ExportResponse, err error) {
	yearMonth, err := validateYearMonth(req.YearMonth)
	if err != nil {
		return nil, statusError(errcodes.MonthReportInvalidCode, err)
	}

	var orgUuids []string
	if req.OrgUuid != "" {
		orgUuids, err = subOrgUuids(l.ctx, l.svcCtx, req.OrgUuid)
		if err != nil {
			return nil, err
		}
	}
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

	if req.AttGroupName != "" {
		filtered := make([]MonthReportRow, 0, len(rows))
		for _, row := range rows {
			if containsStr(row.AttGroupName, req.AttGroupName) {
				filtered = append(filtered, row)
			}
		}
		rows = filtered
	}
	if req.OnlyAbnormal == "1" || req.OnlyAbnormal == "true" {
		filtered := make([]MonthReportRow, 0, len(rows))
		for _, row := range rows {
			if abnormalRow(row) {
				filtered = append(filtered, row)
			}
		}
		rows = filtered
	}
	if req.OnlyChanged == "1" || req.OnlyChanged == "true" {
		filtered := make([]MonthReportRow, 0, len(rows))
		for _, row := range rows {
			if changedFlags[row.EmployeeUuid] {
				filtered = append(filtered, row)
			}
		}
		rows = filtered
	}

	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	if err := f.SetSheetRow(sheet, "A1", &monthReportHeader); err != nil {
		return nil, err
	}
	for i, row := range rows {
		line := []string{
			row.OrgName, row.AttGroupName, row.EmployeeName, row.EmployeeNo,
			fmt.Sprintf("%d", row.ExpectedDays), fmt.Sprintf("%d", row.AttendDays),
			fmt.Sprintf("%d", row.LateCount), fmt.Sprintf("%d", row.EarlyCount),
			fmt.Sprintf("%d", row.MissingCount), fmt.Sprintf("%d", row.AbsentDays),
			fmt.Sprintf("%.2f", row.WorkdayOtH), fmt.Sprintf("%.2f", row.RestdayOtH),
			fmt.Sprintf("%.2f", row.HolidayOtH), fmt.Sprintf("%.2f", row.CompAccrueH),
			changedName(changedFlags[row.EmployeeUuid]),
		}
		cell, _ := excelize.CoordinatesToCellName(1, i+2)
		if err := f.SetSheetRow(sheet, cell, &line); err != nil {
			return nil, err
		}
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return &types.ExportResponse{
		FileName: fmt.Sprintf("month-report-%s.xlsx", time.Now().Format("20060102150405")),
		Content:  base64.StdEncoding.EncodeToString(buf.Bytes()),
	}, nil
}

func changedName(changed bool) string {
	if changed {
		return "是"
	}
	return ""
}

// containsStr strings.Contains（保持与列表同款过滤语义）。
func containsStr(s, sub string) bool {
	if sub == "" {
		return true
	}
	return len(s) >= len(sub) && stringsContains(s, sub)
}

func stringsContains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
