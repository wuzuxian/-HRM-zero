package schedulecycle

import (
	"context"
	"net/http"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_schedule_cycle"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/schedulecycle"
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

// List 排班周期分页列表，支持按周期名称/状态筛选.
func (l *List) List(req *types.ListRequest) (resp *types.ListResponse, err error) {
	rows, total, err := l.svcCtx.Model.HrmScheduleCycle.PageByCondition(l.ctx, nil,
		condition.NewChain().
			Page(req.Current, req.Size).
			Like(hrm_schedule_cycle.CycleName, "%"+req.CycleName+"%", condition.WithSkip(req.CycleName == "")).
			Equal(hrm_schedule_cycle.Status, req.Status, condition.WithSkip(req.Status == "")).
			OrderByDesc(hrm_schedule_cycle.Id).
			Build()...)
	if err != nil {
		return nil, err
	}

	records := make([]types.HrmScheduleCycle, 0, len(rows))
	for _, row := range rows {
		records = append(records, cycleToType(row))
	}

	return &types.ListResponse{PageResponse: types.PageResponse{
		Current: req.Current,
		Size:    req.Size,
		Total:   total,
	}, Records: records}, nil
}

// cycleToType row → api type（列表/编辑回显共用）.
func cycleToType(row *hrm_schedule_cycle.HrmScheduleCycle) types.HrmScheduleCycle {
	return types.HrmScheduleCycle{
		Uuid:       row.Uuid,
		CycleName:  row.CycleName,
		ShiftUuids: splitCsv(row.ShiftUuids),
		GroupUuids: splitCsv(row.GroupUuids),
		Status:     row.Status,
		Remark:     row.Remark,
		CreateTime: row.CreateTime.Format(time.DateTime),
		UpdateTime: row.UpdateTime.Format(time.DateTime),
	}
}
