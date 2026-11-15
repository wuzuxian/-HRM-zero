package attendanceconfig

import (
	"strings"
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_attendance_config"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_org"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/attendanceconfig"
)

type Save struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewSave(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Save {
	return &Save{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx, r: r,
	}
}

// Save 保存组织默认规则（每组织一条; upsert）. 快速开始/新增/编辑共用此端点.
func (l *Save) Save(req *types.SaveRequest) (resp *types.SaveResponse, err error) {
	if req.OrgUuid == "" {
		return nil, status.ErrorMessage(errcodes.AttendanceConfigOrgNotFoundCode, "必须指定组织")
	}
	if _, err = l.svcCtx.Model.HrmOrg.FindOneByUuid(l.ctx, nil, req.OrgUuid); err != nil {
		if err == hrm_org.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.AttendanceConfigOrgNotFoundCode, "组织不存在")
		}
		return nil, err
	}
	if err = validateConfigBasic(req.StartTime, req.EndTime, req.WorkDays); err != nil {
		return nil, err
	}

	row, err := l.svcCtx.Model.HrmAttendanceConfig.FindOneByOrgUuid(l.ctx, nil, req.OrgUuid)
	if err != nil {
		if err == hrm_attendance_config.ErrNotFound {
			// 新增
			if _, err = l.svcCtx.Model.HrmAttendanceConfig.Insert(l.ctx, nil, &hrm_attendance_config.HrmAttendanceConfig{
				Uuid:      uuid.New().String(),
				OrgUuid:   req.OrgUuid,
				StartTime: req.StartTime,
				EndTime:   req.EndTime,
				WorkDays:  joinInt64s(req.WorkDays),
				Remark:    req.Remark,
			}); err != nil {
				return nil, err
			}
			return &types.SaveResponse{}, nil
		}
		return nil, err
	}

	// 编辑（upsert）
	row.StartTime = req.StartTime
	row.EndTime = req.EndTime
	row.WorkDays = joinInt64s(req.WorkDays)
	row.Remark = req.Remark
	if err = l.svcCtx.Model.HrmAttendanceConfig.Update(l.ctx, nil, row); err != nil {
		return nil, err
	}
	return &types.SaveResponse{}, nil
}

// validateConfigBasic 组织规则字段校验（时间可选; 填了必须成对合法）.
func validateConfigBasic(startTime, endTime string, workDays []int64) error {
	if startTime == "" && endTime == "" {
		if len(workDays) == 0 {
			return status.ErrorMessage(errcodes.AttendanceConfigInvalidCode, "考勤时间与考勤日至少填写一项")
		}
	} else if startTime == "" || endTime == "" {
		return status.ErrorMessage(errcodes.AttendanceConfigInvalidCode, "上下班时间必须成对填写")
	} else {
		if !validateHHmm(startTime) || !validateHHmm(endTime) {
			return status.ErrorMessage(errcodes.AttendanceConfigInvalidCode, "上下班时间格式必须为 HH:mm")
		}
		if startTime == endTime {
			return status.ErrorMessage(errcodes.AttendanceConfigInvalidCode, "上下班时间不能相同")
		}
	}
	for _, d := range workDays {
		if d < 1 || d > 7 {
			return status.ErrorMessage(errcodes.AttendanceConfigInvalidCode, "考勤日必须是 1-7（周一至周日）")
		}
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
	return strings.Join(parts, ",")
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
