package perf_assessment

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/perf_assessment"
)

type Generate struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewGenerate(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Generate {
	return &Generate{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// Generate 批次发布后生成考核任务（幂等：已存在者跳过，补齐新增参与人）。
func (l *Generate) Generate(req *types.GenerateRequest) (resp *types.GenerateResponse, err error) {
	count, err := GenerateAssessments(l.ctx, l.svcCtx.SqlxConn, l.svcCtx, req.PeriodUuid)
	if err != nil {
		return nil, wrapBizErr(err)
	}
	l.Infof("perf_assessment: 批次生成考核任务 period=%s count=%d", req.PeriodUuid, count)
	return &types.GenerateResponse{GeneratedCount: count}, nil
}
