package bloc_client

import (
	"context"

	"github.com/spf13/cast"
)

type TraceFlag string

const (
	TraceID TraceFlag = "trace_id"
)

func SetTraceIDToContext(traceID string) context.Context {
	return context.WithValue(context.Background(), TraceID, traceID)
}

func GetTraceIDFromContext(ctx context.Context) string {
	val := ctx.Value(TraceID)
	if val == nil {
		return ""
	}
	return cast.ToString(val)
}
