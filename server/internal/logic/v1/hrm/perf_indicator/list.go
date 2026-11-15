package perf_indicator

import (
	"context"
	"net/http"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_indicator"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/perf_indicator"
)

// indicatorRecord renders one 指标库 row for the API.
func indicatorRecord(row *hrm_perf_indicator.HrmPerfIndicator) types.HrmPerfIndicator {
	return types.HrmPerfIndicator{
		Uuid:          row.Uuid,
		Code:          row.Code,
		Name:          row.Name,
		IndicatorType: row.IndicatorType,
		Unit:          row.Unit,
		ScoreRule:     row.ScoreRule,
		DataSource:    row.DataSource,
		DefaultWeight: row.DefaultWeight,
		Status:        row.Status,
		Remark:        row.Remark,
		CreateTime:    row.CreateTime.Format(time.DateTime),
		UpdateTime:    row.UpdateTime.Format(time.DateTime),
	}
}

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
		svcCtx: svcCtx,
		r:      r,
	}
}

func (l *List) List(req *types.ListRequest) (resp *types.ListResponse, err error) {
	rows, total, err := l.svcCtx.Model.HrmPerfIndicator.PageByCondition(l.ctx, nil, condition.NewChain().
		Page(req.Current, req.Size).
		OrderByDesc(hrm_perf_indicator.CreateTime).
		Like(hrm_perf_indicator.Code, "%"+req.Code+"%", condition.WithSkip(req.Code == "")).
		Like(hrm_perf_indicator.Name, "%"+req.Name+"%", condition.WithSkip(req.Name == "")).
		Equal(hrm_perf_indicator.IndicatorType, req.IndicatorType, condition.WithSkip(req.IndicatorType == "")).
		Equal(hrm_perf_indicator.ScoreRule, req.ScoreRule, condition.WithSkip(req.ScoreRule == "")).
		Equal(hrm_perf_indicator.DataSource, req.DataSource, condition.WithSkip(req.DataSource == "")).
		Equal(hrm_perf_indicator.Status, req.Status, condition.WithSkip(req.Status == "")).
		Build()...)
	if err != nil {
		return nil, err
	}

	records := make([]types.HrmPerfIndicator, 0, len(rows))
	for _, row := range rows {
		records = append(records, indicatorRecord(row))
	}
	return &types.ListResponse{
		PageResponse: types.PageResponse{
			Current: req.Current,
			Size:    req.Size,
			Total:   total,
		},
		Records: records,
	}, nil
}
