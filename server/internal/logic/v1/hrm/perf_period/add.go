package perf_period

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/perf_period"
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
	if _, err = AddPeriod(l.ctx, l.svcCtx.SqlxConn, req.Name, req.PeriodType, req.SchemeUuid, req.Scope,
		req.PeriodStart, req.PeriodEnd, req.ResultPublishDate, req.Remark); err != nil {
		return nil, wrapBizErr(err)
	}
	l.Infof("perf_period: 批次已创建 name=%s scheme=%s", req.Name, req.SchemeUuid)
	return &types.AddResponse{}, nil
}
