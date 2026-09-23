// Copyright Dose de Telemetria GmbH
// SPDX-License-Identifier: Apache-2.0

package telemetry

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
)

const defaultEndpoint = "http://localhost:4318"
const defaultServiceName = "projeto-otel-na-pratica"

// Setup exports traces and logs over OTLP/HTTP.
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
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExp),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	logExp, err := otlploghttp.New(ctx, otlploghttp.WithEndpointURL(endpoint))
	if err != nil {
		return fmt.Errorf("otel log exporter: %w", err)
	}
	lp := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExp)),
		sdklog.WithResource(res),
	)
	global.SetLoggerProvider(lp)

	metricExp, err := otlpmetrichttp.New(ctx, otlpmetrichttp.WithEndpointURL(endpoint))
	if err != nil {
		return fmt.Errorf("otel metric exporter: %w", err)
	}
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExp, sdkmetric.WithInterval(5*time.Second))),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(mp)

	return nil
}
