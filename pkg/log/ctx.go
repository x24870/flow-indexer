package log

import (
	"context"

	"go.uber.org/zap"
)

var Nop = zap.NewNop()

type ctxKey struct{}

// CtxKey is the key for logger stored in context.
func CtxKey() interface{} { return ctxKey{} }

// Context gets from context or creates a logger with tags.
func Context(ctx context.Context, names ...string) (context.Context, *zap.Logger) {
	l := GetFromCtx(ctx)
	if l != Nop {
		return ctx, l
	}
	l = zap.L()
	for _, n := range names {
		l = l.Named(n)
	}
	ctx = context.WithValue(ctx, CtxKey(), l)
	return ctx, l
}

// GetFromCtx returns logger from context, nil for no logger found.
func GetFromCtx(ctx context.Context) *zap.Logger {
	val := ctx.Value(CtxKey())
	if val == nil {
		return Nop
	}
	return val.(*zap.Logger)
}
