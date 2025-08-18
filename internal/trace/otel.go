package trace

import (
	"context"
	"fmt"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// Init configures a stdout exporter and sets the global tracer provider.
// It returns a shutdown function to flush traces when the program exits.
func Init(collector string) (func(context.Context) error, error) {
	var (
		exp sdktrace.SpanExporter
		err error
	)
	ctx := context.Background()
	if collector != "" {
		addr := collector
		if strings.HasPrefix(addr, "http://") {
			addr = strings.TrimPrefix(addr, "http://")
		}
		client := otlptracehttp.NewClient(
			otlptracehttp.WithEndpoint(addr),
			otlptracehttp.WithInsecure(),
		)
		exp, err = otlptrace.New(ctx, client)
	} else {
		exp, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
	}
	if err != nil {
		return nil, err
	}
	tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exp))
	otel.SetTracerProvider(tp)
	return tp.Shutdown, nil
}

// OTelWriter forwards trace events to OpenTelemetry.
type OTelWriter struct {
	tracer oteltrace.Tracer
}

func NewOTel() *OTelWriter { return &OTelWriter{tracer: otel.Tracer("agentry")} }

func (o *OTelWriter) Write(ctx context.Context, e Event) {
	span := oteltrace.SpanFromContext(ctx)
	if !span.IsRecording() {
		var s oteltrace.Span
		ctx, s = o.tracer.Start(ctx, string(e.Type))
		s.SetAttributes(
			attribute.String("agent_id", e.AgentID),
			attribute.String("data", fmt.Sprintf("%v", e.Data)),
		)
		s.End()
		return
	}
	span.AddEvent(string(e.Type), oteltrace.WithAttributes(
		attribute.String("agent_id", e.AgentID),
		attribute.String("data", fmt.Sprintf("%v", e.Data)),
	))
}

// Export sends a slice of trace events as individual spans using the global tracer.
func Export(ctx context.Context, events []Event) {
	tr := otel.Tracer("agentry")
	for _, e := range events {
		_, span := tr.Start(ctx, string(e.Type))
		span.SetAttributes(
			attribute.String("agent_id", e.AgentID),
			attribute.String("data", fmt.Sprintf("%v", e.Data)),
		)
		span.End()
	}
}
