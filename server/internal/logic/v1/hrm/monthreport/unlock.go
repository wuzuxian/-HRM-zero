package monthreport

// Unlock 解锁月报（管理员）：删除快照恢复实时聚合 + 流水。

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/monthreport"
)

type Unlock struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewUnlock(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Unlock {
	return &Unlock{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

func (l *Unlock) Unlock(req *types.UnlockRequest) (resp *types.UnlockResponse, err error) {
	if err := UnlockMonth(l.ctx, l.svcCtx, req.YearMonth, currentUserUuid(l.ctx), currentUserName(l.ctx), req.Remark); err != nil {
		return nil, err
	}
	return &types.UnlockResponse{}, nil
}
