package utils

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// TraceContext holds tracing information for a request
type TraceContext struct {
	TraceID   string
	SpanID    string
	ParentID  string
	StartTime time.Time
	Tags      map[string]string
	Logger    *zap.Logger
}

// NewTraceContext creates a new trace context
func NewTraceContext(logger *zap.Logger) *TraceContext {
	return &TraceContext{
		TraceID:   generateTraceID(),
		SpanID:    generateSpanID(),
		StartTime: time.Now(),
		Tags:      make(map[string]string),
		Logger:    logger,
	}
}

// NewChildSpan creates a child span from the current trace context
func (tc *TraceContext) NewChildSpan(operation string) *TraceContext {
	child := &TraceContext{
		TraceID:   tc.TraceID,
		SpanID:    generateSpanID(),
		ParentID:  tc.SpanID,
		StartTime: time.Now(),
		Tags:      make(map[string]string),
		Logger:    tc.Logger,
	}
	
	child.SetTag("operation", operation)
	return child
}

// SetTag adds a tag to the trace context
func (tc *TraceContext) SetTag(key, value string) {
	tc.Tags[key] = value
}

// LogInfo logs an info message with trace context
func (tc *TraceContext) LogInfo(message string, fields ...zap.Field) {
	fields = append(fields, 
		zap.String("trace_id", tc.TraceID),
		zap.String("span_id", tc.SpanID),
		zap.String("parent_id", tc.ParentID),
	)
	tc.Logger.Info(message, fields...)
}

// LogError logs an error message with trace context
func (tc *TraceContext) LogError(message string, err error, fields ...zap.Field) {
	fields = append(fields, 
		zap.String("trace_id", tc.TraceID),
		zap.String("span_id", tc.SpanID),
		zap.String("parent_id", tc.ParentID),
		zap.Error(err),
	)
	tc.Logger.Error(message, fields...)
}

// LogDebug logs a debug message with trace context
func (tc *TraceContext) LogDebug(message string, fields ...zap.Field) {
	fields = append(fields, 
		zap.String("trace_id", tc.TraceID),
		zap.String("span_id", tc.SpanID),
		zap.String("parent_id", tc.ParentID),
	)
	tc.Logger.Debug(message, fields...)
}

// Finish completes the span and logs the duration
func (tc *TraceContext) Finish() {
	duration := time.Since(tc.StartTime)
	tc.LogInfo("Span finished", 
		zap.Duration("duration", duration),
		zap.Any("tags", tc.Tags),
	)
}

// generateTraceID generates a unique trace ID
func generateTraceID() string {
	return uuid.New().String()
}

// generateSpanID generates a unique span ID
func generateSpanID() string {
	return uuid.New().String()[:8]
}

// TraceContextKey is the key used to store trace context in context.Context
type TraceContextKey struct{}

// WithTraceContext adds a trace context to a context.Context
func WithTraceContext(ctx context.Context, tc *TraceContext) context.Context {
	return context.WithValue(ctx, TraceContextKey{}, tc)
}

// GetTraceContext retrieves a trace context from a context.Context
func GetTraceContext(ctx context.Context) (*TraceContext, bool) {
	tc, ok := ctx.Value(TraceContextKey{}).(*TraceContext)
	return tc, ok
}