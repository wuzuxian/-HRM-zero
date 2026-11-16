package perf_period

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/perf_period"
)

type Advance struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewAdvance(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Advance {
	return &Advance{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

func (l *Advance) Advance(req *types.AdvanceRequest) (resp *types.AdvanceResponse, err error) {
	if err = wrapBizErr(AdvancePeriod(l.ctx, l.svcCtx.SqlxConn, req.Uuid)); err != nil {
		return nil, err
	}
	return &types.AdvanceResponse{}, nil
}
