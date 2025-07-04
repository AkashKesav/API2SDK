package utils

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Span represents a single operation within a trace
type Span struct {
	ID           string
	TraceID      string
	ParentID     string
	Name         string
	StartTime    time.Time
	EndTime      time.Time
	Duration     time.Duration
	Tags         map[string]string
	Events       []SpanEvent
	Status       string
	ErrorMessage string
	mutex        sync.Mutex
}

// SpanEvent represents an event that occurred during a span
type SpanEvent struct {
	Name      string
	Timestamp time.Time
	Tags      map[string]string
}

// Tracer manages the creation and collection of spans
type Tracer struct {
	spans  map[string]*Span
	mutex  sync.RWMutex
	logger *zap.Logger
}

// NewTracer creates a new tracer
func NewTracer(logger *zap.Logger) *Tracer {
	return &Tracer{
		spans:  make(map[string]*Span),
		logger: logger,
	}
}

// StartSpan starts a new span
func (t *Tracer) StartSpan(ctx context.Context, name string) (*Span, context.Context) {
	// Get parent span from context if it exists
	var parentID string
	var traceID string

	if parent := SpanFromContext(ctx); parent != nil {
		parentID = parent.ID
		traceID = parent.TraceID
	} else {
		// This is a root span, generate a new trace ID
		traceID = uuid.New().String()
	}

	// Create new span
	span := &Span{
		ID:        uuid.New().String(),
		TraceID:   traceID,
		ParentID:  parentID,
		Name:      name,
		StartTime: time.Now(),
		Tags:      make(map[string]string),
		Status:    "OK",
	}

	// Store span in tracer
	t.mutex.Lock()
	t.spans[span.ID] = span
	t.mutex.Unlock()

	// Create new context with span
	newCtx := ContextWithSpan(ctx, span)

	return span, newCtx
}

// FinishSpan marks a span as finished
func (t *Tracer) FinishSpan(span *Span) {
	if span == nil {
		return
	}

	span.mutex.Lock()
	defer span.mutex.Unlock()

	span.EndTime = time.Now()
	span.Duration = span.EndTime.Sub(span.StartTime)

	t.logger.Debug("Span finished",
		zap.String("span_id", span.ID),
		zap.String("trace_id", span.TraceID),
		zap.String("name", span.Name),
		zap.Duration("duration", span.Duration),
		zap.String("status", span.Status))
}

// GetSpan retrieves a span by ID
func (t *Tracer) GetSpan(spanID string) *Span {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.spans[spanID]
}

// GetTrace retrieves all spans for a trace
func (t *Tracer) GetTrace(traceID string) []*Span {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	var spans []*Span
	for _, span := range t.spans {
		if span.TraceID == traceID {
			spans = append(spans, span)
		}
	}

	return spans
}

// SetTag adds a tag to a span
func (s *Span) SetTag(key, value string) *Span {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.Tags[key] = value
	return s
}

// LogEvent adds an event to a span
func (s *Span) LogEvent(name string, tags map[string]string) *Span {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.Events = append(s.Events, SpanEvent{
		Name:      name,
		Timestamp: time.Now(),
		Tags:      tags,
	})

	return s
}

// SetError marks a span as having an error
func (s *Span) SetError(err error) *Span {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.Status = "ERROR"
	if err != nil {
		s.ErrorMessage = err.Error()
	}

	return s
}

// Finish marks a span as finished
func (s *Span) Finish() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.EndTime = time.Now()
	s.Duration = s.EndTime.Sub(s.StartTime)
}

type spanKey struct{}

// ContextWithSpan adds a span to a context
func ContextWithSpan(ctx context.Context, span *Span) context.Context {
	return context.WithValue(ctx, spanKey{}, span)
}

// SpanFromContext extracts a span from a context
func SpanFromContext(ctx context.Context) *Span {
	if ctx == nil {
		return nil
	}
	if span, ok := ctx.Value(spanKey{}).(*Span); ok {
		return span
	}
	return nil
}

// GlobalTracer is a singleton tracer
var (
	globalTracer     *Tracer
	globalTracerOnce sync.Once
)

// GetGlobalTracer returns the global tracer
func GetGlobalTracer(logger *zap.Logger) *Tracer {
	globalTracerOnce.Do(func() {
		globalTracer = NewTracer(logger)
	})
	return globalTracer
}

// TraceFunction wraps a function with tracing
func TraceFunction(ctx context.Context, name string, fn func(context.Context) error, logger *zap.Logger) error {
	tracer := GetGlobalTracer(logger)
	span, ctx := tracer.StartSpan(ctx, name)
	defer tracer.FinishSpan(span)

	err := fn(ctx)
	if err != nil {
		span.SetError(err)
	}

	return err
}

// TraceMiddleware creates middleware for tracing HTTP requests
func TraceMiddleware(logger *zap.Logger) func(next func(ctx context.Context) error) func(ctx context.Context) error {
	return func(next func(ctx context.Context) error) func(ctx context.Context) error {
		return func(ctx context.Context) error {
			tracer := GetGlobalTracer(logger)
			span, ctx := tracer.StartSpan(ctx, "http_request")
			defer tracer.FinishSpan(span)

			err := next(ctx)
			if err != nil {
				span.SetError(err)
			}

			return err
		}
	}
}
