package offer

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
// with only (ctx, session, billUuid) — same 约定 as leave 包 settle.go /
// cardreplacement 包 ctx.go.
type svcCtxKey struct{}

// WithSvcCtx attaches the ServiceContext used by the approval 三联动 write-back
// (职位已录用+1 / 候选人→hired / 草稿入职单).
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
// approval transaction session, so the 三联动 nested model calls (IncreaseHiredCountTx /
// candidate row update / entry bill insert) read AND write inside the same tx.
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
