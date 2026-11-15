package perf_indicator

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_indicator"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/perf_indicator"
)

type GetAll struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewGetAll(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *GetAll {
	return &GetAll{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// GetAll returns every enabled 指标库 indicator, for the 考核模板 indicator
// picker (模板内配置指标与权重).
func (l *GetAll) GetAll(req *types.GetAllResponse) (resp *types.GetAllResponse, err error) {
	rows, err := l.svcCtx.Model.HrmPerfIndicator.FindByCondition(l.ctx, nil, condition.NewChain().
		Equal(hrm_perf_indicator.Status, StatusEnabled).
		OrderByAsc(hrm_perf_indicator.Code).
		Build()...)
	if err != nil {
		return nil, err
	}

	records := make([]types.HrmPerfIndicator, 0, len(rows))
	for _, row := range rows {
		records = append(records, indicatorRecord(row))
	}
	return &types.GetAllResponse{Records: records}, nil
}
