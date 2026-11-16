package salary_confirm

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_confirm"
)

type Submit struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewSubmit(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Submit {
	return &Submit{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// Submit 提交薪资确认审批端点（守卫与事务在 approval.go SubmitSalaryConfirm）。
func (l *Submit) Submit(req *types.SubmitRequest) (resp *types.SubmitResponse, err error) {
	uuid, err := SubmitSalaryConfirm(l.ctx, l.svcCtx.SqlxConn, l.svcCtx, req.BatchUuid, req.Opinion)
	if err != nil {
		return nil, err
	}
	l.Infof("salary_confirm: 提交薪资确认审批 batch=%s bill=%s", req.BatchUuid, uuid)
	return &types.SubmitResponse{Uuid: uuid}, nil
}
