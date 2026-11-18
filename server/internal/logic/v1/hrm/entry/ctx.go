package entry

import (
	"context"

	"github.com/jzero-io/jzero-admin/server/internal/global"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
)

// svcCtxKey carries the *svc.ServiceContext (tests wire a real one via
// WithSvcCtx) into the approval OnApproved closure, which the bill base calls
// with only (ctx, session, billUuid) — same 约定 as offer 包 ctx.go /
// cardreplacement 包 ctx.go / leave 包 settle.go.
type svcCtxKey struct{}

// WithSvcCtx attaches the ServiceContext used by the approval write-back
// (自动建档 + 建档即开号, HRM-50 ADR 0007).
func WithSvcCtx(ctx context.Context, svcCtx *svc.ServiceContext) context.Context {
	return context.WithValue(ctx, svcCtxKey{}, svcCtx)
}

// svcCtxOf resolves the ServiceContext from ctx; falls back to the process-wide
// global (cmd/server.go 装配) when absent — server 运行态走全局，测试走 WithSvcCtx.
// 入职单建档/开号全部显式传 session (无需 offer/cardreplacement 的 TxBoundSvcCtx
// 整包换绑), 这里保持最小接缝。
func svcCtxOf(ctx context.Context) *svc.ServiceContext {
	if ctx != nil {
		if v, ok := ctx.Value(svcCtxKey{}).(*svc.ServiceContext); ok && v != nil {
			return v
		}
	}
	return &global.ServiceContext
}
