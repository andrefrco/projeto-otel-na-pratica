// Copyright Dose de Telemetria GmbH
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/app"
	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/config"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	otelconfig "go.opentelemetry.io/contrib/config/v0.3.0"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
)

const name = "cmd/users"

func main() {
	ctx := context.Background()

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	otelShutdown, err := setupOTel(ctx)
	if err != nil {
		log.Fatal("Failed to start OpenTelemetry setup", err)
	}

	logger := otelslog.NewLogger(name)
	logger.InfoContext(ctx, "OpenTelemetry SDK configured successfully")

	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
		logger.ErrorContext(ctx, "Error shutting down OpenTelemetry", "error", err)
		os.Exit(1)
	}()

	configFlag := flag.String("config", "config.yaml", "path to the config file")
	flag.Parse()
	c, err := config.LoadConfig(*configFlag)
	if err != nil {
		logger.ErrorContext(ctx, "Error loading app configurations", "error", err)
		os.Exit(1)
	}

	logger.InfoContext(ctx, "Application configuration loaded successfully", "port", c.Server.Endpoint.HTTP)

	a := app.NewUser(&c.Users)
	a.RegisterRoutes(http.DefaultServeMux)

	logger.InfoContext(ctx, "User service initialized and routes registered successfully")

	err = http.ListenAndServe(c.Server.Endpoint.HTTP, http.DefaultServeMux)
	if err != nil {
		logger.ErrorContext(ctx, "Error starting the HTTP server", "error", err)
		os.Exit(1)
	}
	logger.InfoContext(ctx, "HTTP server started successfully and is listening for requests")
}

func setupOTel(ctx context.Context) (func(context.Context) error, error) {
	confFlag := flag.String("otelConfig", "config-otel.yaml", "otel config file to parse")
	flag.Parse()
	b, err := os.ReadFile(*confFlag)
	if err != nil {
		return nil, err
	}
	conf, err := otelconfig.ParseYAML(b)
	if err != nil {
		return nil, err
	}
	sdk, err := otelconfig.NewSDK(otelconfig.WithContext(ctx), otelconfig.WithOpenTelemetryConfiguration(*conf))
	if err != nil {
		return nil, err
	}
	otel.SetTracerProvider(sdk.TracerProvider())
	otel.SetMeterProvider(sdk.MeterProvider())
	global.SetLoggerProvider(sdk.LoggerProvider())
	return sdk.Shutdown, nil
}
