package perf_template

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/perf_template"
)

type Add struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewAdd(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Add {
	return &Add{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

func (l *Add) Add(req *types.AddRequest) (resp *types.AddResponse, err error) {
	err = wrapBizErr(AddTemplate(l.ctx, l.svcCtx.SqlxConn, req.Name, req.Status, req.Remark, toLogicItems(req.Items)))
	if err != nil {
		return nil, err
	}
	return &types.AddResponse{}, nil
}

// toLogicItems converts the API 指标权重配置 into the domain TemplateItem.
func toLogicItems(items []types.TemplateItemRequest) []TemplateItem {
	logicItems := make([]TemplateItem, 0, len(items))
	for _, item := range items {
		logicItems = append(logicItems, TemplateItem{
			IndicatorUuid: item.IndicatorUuid,
			Weight:        item.Weight,
		})
	}
	return logicItems
}
