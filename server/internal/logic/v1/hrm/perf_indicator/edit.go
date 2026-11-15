package perf_indicator

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/perf_indicator"
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
	err = wrapBizErr(EditIndicator(l.ctx, l.svcCtx.SqlxConn, Indicator{
		Uuid:          req.Uuid,
		Code:          req.Code,
		Name:          req.Name,
		IndicatorType: req.IndicatorType,
		Unit:          req.Unit,
		ScoreRule:     req.ScoreRule,
		DataSource:    req.DataSource,
		DefaultWeight: req.DefaultWeight,
		Status:        req.Status,
		Remark:        req.Remark,
	}))
	if err != nil {
		return nil, err
	}
	return &types.EditResponse{}, nil
}
