package perf_period

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/perf_period"
)

type Publish struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewPublish(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Publish {
	return &Publish{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

func (l *Publish) Publish(req *types.PublishRequest) (resp *types.PublishResponse, err error) {
	if err = wrapBizErr(PublishPeriod(l.ctx, l.svcCtx.SqlxConn, req.Uuid)); err != nil {
		return nil, err
	}
	l.Infof("perf_period: 批次已发布 period=%s", req.Uuid)
	return &types.PublishResponse{}, nil
}
