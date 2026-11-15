package perf_template

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/perf_template"
)

type GetDetail struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewGetDetail(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *GetDetail {
	return &GetDetail{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// GetDetail returns one 考核模板 with its 指标权重配置 (模板内配置指标与权重).
func (l *GetDetail) GetDetail(req *types.GetDetailRequest) (resp *types.GetDetailResponse, err error) {
	detail, err := GetTemplateDetail(l.ctx, l.svcCtx.SqlxConn, req.Uuid)
	if err != nil {
		return nil, wrapBizErr(err)
	}

	items := make([]types.HrmPerfTemplateItem, 0, len(detail.Items))
	for _, item := range detail.Items {
		items = append(items, types.HrmPerfTemplateItem{
			Uuid:          item.Uuid,
			IndicatorUuid: item.IndicatorUuid,
			Weight:        item.Weight,
			IndicatorCode: item.IndicatorCode,
			IndicatorName: item.IndicatorName,
			IndicatorType: item.IndicatorType,
			Unit:          item.Unit,
			ScoreRule:     item.ScoreRule,
			DataSource:    item.DataSource,
		})
	}
	return &types.GetDetailResponse{
		HrmPerfTemplate: templateRecord(detail.HrmPerfTemplate, items),
	}, nil
}
