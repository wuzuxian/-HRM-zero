package perf_period

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/perf_period"
)

type Edit struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewEdit(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Edit {
	return &Edit{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

func (l *Edit) Edit(req *types.EditRequest) (resp *types.EditResponse, err error) {
	if err = wrapBizErr(EditPeriod(l.ctx, l.svcCtx.SqlxConn, req.Uuid, req.Name, req.PeriodType,
		req.SchemeUuid, req.Scope, req.PeriodStart, req.PeriodEnd, req.ResultPublishDate, req.Remark)); err != nil {
		return nil, err
	}
	return &types.EditResponse{}, nil
}
