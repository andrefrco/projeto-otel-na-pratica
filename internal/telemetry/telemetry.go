// Copyright Dose de Telemetria GmbH
// SPDX-License-Identifier: Apache-2.0

package telemetry

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
	"go.opentelemetry.io/otel/trace"
)

// Start emits a span, an info log and a counter for the same operation.
func Start(ctx context.Context, scope, name string, kind trace.SpanKind) (context.Context, trace.Span) {
	ctx, span := otel.Tracer(scope).Start(ctx, name, trace.WithSpanKind(kind))
	otelslog.NewLogger(scope).InfoContext(ctx, name)
	counter, err := otel.Meter(scope).Int64Counter(scope + ".operations")
	if err == nil {
		counter.Add(ctx, 1, metric.WithAttributes(attribute.String("operation", name)))
	}
	return ctx, span
}

const defaultEndpoint = "http://localhost:4318"
const defaultServiceName = "projeto-otel-na-pratica"

var (
	tracerProvider *sdktrace.TracerProvider
	loggerProvider *sdklog.LoggerProvider
	meterProvider  *sdkmetric.MeterProvider
)

// Setup exports traces, logs and metrics over OTLP/HTTP.
// OTEL_EXPORTER_OTLP_ENDPOINT overrides the collector address.
// OTEL_SERVICE_NAME overrides the service name.
func Setup(ctx context.Context) error {
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = defaultEndpoint
	}

	opts := []resource.Option{
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		// deployment.environment.name and service.instance.id are omitted on purpose.
		resource.WithAttributes(
			// Discouraged resource attributes.
			semconv.HostIP("10.0.0.15"),
			semconv.OSDescription("Ubuntu 24.04.1 LTS (Noble Numbat)"),
			// http.request.method belongs on spans.
			semconv.HTTPRequestMethodGet,
			// HIGH is not one of mission_critical, high, medium, low.
			attribute.String("service.criticality", "HIGH"),
		),
	}
	if os.Getenv("OTEL_SERVICE_NAME") == "" {
		opts = append(opts, resource.WithAttributes(semconv.ServiceName(defaultServiceName)))
	}

	res, err := resource.New(ctx, opts...)
	if err != nil {
		return fmt.Errorf("otel resource: %w", err)
	}

	traceExp, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpointURL(endpoint))
	if err != nil {
		return fmt.Errorf("otel trace exporter: %w", err)
	}
	tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExp),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	logExp, err := otlploghttp.New(ctx, otlploghttp.WithEndpointURL(endpoint))
	if err != nil {
		return fmt.Errorf("otel log exporter: %w", err)
	}
	loggerProvider = sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExp)),
		sdklog.WithResource(res),
	)
	global.SetLoggerProvider(loggerProvider)

	metricExp, err := otlpmetrichttp.New(ctx, otlpmetrichttp.WithEndpointURL(endpoint))
	if err != nil {
		return fmt.Errorf("otel metric exporter: %w", err)
	}
	meterProvider = sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExp, sdkmetric.WithInterval(5*time.Second))),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(meterProvider)

	return nil
}

// Shutdown flushes the batch exporters.
func Shutdown(ctx context.Context) error {
	var err error
	if meterProvider != nil {
		err = errors.Join(err, meterProvider.Shutdown(ctx))
	}
	if loggerProvider != nil {
		err = errors.Join(err, loggerProvider.Shutdown(ctx))
	}
	if tracerProvider != nil {
		err = errors.Join(err, tracerProvider.Shutdown(ctx))
	}
	return err
}

// FlushOnStop exports the last batch when the process receives SIGINT or SIGTERM.
func FlushOnStop() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ctx.Done()
		stop()
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = Shutdown(shutCtx)
		os.Exit(0)
	}()
}
