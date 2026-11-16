package perf_assessment

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/perf_assessment"
)

type ConfirmPlan struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewConfirmPlan(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *ConfirmPlan {
	return &ConfirmPlan{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

func (l *ConfirmPlan) ConfirmPlan(req *types.ConfirmPlanRequest) (resp *types.ConfirmPlanResponse, err error) {
	actor, err := resolveActor(l.ctx, l.svcCtx)
	if err != nil {
		return nil, err
	}
	if err = wrapBizErr(ConfirmPlanForTask(l.ctx, l.svcCtx.SqlxConn, req.Uuid, req.Comment, actor)); err != nil {
		return nil, err
	}
	l.Infof("perf_assessment: 目标确认 assessment=%s", req.Uuid)
	return &types.ConfirmPlanResponse{}, nil
}
