package perf_assessment

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/perf_assessment"
)

type HandleAppeal struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewHandleAppeal(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *HandleAppeal {
	return &HandleAppeal{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// HandleAppeal 申诉裁决（校准人：维持/调整得分→关闭申诉）。
func (l *HandleAppeal) HandleAppeal(req *types.HandleAppealRequest) (resp *types.HandleAppealResponse, err error) {
	actor, err := resolveActor(l.ctx, l.svcCtx.SqlxConn)
	if err != nil {
		return nil, err
	}
	if err = wrapBizErr(HandleAppealForTask(l.ctx, l.svcCtx.SqlxConn, req.Uuid, req.Decision, req.Comment,
		req.AdjustedScore, req.AdjustedGrade, actor)); err != nil {
		return nil, err
	}
	l.Infof("perf_assessment: 申诉裁决 appeal=%s decision=%s", req.Uuid, req.Decision)
	return &types.HandleAppealResponse{}, nil
}
