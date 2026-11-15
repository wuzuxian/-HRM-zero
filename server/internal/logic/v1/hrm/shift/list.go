package shift

import (
	"context"
	"net/http"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_shift"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/shift"
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

// List 班次列表，支持按班次名称/状态筛选.
func (l *List) List(req *types.ListRequest) (resp *types.ListResponse, err error) {
	rows, err := l.svcCtx.Model.HrmShift.FindByCondition(l.ctx, nil,
		condition.NewChain().
			Like(hrm_shift.ShiftName, "%"+req.ShiftName+"%", condition.WithSkip(req.ShiftName == "")).
			Equal(hrm_shift.Status, req.Status, condition.WithSkip(req.Status == "")).
			OrderByDesc(hrm_shift.Id).
			Build()...)
	if err != nil {
		return nil, err
	}

	records := make([]types.HrmShift, 0, len(rows))
	for _, row := range rows {
		records = append(records, types.HrmShift{
			Uuid:            row.Uuid,
			ShiftName:       row.ShiftName,
			StartTime:       row.StartTime,
			EndTime:         row.EndTime,
			CrossDay:        row.CrossDay,
			RestStart:       row.RestStart,
			RestEnd:         row.RestEnd,
			StandardHours:   row.StandardHours,
			FlexibleMinutes: row.FlexibleMinutes,
			Status:          row.Status,
			Remark:          row.Remark,
			CreateTime:      row.CreateTime.Format(time.DateTime),
			UpdateTime:      row.UpdateTime.Format(time.DateTime),
		})
	}

	return &types.ListResponse{Records: records}, nil
}
