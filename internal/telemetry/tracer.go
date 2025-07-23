package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

var (
	TracerEndpoint = "127.0.0.1:4317"
	ServiceName    = "govf-service"
)

func InitTracer(ctx context.Context) (*sdktrace.TracerProvider, error) {
	exp, err := otlptracegrpc.New(ctx, otlptracegrpc.WithEndpoint(TracerEndpoint), otlptracegrpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	res, _ := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(ServiceName),
		),
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	return tp, nil
}

func TraceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	sc := span.SpanContext()
	if !sc.IsValid() {
		return "none"
	}
	return sc.TraceID().String()
}

func InitMeter() metric.Meter {
	return otel.Meter(ServiceName)
}
