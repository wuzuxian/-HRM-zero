package perf_assessment

import "context"

// codesKey carries the current user's button/API permission codes through ctx.
// Authx middleware already casbin-enforced the endpoint code at the gateway;
// the business "who may fill which stage" decision needs the coarser codes
// (review/calibrate), which handlers pass via WithCodes. Tests construct
// actorContext directly.
type codesKey struct{}

// WithCodes returns a ctx carrying permission codes.
func WithCodes(ctx context.Context, codes []string) context.Context {
	return context.WithValue(ctx, codesKey{}, codes)
}

// codesFromCtx reads the codes injected by WithCodes (nil when absent).
func codesFromCtx(ctx context.Context) []string {
	if v, ok := ctx.Value(codesKey{}).([]string); ok {
		return v
	}
	return nil
}
