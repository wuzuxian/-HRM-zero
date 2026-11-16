package scheduleboard

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/jzero-io/jzero/core/status"
	"github.com/xuri/excelize/v2"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/scheduleboard"
)

type Import struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewImport(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Import {
	return &Import{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

func (l *Import) Import(req *types.ImportRequest) (resp *types.ImportResponse, err error) {
	// 导入排班矩阵模板（xlsx base64）：行=员工工号，列=1..31 班次编码（空/休=清空）。
	// 表头约定: 第 1 列 employee_no，其后列名 1..31（按当月天数取前 N 列）。
	if req.FileContent == "" {
		return nil, status.ErrorMessage(errcodes.ScheduleBoardImportInvalidCode, "导入内容不能为空")
	}
	first, next, err := parseMonth(req.Month)
	if err != nil {
		return nil, err
	}
	_, members, err := groupAndMembers(l.ctx, l.svcCtx, req.GroupUuid)
	if err != nil {
		return nil, err
	}
	memberByNo := make(map[string]string)
	for _, empUuid := range members {
		employee, err := l.svcCtx.Model.HrmEmployee.FindOneByUuid(l.ctx, nil, empUuid)
		if err != nil {
			continue
		}
		memberByNo[employee.EmployeeNo] = empUuid
	}

	days := monthDays(first, next)
	daysInMonth := len(days)

	raw, err := xlsxBytes(req.FileContent)
	if err != nil {
		return nil, err
	}
	f, err := excelize.OpenReader(strings.NewReader(string(raw)))
	if err != nil {
		return nil, status.ErrorMessage(errcodes.ScheduleBoardImportInvalidCode, "xlsx 解析失败: "+err.Error())
	}
	sheet := f.GetSheetName(0)
	grid, err := f.GetRows(sheet)
	if err != nil {
		return nil, status.ErrorMessage(errcodes.ScheduleBoardImportInvalidCode, "xlsx 读取失败: "+err.Error())
	}
	if len(grid) < 2 {
		return nil, status.ErrorMessage(errcodes.ScheduleBoardImportInvalidCode, "导入内容至少要有表头和一行数据")
	}

	// 表头: ["employee_no", ("name",), "1", "2", ...] —— name 列可选（与导出格式 round-trip），
	// 其后列名必须为 1..31 数字（日序号 = 列名）。
	header := grid[0]
	if len(header) < 2 || strings.TrimSpace(header[0]) != "employee_no" {
		return nil, status.ErrorMessage(errcodes.ScheduleBoardImportInvalidCode, "第 1 列必须为 employee_no")
	}
	colOffset := 1
	if strings.TrimSpace(header[1]) == "name" {
		colOffset = 2
	}
	dayColumns := make([]int, 0, daysInMonth) // 网格列索引 → 日序号（dayIdx）
	for dayIdx, col := 0, colOffset; col < len(header) && dayIdx < daysInMonth; col, dayIdx = col+1, dayIdx+1 {
		if strings.TrimSpace(header[col]) != fmt.Sprintf("%d", dayIdx+1) {
			return nil, status.ErrorMessage(errcodes.ScheduleBoardImportInvalidCode,
				fmt.Sprintf("表头第 %d 列必须为 %d", col+1, dayIdx+1))
		}
		dayColumns = append(dayColumns, col)
	}
	if len(dayColumns) == 0 {
		return nil, status.ErrorMessage(errcodes.ScheduleBoardImportInvalidCode, "没有可导入的日期列")
	}

	// 班次编码 → shift uuid（编码 = 班次名称精确匹配；空/休 = 清空）
	shifts, err := l.svcCtx.Model.HrmShift.FindByCondition(l.ctx, nil)
	if err != nil {
		return nil, err
	}
	shiftByCode := make(map[string]string, len(shifts))
	for _, s := range shifts {
		shiftByCode[s.ShiftName] = s.Uuid
	}

	type cell struct{ emp, date, shift string }
	cells := make([]cell, 0, len(grid)*len(dayColumns))
	for lineNo, row := range grid[1:] {
		if len(row) == 0 || strings.TrimSpace(row[0]) == "" {
			continue
		}
		employeeNo := strings.TrimSpace(row[0])
		empUuid, ok := memberByNo[employeeNo]
		if !ok {
			return nil, status.ErrorMessage(errcodes.ScheduleBoardEmployeeNotFoundCode,
				fmt.Sprintf("第 %d 行工号不在该考勤组适用范围内: %s", lineNo+2, employeeNo))
		}
		for dayIdx, col := range dayColumns {
			code := ""
			if col < len(row) {
				code = strings.TrimSpace(row[col])
			}
			shiftUuid := ""
			if code != "" && code != "休" {
				var ok bool
				shiftUuid, ok = shiftByCode[code]
				if !ok {
					return nil, status.ErrorMessage(errcodes.ScheduleBoardImportInvalidCode,
						fmt.Sprintf("第 %d 行第 %d 列班次编码不存在: %s", lineNo+2, col+1, code))
				}
			}
			cells = append(cells, cell{empUuid, days[dayIdx], shiftUuid})
		}
	}

	var saved int64
	if err = l.svcCtx.SqlxConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		for _, c := range cells {
			if err := upsertEntry(ctx, l.svcCtx, session, c.emp, c.date, c.shift, req.GroupUuid, false); err != nil {
				return err
			}
			saved++
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &types.ImportResponse{SuccessCount: saved}, nil
}
