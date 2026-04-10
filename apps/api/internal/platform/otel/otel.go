// Package otel provides OpenTelemetry tracer and log provider initialization
// with OTLP gRPC exporters for the Zenvikar API.
package otel

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// Providers holds the initialized OTel providers for cleanup.
type Providers struct {
	Tracer *sdktrace.TracerProvider
	Logger *sdklog.LoggerProvider
}

// Shutdown gracefully shuts down all providers.
func (p *Providers) Shutdown(ctx context.Context) {
	if p.Tracer != nil {
		p.Tracer.Shutdown(ctx)
	}
	if p.Logger != nil {
		p.Logger.Shutdown(ctx)
	}
}

// Init initializes OpenTelemetry trace and log providers with OTLP gRPC
// exporters pointing at the given endpoint. Pass an empty endpoint to get
// no-op providers.
func Init(endpoint string) (*Providers, error) {
	if endpoint == "" {
		tp := sdktrace.NewTracerProvider()
		otel.SetTracerProvider(tp)
		return &Providers{Tracer: tp}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("zenvikar-api"),
		),
	)
	if err != nil {
		return nil, err
	}

	// Trace exporter
	traceExp, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExp),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	// Log exporter
	logExp, err := otlploggrpc.New(ctx,
		otlploggrpc.WithEndpoint(endpoint),
		otlploggrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	lp := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExp)),
		sdklog.WithResource(res),
	)

	return &Providers{Tracer: tp, Logger: lp}, nil
}
