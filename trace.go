package bloc_client

import (
	"context"

	"github.com/spf13/cast"
)

type TraceFlag string

const (
	TraceID TraceFlag = "trace_id"
	SpanID  TraceFlag = "span_id"
)

func NewSpanID() string {
	return NewUUID().String()
}

func SetTraceIDAndSpanIDToContext(traceID, spanID string) context.Context {
	ctx := context.WithValue(context.Background(), TraceID, traceID)
	return context.WithValue(ctx, SpanID, spanID)
}

func GetTraceIDFromContext(ctx context.Context) string {
	val := ctx.Value(TraceID)
	if val == nil {
		return ""
	}
	return cast.ToString(val)
}

func GetSpanIDFromContext(ctx context.Context) string {
	val := ctx.Value(SpanID)
	if val == nil {
		return ""
	}
	return cast.ToString(val)
}
