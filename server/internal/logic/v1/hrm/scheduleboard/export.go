package scheduleboard

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/xuri/excelize/v2"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero/core/stores/condition"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_schedule_entry"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/scheduleboard"
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

func (l *Export) Export(req *types.ExportRequest) (resp *types.ExportResponse, err error) {
	// 导出排班矩阵（xlsx：行=工号/姓名，列=1..当月天数，值=班次名称，休=空）。
	first, next, err := parseMonth(req.Month)
	if err != nil {
		return nil, err
	}
	_, members, err := groupAndMembers(l.ctx, l.svcCtx, req.GroupUuid)
	if err != nil {
		return nil, err
	}
	days := monthDays(first, next)

	entries, err := l.svcCtx.Model.HrmScheduleEntry.FindByCondition(l.ctx, nil,
		condition.NewChain().
			GreaterEqualThan(hrm_schedule_entry.ScheduleDate, days[0]).
			LessEqualThan(hrm_schedule_entry.ScheduleDate, days[len(days)-1]).
			In(hrm_schedule_entry.EmployeeUuid, members).
			Build()...)
	if err != nil {
		return nil, err
	}
	type cellKey struct{ emp, date string }
	entryMap := make(map[cellKey]*hrm_schedule_entry.HrmScheduleEntry, len(entries))
	for _, e := range entries {
		entryMap[cellKey{e.EmployeeUuid, e.ScheduleDate}] = e
	}
	shiftNameByUuid := make(map[string]string)
	shifts, err := l.svcCtx.Model.HrmShift.FindByCondition(l.ctx, nil)
	if err != nil {
		return nil, err
	}
	for _, s := range shifts {
		shiftNameByUuid[s.Uuid] = s.ShiftName
	}

	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	header := []string{"employee_no", "name"}
	for i := range days {
		header = append(header, fmt.Sprintf("%d", i+1))
	}
	if err := f.SetSheetRow(sheet, "A1", &header); err != nil {
		return nil, err
	}
	lineNo := 2
	for _, empUuid := range members {
		employee, err := l.svcCtx.Model.HrmEmployee.FindOneByUuid(l.ctx, nil, empUuid)
		if err != nil {
			if err == hrm_employee.ErrNotFound {
				continue
			}
			return nil, err
		}
		line := []string{employee.EmployeeNo, employee.Name}
		for _, day := range days {
			code := ""
			if e := entryMap[cellKey{empUuid, day}]; e != nil && e.ShiftUuid != "" {
				code = shiftNameByUuid[e.ShiftUuid]
			}
			line = append(line, code)
		}
		cell0, _ := excelize.CoordinatesToCellName(1, lineNo)
		if err := f.SetSheetRow(sheet, cell0, &line); err != nil {
			return nil, err
		}
		lineNo++
	}
	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	fileName := fmt.Sprintf("schedule-board-%s-%s.xlsx", req.Month, req.GroupUuid[:8])
	return &types.ExportResponse{
		FileName: fileName,
		Content:  base64.StdEncoding.EncodeToString(buf.Bytes()),
	}, nil
}
