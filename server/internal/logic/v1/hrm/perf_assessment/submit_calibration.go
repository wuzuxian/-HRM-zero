package perf_assessment

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/perf_assessment"
)

type SubmitCalibration struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewSubmitCalibration(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *SubmitCalibration {
	return &SubmitCalibration{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

func (l *SubmitCalibration) SubmitCalibration(req *types.SubmitCalibrationRequest) (resp *types.SubmitCalibrationResponse, err error) {
	actor, err := resolveActor(l.ctx, l.svcCtx.SqlxConn)
	if err != nil {
		return nil, err
	}
	if err = wrapBizErr(SubmitCalibrationForTask(l.ctx, l.svcCtx, req.Uuid, req.FinalScore, req.FinalGrade, req.Comment, actor)); err != nil {
		return nil, err
	}
	l.Infof("perf_assessment: 校准填报 assessment=%s final=%.2f", req.Uuid, req.FinalScore)
	return &types.SubmitCalibrationResponse{}, nil
}
