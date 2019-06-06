package logutil

import "context"

var ctxKey = struct{}{}

func FromContext(ctx context.Context) (Log, bool) {
	val := ctx.Value(ctxKey)
	log, ok := val.(Log)
	return log, ok
}

func NewContext(ctx context.Context, log Log) context.Context {
	return context.WithValue(ctx, ctxKey, log)
}

func FromContextOrDefault(ctx context.Context) Log {
	if log, ok := FromContext(ctx); ok {
		return log
	}
	return Default()
}
