package perf_assessment

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/perf_assessment"
)

type ConfirmResult struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewConfirmResult(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *ConfirmResult {
	return &ConfirmResult{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

func (l *ConfirmResult) ConfirmResult(req *types.ConfirmResultRequest) (resp *types.ConfirmResultResponse, err error) {
	actor, err := resolveActor(l.ctx, l.svcCtx.SqlxConn)
	if err != nil {
		return nil, err
	}
	if err = wrapBizErr(ConfirmResultForTask(l.ctx, l.svcCtx, req.Uuid, actor)); err != nil {
		return nil, err
	}
	l.Infof("perf_assessment: 结果确认 assessment=%s", req.Uuid)
	return &types.ConfirmResultResponse{}, nil
}
