package monthreport

// Recalc 重新计算（demo-inventory 4.4 按钮）：重跑聚合写入月报表。
// 未锁定月=缓存刷新；锁定月=快照 diff 标 changed（快照数值不覆盖）。

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/monthreport"
)

type Recalc struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewRecalc(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Recalc {
	return &Recalc{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

func (l *Recalc) Recalc(req *types.RecalcRequest) (resp *types.RecalcResponse, err error) {
	locked, changedCount, err := RecalcMonth(l.ctx, l.svcCtx, req.YearMonth)
	if err != nil {
		return nil, err
	}
	return &types.RecalcResponse{Locked: locked, ChangedCount: changedCount}, nil
}
