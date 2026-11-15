package salary_scheme

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_scheme"
)

type GetSchemeItems struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewGetSchemeItems(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *GetSchemeItems {
	return &GetSchemeItems{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// GetSchemeItems HRM-37 试算消费端点：按方案取项目清单+计算规则（calc_order 升序，含覆盖值）。
// 共享实现 GetSchemeItems(ctx, svcCtx, schemeUuid) 供逻辑层兄弟包直接调用（无 HTTP 依赖）。
func (l *GetSchemeItems) GetSchemeItems(req *types.GetSchemeItemsRequest) (resp *types.GetSchemeItemsResponse, err error) {
	items, err := GetSchemeItemsForScheme(l.ctx, l.svcCtx, req.SchemeUuid)
	if err != nil {
		return nil, err
	}
	return &types.GetSchemeItemsResponse{Items: items}, nil
}
