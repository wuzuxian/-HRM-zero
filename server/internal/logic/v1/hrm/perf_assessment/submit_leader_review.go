package perf_assessment

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/perf_assessment"
)

type SubmitLeaderReview struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewSubmitLeaderReview(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *SubmitLeaderReview {
	return &SubmitLeaderReview{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

func (l *SubmitLeaderReview) SubmitLeaderReview(req *types.SubmitLeaderReviewRequest) (resp *types.SubmitLeaderReviewResponse, err error) {
	actor, err := resolveActor(l.ctx, l.svcCtx)
	if err != nil {
		return nil, err
	}
	items := make([]ScoreItem, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, ScoreItem{IndicatorUuid: item.IndicatorUuid, Score: item.Score})
	}
	if err = wrapBizErr(SubmitLeaderReviewForTask(l.ctx, l.svcCtx, req.Uuid, items, req.Comment, actor)); err != nil {
		return nil, err
	}
	l.Infof("perf_assessment: 上级评审 assessment=%s", req.Uuid)
	return &types.SubmitLeaderReviewResponse{}, nil
}
