package attendancegroup

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_attendance_group"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/attendancegroup"
)

type ShiftUsage struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewShiftUsage(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *ShiftUsage {
	return &ShiftUsage{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx, r: r,
	}
}

// ShiftUsage 班次引用查询: 指定 shiftUuid 时返回引用该班次的考勤组
// （default_shift_uuids 或 unassigned_shift_uuid 命中）; 缺省时返回全部考勤组概要.
// 班次管理页的「引用考勤组」列由此端点支撑（HRM-11 预留列）.
func (l *ShiftUsage) ShiftUsage(req *types.ShiftUsageRequest) (resp *types.ShiftUsageResponse, err error) {
	chain := condition.NewChain().OrderByDesc(hrm_attendance_group.Id)

	rows, err := l.svcCtx.Model.HrmAttendanceGroup.FindByCondition(l.ctx, nil, chain.Build()...)
	if err != nil {
		return nil, err
	}

	groups := make([]types.GroupUsageBrief, 0, len(rows))
	for _, row := range rows {
		if req.ShiftUuid != "" &&
			!containsStr(splitCsv(row.DefaultShiftUuids), req.ShiftUuid) &&
			row.UnassignedShiftUuid != req.ShiftUuid {
			continue
		}
		groups = append(groups, types.GroupUsageBrief{
			Uuid:      row.Uuid,
			GroupName: row.GroupName,
			Status:    row.Status,
		})
	}

	return &types.ShiftUsageResponse{Count: int64(len(groups)), Groups: groups}, nil
}
