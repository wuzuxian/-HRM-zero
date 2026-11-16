package perf_result

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/perf_result"
)

type SubmitApproval struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewSubmitApproval(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *SubmitApproval {
	return &SubmitApproval{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

func (l *SubmitApproval) SubmitApproval(req *types.SubmitApprovalRequest) (resp *types.SubmitApprovalResponse, err error) {
	uuid, err := SubmitPerfResultApproval(l.ctx, l.svcCtx.SqlxConn, l.svcCtx, req.PeriodUuid, req.Opinion)
	if err != nil {
		return nil, wrapBizErr(err)
	}
	l.Infof("perf_result: 提交批次结果审批 period=%s approval=%s", req.PeriodUuid, uuid)
	return &types.SubmitApprovalResponse{Uuid: uuid}, nil
}
