package punch

// Export 管理端导出打卡记录（demo-inventory 4.4 工具栏「导出」）：同列表筛选条件，
// 输出 xlsx（base64），复用 employee 导出的 excelize 惯例。

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/xuri/excelize/v2"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_punch_record"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/punch"
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

var exportHeader = []string{"打卡日期", "星期", "姓名", "登录账号", "签到时间", "签到状态", "签到地点", "签退时间", "签退状态", "签退地点", "加班", "日结状态", "迟到(分钟)", "早退(分钟)", "工时(小时)"}

func (l *Export) Export(req *types.ExportRequest) (resp *types.ExportResponse, err error) {
	chain := condition.NewChain().
		Equal(hrm_punch_record.ClockInStatus, req.ClockInStatus,
			condition.WithSkip(req.ClockInStatus != SettleMissing && !isPunchStatus(req.ClockInStatus)))
	if req.ClockInStatus == SettleMissing {
		chain = condition.NewChain().Equal(hrm_punch_record.ClockInStatus, "")
	}
	if req.StartDate != "" {
		chain = chain.GreaterEqualThan(hrm_punch_record.PunchDate, req.StartDate)
	}
	if req.EndDate != "" {
		chain = chain.LessEqualThan(hrm_punch_record.PunchDate, req.EndDate)
	}
	if req.Name != "" {
		employeeUuids, err := matchEmployeeUuids(l.ctx, l.svcCtx, req.Name)
		if err != nil {
			return nil, err
		}
		if len(employeeUuids) == 0 {
			return &types.ExportResponse{FileName: "punch-record.xlsx", Content: base64.StdEncoding.EncodeToString(emptyXlsx())}, nil
		}
		chain = chain.In(hrm_punch_record.EmployeeUuid, employeeUuids)
	}

	rows, err := l.svcCtx.Model.HrmPunchRecord.FindByCondition(l.ctx, nil,
		chain.OrderByDesc(hrm_punch_record.PunchDate).
			OrderByAsc(hrm_punch_record.EmployeeNo).
			Build()...)
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	if err := f.SetSheetRow(sheet, "A1", &exportHeader); err != nil {
		return nil, err
	}
	for i, row := range rows {
		record := toRecord(l.ctx, l.svcCtx, row)
		line := []string{
			record.PunchDate, record.Weekday, record.EmployeeName, record.Username,
			record.ClockInTime, clockStatusName(record.ClockInStatus), record.ClockInLocation,
			record.ClockOutTime, clockStatusName(record.ClockOutStatus), record.ClockOutLocation,
			overtimeName(record.Overtime), settleStatusName(record.SettleStatus),
			fmt.Sprintf("%d", record.SettleLateMinutes), fmt.Sprintf("%d", record.SettleEarlyMinutes),
			fmt.Sprintf("%.2f", record.SettleWorkHours),
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
		FileName: fmt.Sprintf("punch-record-%s.xlsx", time.Now().Format("20060102150405")),
		Content:  base64.StdEncoding.EncodeToString(buf.Bytes()),
	}, nil
}

func emptyXlsx() []byte {
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	_ = f.SetSheetRow(sheet, "A1", &exportHeader)
	var buf bytes.Buffer
	_ = f.Write(&buf)
	return buf.Bytes()
}

func clockStatusName(v string) string {
	switch v {
	case StatusNormal:
		return "正常"
	case StatusField:
		return "外勤"
	case SettleMissing:
		return "缺卡"
	}
	return ""
}

func overtimeName(v string) string {
	if v == OvertimeYes {
		return "是"
	}
	return "否"
}

func settleStatusName(v string) string {
	switch v {
	case SettleNormal:
		return "正常"
	case SettleLate:
		return "迟到"
	case SettleEarly:
		return "早退"
	case SettleMissing:
		return "缺卡"
	case SettleRest:
		return "休息日"
	case SettleNoRule:
		return "未判定"
	}
	return "未日结"
}
