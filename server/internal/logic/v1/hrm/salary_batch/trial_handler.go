package salary_batch

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_batch"
)

type Trial struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewTrial(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Trial {
	return &Trial{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// Trial 薪资试算端点（引擎本体在 trial.go RunPayrollTrial；HRM-38/39 消费同一入口）。
func (l *Trial) Trial(req *types.TrialRequest) (resp *types.TrialResponse, err error) {
	return RunPayrollTrial(l.ctx, l.svcCtx, req.BatchUuid)
}
