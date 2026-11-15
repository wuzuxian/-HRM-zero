package cardreplacement

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"

	coresvc "github.com/jzero-io/jzero-admin/core-engine/svc"
	"github.com/jzero-io/jzero-admin/server/internal/global"
	"github.com/jzero-io/jzero-admin/server/internal/model"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
)

// svcCtxKey carries the *svc.ServiceContext (tests wire a real one via
// WithSvcCtx) into the approval OnApproved closure, which the bill base calls
// with only (ctx, session, billUuid) — same约定 as leave 包 settle.go.
type svcCtxKey struct{}

// WithSvcCtx attaches the ServiceContext used by the approval write-back
// (补卡写打卡行 + SettleDay 重算日结).
func WithSvcCtx(ctx context.Context, svcCtx *svc.ServiceContext) context.Context {
	return context.WithValue(ctx, svcCtxKey{}, svcCtx)
}

// svcCtxOf resolves the ServiceContext from ctx; falls back to the process-wide
// global (cmd/server.go 装配) when absent — server 运行态走全局，测试走 WithSvcCtx.
func svcCtxOf(ctx context.Context) *svc.ServiceContext {
	if ctx != nil {
		if v, ok := ctx.Value(svcCtxKey{}).(*svc.ServiceContext); ok && v != nil {
			return v
		}
	}
	return &global.ServiceContext
}

// TxBoundSvcCtx returns a shallow copy of svcCtx whose Model is rebuilt on the
// approval transaction session, so nested domain calls (punch.SettleDay) read
// AND write inside the same tx. Rule-resolution reads also hit the session —
// the tx has already written the punch row, which is exactly the state the
// settle re-run must see.
func TxBoundSvcCtx(ctx context.Context, session sqlx.Session, base *svc.ServiceContext) *svc.ServiceContext {
	if session == nil {
		return base
	}
	conn := sqlx.NewSqlConnFromSession(session)
	out := *base
	out.ServiceContext = &coresvc.ServiceContext{SqlxConn: conn}
	out.Model = model.NewModel(conn)
	return &out
}
