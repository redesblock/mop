package tracer_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"testing"

	"github.com/redesblock/mop/core/log"
	"github.com/redesblock/mop/core/p2p"
	"github.com/redesblock/mop/core/tracer"
	"github.com/uber/jaeger-client-go"
)

func TestSpanFromHeaders(t *testing.T) {
	tracer, closer := newTracer(t)
	defer closer.Close()

	span, _, ctx := tracer.StartSpanFromContext(context.Background(), "some-operation", nil)
	defer span.Finish()

	headers := make(p2p.Headers)
	if err := tracer.AddContextHeader(ctx, headers); err != nil {
		t.Fatal(err)
	}

	gotSpanContext, err := tracer.FromHeaders(headers)
	if err != nil {
		t.Fatal(err)
	}

	if fmt.Sprint(gotSpanContext) == "" {
		t.Fatal("got empty span context")
	}

	wantSpanContext := span.Context()
	if fmt.Sprint(wantSpanContext) == "" {
		t.Fatal("got empty start span context")
	}

	if fmt.Sprint(gotSpanContext) != fmt.Sprint(wantSpanContext) {
		t.Errorf("got span context %+v, want %+v", gotSpanContext, wantSpanContext)
	}
}

func TestSpanWithContextFromHeaders(t *testing.T) {
	s, closer := newTracer(t)
	defer closer.Close()

	span, _, ctx := s.StartSpanFromContext(context.Background(), "some-operation", nil)
	defer span.Finish()

	headers := make(p2p.Headers)
	if err := s.AddContextHeader(ctx, headers); err != nil {
		t.Fatal(err)
	}

	ctx, err := s.WithContextFromHeaders(context.Background(), headers)
	if err != nil {
		t.Fatal(err)
	}

	gotSpanContext := tracer.FromContext(ctx)
	if fmt.Sprint(gotSpanContext) == "" {
		t.Fatal("got empty span context")
	}

	wantSpanContext := span.Context()
	if fmt.Sprint(wantSpanContext) == "" {
		t.Fatal("got empty start span context")
	}

	if fmt.Sprint(gotSpanContext) != fmt.Sprint(wantSpanContext) {
		t.Errorf("got span context %+v, want %+v", gotSpanContext, wantSpanContext)
	}
}

func TestFromContext(t *testing.T) {
	s, closer := newTracer(t)
	defer closer.Close()

	span, _, ctx := s.StartSpanFromContext(context.Background(), "some-operation", nil)
	defer span.Finish()

	wantSpanContext := span.Context()
	if fmt.Sprint(wantSpanContext) == "" {
		t.Fatal("got empty start span context")
	}

	gotSpanContext := tracer.FromContext(ctx)
	if fmt.Sprint(gotSpanContext) == "" {
		t.Fatal("got empty span context")
	}

	if fmt.Sprint(gotSpanContext) != fmt.Sprint(wantSpanContext) {
		t.Errorf("got span context %+v, want %+v", gotSpanContext, wantSpanContext)
	}
}

func TestWithContext(t *testing.T) {
	s, closer := newTracer(t)
	defer closer.Close()

	span, _, _ := s.StartSpanFromContext(context.Background(), "some-operation", nil)
	defer span.Finish()

	wantSpanContext := span.Context()
	if fmt.Sprint(wantSpanContext) == "" {
		t.Fatal("got empty start span context")
	}

	ctx := tracer.WithContext(context.Background(), span.Context())

	gotSpanContext := tracer.FromContext(ctx)
	if fmt.Sprint(gotSpanContext) == "" {
		t.Fatal("got empty span context")
	}

	if fmt.Sprint(gotSpanContext) != fmt.Sprint(wantSpanContext) {
		t.Errorf("got span context %+v, want %+v", gotSpanContext, wantSpanContext)
	}
}

func TestStartSpanFromContext_logger(t *testing.T) {
	s, closer := newTracer(t)
	defer closer.Close()

	buf := new(bytes.Buffer)

	span, logger, _ := s.StartSpanFromContext(context.Background(), "some-operation", log.NewLogger("test", log.WithSink(buf), log.WithJSONOutput()))
	defer span.Finish()

	wantTraceID := span.Context().(jaeger.SpanContext).TraceID()

	logger.Info("msg")
	data := make(map[string]interface{})
	if err := json.Unmarshal(buf.Bytes(), &data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	v, ok := data[tracer.LogField]
	if !ok {
		t.Fatalf("log field %q not found", tracer.LogField)
	}

	gotTraceID, ok := v.(string)
	if !ok {
		t.Fatalf("log field %q is not string", tracer.LogField)
	}

	if gotTraceID != wantTraceID.String() {
		t.Errorf("got trace id %q, want %q", gotTraceID, wantTraceID.String())
	}
}

func TestStartSpanFromContext_nilLogger(t *testing.T) {
	s, closer := newTracer(t)
	defer closer.Close()

	span, logger, _ := s.StartSpanFromContext(context.Background(), "some-operation", nil)
	defer span.Finish()

	if logger != nil {
		t.Error("logger is not nil")
	}
}

func TestNewLoggerWithTraceID(t *testing.T) {
	s, closer := newTracer(t)
	defer closer.Close()

	span, _, ctx := s.StartSpanFromContext(context.Background(), "some-operation", nil)
	defer span.Finish()

	buf := new(bytes.Buffer)

	logger := tracer.NewLoggerWithTraceID(ctx, log.NewLogger("test", log.WithSink(buf), log.WithJSONOutput()))

	wantTraceID := span.Context().(jaeger.SpanContext).TraceID()

	logger.Info("msg")
	data := make(map[string]interface{})
	if err := json.Unmarshal(buf.Bytes(), &data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	v, ok := data[tracer.LogField]
	if !ok {
		t.Fatalf("log field %q not found", tracer.LogField)
	}

	gotTraceID, ok := v.(string)
	if !ok {
		t.Fatalf("log field %q is not string", tracer.LogField)
	}

	if gotTraceID != wantTraceID.String() {
		t.Errorf("got trace id %q, want %q", gotTraceID, wantTraceID.String())
	}
}

func TestNewLoggerWithTraceID_nilLogger(t *testing.T) {
	s, closer := newTracer(t)
	defer closer.Close()

	span, _, ctx := s.StartSpanFromContext(context.Background(), "some-operation", nil)
	defer span.Finish()

	logger := tracer.NewLoggerWithTraceID(ctx, nil)

	if logger != nil {
		t.Error("logger is not nil")
	}
}

func newTracer(t *testing.T) (*tracer.Tracer, io.Closer) {
	t.Helper()

	tracer, closer, err := tracer.NewTracer(&tracer.Options{
		Enabled:     true,
		ServiceName: "test",
	})
	if err != nil {
		t.Fatal(err)
	}

	return tracer, closer
}
