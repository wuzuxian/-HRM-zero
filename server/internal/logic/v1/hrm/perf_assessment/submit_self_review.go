package perf_assessment

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/perf_assessment"
)

type SubmitSelfReview struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewSubmitSelfReview(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *SubmitSelfReview {
	return &SubmitSelfReview{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

func (l *SubmitSelfReview) SubmitSelfReview(req *types.SubmitSelfReviewRequest) (resp *types.SubmitSelfReviewResponse, err error) {
	actor, err := resolveActor(l.ctx, l.svcCtx.SqlxConn)
	if err != nil {
		return nil, err
	}
	items := make([]ScoreItem, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, ScoreItem{IndicatorUuid: item.IndicatorUuid, Score: item.Score})
	}
	if err = wrapBizErr(SubmitSelfReviewForTask(l.ctx, l.svcCtx, req.Uuid, items, req.Comment, actor)); err != nil {
		return nil, err
	}
	l.Infof("perf_assessment: 自评填报 assessment=%s", req.Uuid)
	return &types.SubmitSelfReviewResponse{}, nil
}
