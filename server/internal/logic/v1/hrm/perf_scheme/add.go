package perf_scheme

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/perf_scheme"
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
	err = wrapBizErr(AddScheme(l.ctx, l.svcCtx.SqlxConn, req.Name, req.TemplateUuid, req.PeriodType,
		req.AssessmentMode, req.ScoreMode, req.PayLinkMode, req.Status, req.Remark))
	if err != nil {
		return nil, err
	}
	return &types.AddResponse{}, nil
}
