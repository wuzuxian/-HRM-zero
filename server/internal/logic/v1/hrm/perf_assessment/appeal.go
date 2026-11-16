package perf_assessment

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/perf_assessment"
)

type Appeal struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewAppeal(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Appeal {
	return &Appeal{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// Appeal 发起申诉（仅结果确认阶段、被考核人本人；原因必填）。
func (l *Appeal) Appeal(req *types.AppealRequest) (resp *types.AppealResponse, err error) {
	actor, err := resolveActor(l.ctx, l.svcCtx.SqlxConn)
	if err != nil {
		return nil, err
	}
	if err = wrapBizErr(AppealForTask(l.ctx, l.svcCtx.SqlxConn, req.AssessmentUuid, req.Reason, actor)); err != nil {
		return nil, err
	}
	l.Infof("perf_assessment: 发起申诉 assessment=%s", req.AssessmentUuid)
	return &types.AppealResponse{}, nil
}
