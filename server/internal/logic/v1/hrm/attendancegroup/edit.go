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

type Edit struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewEdit(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Edit {
	return &Edit{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx, r: r,
	}
}

// Edit 编辑考勤组（覆盖全部可编辑字段）.
func (l *Edit) Edit(req *types.EditRequest) (resp *types.EditResponse, err error) {
	if req.Uuid == "" {
		return nil, status.ErrorMessage(errcodes.AttendanceGroupNotFoundCode, "考勤组不存在")
	}
	if err = validateGroupBasic(req.GroupName, req.GroupType, req.WorkDays, req.StartTime, req.EndTime, req.Status); err != nil {
		return nil, err
	}

	row, err := l.svcCtx.Model.HrmAttendanceGroup.FindOneByUuid(l.ctx, nil, req.Uuid)
	if err != nil {
		if err == hrm_attendance_group.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.AttendanceGroupNotFoundCode, "考勤组不存在")
		}
		return nil, err
	}
	row.GroupName = req.GroupName
	row.GroupType = req.GroupType
	row.WorkDays = joinInt64s(req.WorkDays)
	row.StartTime = req.StartTime
	row.EndTime = req.EndTime
	row.FlexibleMinutes = req.FlexibleMinutes
	row.UnassignedShiftUuid = req.UnassignedShiftUuid
	row.DefaultShiftUuids = joinCsv(req.DefaultShiftUuids)
	row.OrgUuids = joinCsv(req.OrgUuids)
	row.MemberUuids = joinCsv(req.MemberUuids)
	row.Managers = joinCsv(req.Managers)
	row.Status = req.Status
	row.Remark = req.Remark

	if err = l.svcCtx.Model.HrmAttendanceGroup.Update(l.ctx, nil, row); err != nil {
		return nil, err
	}
	return &types.EditResponse{}, nil
}
