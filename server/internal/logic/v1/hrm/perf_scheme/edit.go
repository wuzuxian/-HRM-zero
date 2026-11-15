package perf_scheme

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/perf_scheme"
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
	err = wrapBizErr(EditScheme(l.ctx, l.svcCtx.SqlxConn, req.Uuid, req.Name, req.TemplateUuid, req.PeriodType,
		req.AssessmentMode, req.ScoreMode, req.PayLinkMode, req.Status, req.Remark))
	if err != nil {
		return nil, err
	}
	return &types.EditResponse{}, nil
}
