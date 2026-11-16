package monthreport

// Lock 锁定月报（demo-inventory 4.4 按钮）：当月实时聚合写入快照 + 流水。

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/monthreport"
)

type Lock struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewLock(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Lock {
	return &Lock{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

func (l *Lock) Lock(req *types.LockRequest) (resp *types.LockResponse, err error) {
	if err := LockMonth(l.ctx, l.svcCtx, req.YearMonth, currentUserUuid(l.ctx), currentUserName(l.ctx), req.Remark); err != nil {
		return nil, err
	}
	return &types.LockResponse{}, nil
}
