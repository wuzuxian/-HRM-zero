package punch

// Import 管理端导入外部考勤数据（demo-inventory 4.4 工具栏；HRM-22 考勤统计复用同一 seam）。
// CSV 列约定（首行为表头）:
//   date,employee_no,clock_in_time,clock_in_status,clock_in_location,
//   clock_out_time,clock_out_status,clock_out_location,overtime
// 单行非法 → 整批拒绝 10524（事务回滚，不落任何行）。

import (
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_punch_record"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/punch"
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

var importHeader = []string{"date", "employee_no", "clock_in_time", "clock_in_status", "clock_in_location", "clock_out_time", "clock_out_status", "clock_out_location", "overtime"}

func (l *Import) Import(req *types.ImportRequest) (resp *types.ImportResponse, err error) {
	if req.FileContent == "" {
		return nil, status.ErrorMessage(errcodes.PunchImportInvalidCode, "导入内容不能为空")
	}

	reader := csv.NewReader(strings.NewReader(req.FileContent))
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		return nil, status.ErrorMessage(errcodes.PunchImportInvalidCode, "CSV 解析失败: "+err.Error())
	}
	if len(records) < 2 {
		return nil, status.ErrorMessage(errcodes.PunchImportInvalidCode, "导入内容至少要有表头和一行数据")
	}
	for i, col := range importHeader {
		if i >= len(records[0]) || strings.TrimSpace(records[0][i]) != col {
			return nil, status.ErrorMessage(errcodes.PunchImportInvalidCode, fmt.Sprintf("表头第 %d 列必须为 %s", i+1, col))
		}
	}

	// 预校验全部行（员工/日期/时间），任一非法即整批拒绝。
	type pendingRow struct {
		employee  *hrm_employee.HrmEmployee
		date      string
		in        sql.NullTime
		out       sql.NullTime
		inStatus  string
		inLoc     string
		outStatus string
		outLoc    string
		overtime  string
	}
	var pending []pendingRow
	seen := map[string]bool{}
	employeeCache := map[string]*hrm_employee.HrmEmployee{}
	for lineNo, cols := range records[1:] {
		lineNo = lineNo + 2
		if len(cols) < len(importHeader) {
			return nil, status.ErrorMessage(errcodes.PunchImportInvalidCode, fmt.Sprintf("第 %d 行列数不足", lineNo))
		}
		date := strings.TrimSpace(cols[0])
		if _, parseErr := time.ParseInLocation(DateLayout, date, time.Local); parseErr != nil {
			return nil, status.ErrorMessage(errcodes.PunchImportInvalidCode, fmt.Sprintf("第 %d 行日期非法: %s", lineNo, date))
		}
		employeeNo := strings.TrimSpace(cols[1])
		employee, cached := employeeCache[employeeNo]
		if !cached {
			row, findErr := l.svcCtx.Model.HrmEmployee.FindOneByCondition(l.ctx, nil,
				condition.NewChain().Equal(hrm_employee.EmployeeNo, employeeNo).Build()...)
			if findErr != nil {
				if findErr == hrm_employee.ErrNotFound {
					return nil, status.ErrorMessage(errcodes.PunchImportInvalidCode, fmt.Sprintf("第 %d 行工号不存在: %s", lineNo, employeeNo))
				}
				return nil, findErr
			}
			employeeCache[employeeNo] = row
			employee = row
		}
		key := employee.Uuid + "|" + date
		if seen[key] {
			return nil, status.ErrorMessage(errcodes.PunchImportInvalidCode, fmt.Sprintf("第 %d 行该员工当日重复: %s %s", lineNo, employeeNo, date))
		}
		seen[key] = true

		in, inStatus, inLoc, err := parseImportClock(lineNo, cols[2], cols[3], cols[4])
		if err != nil {
			return nil, err
		}
		out, outStatus, outLoc, err := parseImportClock(lineNo, cols[5], cols[6], cols[7])
		if err != nil {
			return nil, err
		}
		overtime := strings.TrimSpace(cols[8])
		if overtime != "" && overtime != OvertimeYes && overtime != OvertimeNo {
			return nil, status.ErrorMessage(errcodes.PunchImportInvalidCode, fmt.Sprintf("第 %d 行加班标记必须为 0/1", lineNo))
		}
		if overtime == "" {
			overtime = OvertimeNo
		}
		pending = append(pending, pendingRow{
			employee: employee, date: date, in: in, out: out,
			inStatus: inStatus, inLoc: inLoc, outStatus: outStatus, outLoc: outLoc, overtime: overtime,
		})
	}

	// 事务内落库（整批原子）
	err = l.svcCtx.SqlxConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 事务内构造 model（leavebalance.modelOf 同款）
		m := hrm_punch_record.NewHrmPunchRecordModel(sqlx.NewSqlConnFromSession(session))
		for _, p := range pending {
			existing, findErr := m.FindOneByCondition(l.ctx, session,
				condition.NewChain().Equal(hrm_punch_record.PunchDate, p.date).
					Equal(hrm_punch_record.EmployeeUuid, p.employee.Uuid).Build()...)
			if findErr != nil && findErr != hrm_punch_record.ErrNotFound {
				return findErr
			}
			if existing != nil {
				// 已有行: 覆盖导入列（外部数据以导入为准）
				if updErr := m.UpdateFieldsByCondition(l.ctx, session, map[string]any{
					"clock_in_time": p.in, "clock_in_status": p.inStatus, "clock_in_location": p.inLoc,
					"clock_out_time": p.out, "clock_out_status": p.outStatus, "clock_out_location": p.outLoc,
					"overtime": p.overtime, "source": SourceImport,
				}, condition.NewChain().Equal(hrm_punch_record.Uuid, existing.Uuid).Build()...); updErr != nil {
					return updErr
				}
				continue
			}
			row := &hrm_punch_record.HrmPunchRecord{
				Uuid:         newUuid(),
				PunchDate:    p.date,
				EmployeeUuid: p.employee.Uuid,
				UserUuid:     p.employee.UserUuid,
				Username:     usernameOf(ctx, l.svcCtx, p.employee.UserUuid),
				EmployeeNo:   p.employee.EmployeeNo,
				EmployeeName: p.employee.Name,
				ClockInTime:  p.in, ClockInStatus: p.inStatus, ClockInLocation: p.inLoc,
				ClockOutTime: p.out, ClockOutStatus: p.outStatus, ClockOutLocation: p.outLoc,
				Overtime: p.overtime,
				Source:   SourceImport,
			}
			if _, insErr := m.Insert(l.ctx, session, row); insErr != nil {
				return insErr
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 导入后逐行日结（快照固化）
	for _, p := range pending {
		if _, err := SettleDay(l.ctx, l.svcCtx, p.employee.Uuid, p.date); err != nil {
			return nil, err
		}
	}

	return &types.ImportResponse{SuccessCount: int64(len(pending))}, nil
}

// parseImportClock 单列打卡三元组（时间/状态/地点），空时间 = 缺卡。
func parseImportClock(lineNo int, clockTime, clockStatus, location string) (sql.NullTime, string, string, error) {
	clockTime = strings.TrimSpace(clockTime)
	clockStatus = strings.TrimSpace(clockStatus)
	location = strings.TrimSpace(location)
	if clockTime == "" {
		return sql.NullTime{}, "", location, nil
	}
	tm, err := time.ParseInLocation(time.DateTime, clockTime, time.Local)
	if err != nil {
		return sql.NullTime{}, "", "", status.ErrorMessage(errcodes.PunchImportInvalidCode,
			fmt.Sprintf("第 %d 行打卡时间非法（须为 yyyy-MM-dd HH:mm:ss）: %s", lineNo, clockTime))
	}
	if clockStatus == "" {
		clockStatus = StatusNormal
	}
	if clockStatus != StatusNormal && clockStatus != StatusField {
		return sql.NullTime{}, "", "", status.ErrorMessage(errcodes.PunchImportInvalidCode,
			fmt.Sprintf("第 %d 行打卡状态必须为 normal/field", lineNo))
	}
	return sql.NullTime{Time: tm, Valid: true}, clockStatus, location, nil
}
