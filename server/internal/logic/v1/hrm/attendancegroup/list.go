package attendancegroup

import (
	"context"
	"net/http"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_attendance_group"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/attendancegroup"
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
		svcCtx: svcCtx, r: r,
	}
}

// List 考勤组分页列表，支持按名称/班制/状态筛选.
func (l *List) List(req *types.ListRequest) (resp *types.ListResponse, err error) {
	rows, total, err := l.svcCtx.Model.HrmAttendanceGroup.PageByCondition(l.ctx, nil,
		condition.NewChain().
			Page(req.Current, req.Size).
			Like(hrm_attendance_group.GroupName, "%"+req.GroupName+"%", condition.WithSkip(req.GroupName == "")).
			Equal(hrm_attendance_group.GroupType, req.GroupType, condition.WithSkip(req.GroupType == "")).
			Equal(hrm_attendance_group.Status, req.Status, condition.WithSkip(req.Status == "")).
			OrderByDesc(hrm_attendance_group.Id).
			Build()...)
	if err != nil {
		return nil, err
	}

	records := make([]types.HrmAttendanceGroup, 0, len(rows))
	for _, row := range rows {
		records = append(records, groupToType(row))
	}

	return &types.ListResponse{PageResponse: types.PageResponse{
		Current: req.Current,
		Size:    req.Size,
		Total:   total,
	}, Records: records}, nil
}

// groupToType row → api type（列表/编辑回显共用）.
func groupToType(row *hrm_attendance_group.HrmAttendanceGroup) types.HrmAttendanceGroup {
	return types.HrmAttendanceGroup{
		Uuid:                row.Uuid,
		GroupName:           row.GroupName,
		GroupType:           row.GroupType,
		WorkDays:            csvToInt64s(row.WorkDays),
		StartTime:           row.StartTime,
		EndTime:             row.EndTime,
		FlexibleMinutes:     row.FlexibleMinutes,
		UnassignedShiftUuid: row.UnassignedShiftUuid,
		DefaultShiftUuids:   splitCsv(row.DefaultShiftUuids),
		OrgUuids:            splitCsv(row.OrgUuids),
		MemberUuids:         splitCsv(row.MemberUuids),
		Managers:            splitCsv(row.Managers),
		Status:              row.Status,
		Remark:              row.Remark,
		CreateTime:          row.CreateTime.Format(time.DateTime),
		UpdateTime:          row.UpdateTime.Format(time.DateTime),
	}
}
