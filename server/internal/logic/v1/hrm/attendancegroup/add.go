package attendancegroup

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_attendance_group"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/attendancegroup"
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
		svcCtx: svcCtx, r: r,
	}
}

// Add 新增考勤组（GLOSSARY「考勤组」）: 班制/考勤时间/适用范围/负责人/启停.
func (l *Add) Add(req *types.AddRequest) (resp *types.AddResponse, err error) {
	if req.Status == "" {
		req.Status = statusEnabled
	}
	if err = validateGroupBasic(req.GroupName, req.GroupType, req.WorkDays, req.StartTime, req.EndTime, req.Status); err != nil {
		return nil, err
	}

	_, err = l.svcCtx.Model.HrmAttendanceGroup.Insert(l.ctx, nil, &hrm_attendance_group.HrmAttendanceGroup{
		Uuid:                uuid.New().String(),
		GroupName:           req.GroupName,
		GroupType:           req.GroupType,
		WorkDays:            joinInt64s(req.WorkDays),
		StartTime:           req.StartTime,
		EndTime:             req.EndTime,
		FlexibleMinutes:     req.FlexibleMinutes,
		UnassignedShiftUuid: req.UnassignedShiftUuid,
		DefaultShiftUuids:   joinCsv(req.DefaultShiftUuids),
		OrgUuids:            joinCsv(req.OrgUuids),
		MemberUuids:         joinCsv(req.MemberUuids),
		Managers:            joinCsv(req.Managers),
		Status:              req.Status,
		Remark:              req.Remark,
	})
	if err != nil {
		return nil, err
	}

	return &types.AddResponse{}, nil
}

// validateGroupBasic 考勤组字段校验（Add/Edit 共用）.
func validateGroupBasic(groupName, groupType string, workDays []int64, startTime, endTime, groupStatus string) error {
	if groupName == "" {
		return status.ErrorMessage(errcodes.AttendanceGroupInvalidCode, "考勤组名称不能为空")
	}
	switch groupType {
	case groupTypeFixed:
		// 固定班制必须携带考勤时间
		if startTime == "" || endTime == "" {
			return status.ErrorMessage(errcodes.AttendanceGroupInvalidCode, "固定班制必须填写上下班时间")
		}
		if err := validateGroupTimes(startTime, endTime); err != nil {
			return err
		}
		if len(workDays) == 0 {
			return status.ErrorMessage(errcodes.AttendanceGroupInvalidCode, "固定班制必须选择考勤日")
		}
	case groupTypeScheduling, groupTypeFlexible:
		// 排班制/自由工时: 考勤时间可选（排班制由排班兜底, 自由工时无固定时间）
	default:
		return status.ErrorMessage(errcodes.AttendanceGroupInvalidCode, "班制必须是 固定班制/排班制/自由工时")
	}
	if len(workDays) > 0 {
		for _, d := range workDays {
			if d < 1 || d > 7 {
				return status.ErrorMessage(errcodes.AttendanceGroupInvalidCode, "考勤日必须是 1-7（周一至周日）")
			}
		}
	}
	if groupStatus != statusEnabled && groupStatus != statusDisabled {
		return status.ErrorMessage(errcodes.AttendanceGroupInvalidCode, "考勤组状态必须是 启用 或 停用")
	}
	return nil
}

// validateGroupTimes 上下班时间格式校验（HH:mm, 早于上班时间=跨天, 相同非法 — 复用班次口径）.
func validateGroupTimes(startTime, endTime string) error {
	if !validateHHmm(startTime) || !validateHHmm(endTime) {
		return status.ErrorMessage(errcodes.AttendanceGroupInvalidCode, "上下班时间格式必须为 HH:mm")
	}
	if startTime == endTime {
		return status.ErrorMessage(errcodes.AttendanceGroupInvalidCode, "上下班时间不能相同")
	}
	return nil
}

func validateHHmm(v string) bool {
	if len(v) != 5 || v[2] != ':' {
		return false
	}
	h, okH := parseInt64(v[:2])
	m, okM := parseInt64(v[3:])
	return okH && okM && h >= 0 && h <= 23 && m >= 0 && m <= 59
}

func joinInt64s(values []int64) string {
	parts := make([]string, 0, len(values))
	for _, v := range values {
		parts = append(parts, formatInt64(v))
	}
	return joinCsv(parts)
}

func formatInt64(v int64) string {
	if v == 0 {
		return "0"
	}
	digits := ""
	for v > 0 {
		digits = string(rune('0'+v%10)) + digits
		v /= 10
	}
	return digits
}
