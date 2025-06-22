package trace

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// Init configures a stdout exporter and sets the global tracer provider.
// It returns a shutdown function to flush traces when the program exits.
func Init() (func(context.Context) error, error) {
	exp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
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
