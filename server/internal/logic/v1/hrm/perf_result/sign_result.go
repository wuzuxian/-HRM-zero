package perf_result

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"

	hrmuser "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/hrmuser"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/perf_result"
)

type SignResult struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewSignResult(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *SignResult {
	return &SignResult{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// SignResult 签收结果（员工本人；幂等：已签收直接成功）。
func (l *SignResult) SignResult(req *types.SignResultRequest) (resp *types.SignResultResponse, err error) {
	employeeUuid, err := actorEmployeeUuid(l.ctx, l.svcCtx.SqlxConn)
	if err != nil {
		return nil, err
	}
	user := hrmuser.Current(l.ctx)
	if err := SignResultRow(l.ctx, l.svcCtx.SqlxConn, req.Uuid, employeeUuid, user.Username); err != nil {
		return nil, wrapBizErr(err)
	}
	l.Infof("perf_result: 签收结果 result=%s actor=%s", req.Uuid, user.Username)
	return &types.SignResultResponse{}, nil
}
