package attendancegroup

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_attendance_group"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/attendancegroup"
)

type ToggleStatus struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewToggleStatus(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *ToggleStatus {
	return &ToggleStatus{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx, r: r,
	}
}

// ToggleStatus 启用/停用考勤组. 停用后三级回退跳过该考勤组（员工落到组织默认规则）.
func (l *ToggleStatus) ToggleStatus(req *types.ToggleStatusRequest) (resp *types.ToggleStatusResponse, err error) {
	if req.Status != statusEnabled && req.Status != statusDisabled {
		return nil, status.ErrorMessage(errcodes.AttendanceGroupInvalidCode, "考勤组状态必须是 启用 或 停用")
	}
	row, err := l.svcCtx.Model.HrmAttendanceGroup.FindOneByUuid(l.ctx, nil, req.Uuid)
	if err != nil {
		if err == hrm_attendance_group.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.AttendanceGroupNotFoundCode, "考勤组不存在")
		}
		return nil, err
	}
	row.Status = req.Status
	if err = l.svcCtx.Model.HrmAttendanceGroup.Update(l.ctx, nil, row); err != nil {
		return nil, err
	}
	return &types.ToggleStatusResponse{}, nil
}
