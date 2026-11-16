package perf_result

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/perf_result"
)

type Dashboard struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewDashboard(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Dashboard {
	return &Dashboard{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// Dashboard 绩效驾驶舱（HRM-46，demo-inventory 9.3 /hrm/performance/dashboard）：
// 单聚合端点 — 指标卡 + 风险预警明细（低分/逾期任务/申诉中名单）。口径见
// dashboard_domain.go 与 hrm-46-notes.md。
func (l *Dashboard) Dashboard(req *types.HrmPerfDashboardRequest) (resp *types.DashboardResponse, err error) {
	return BuildDashboard(l.ctx, l.svcCtx.SqlxConn, req)
}
