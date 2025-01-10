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
	otelconfig "go.opentelemetry.io/contrib/config/v0.3.0"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
)

func main() {
	ctx := context.Background()

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	otelShutdown, err := setupOTel(ctx)
	if err != nil {
		log.Fatalf("Error to setup otel: %v\n", err)
	}

	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
		log.Fatalf("Error to shutdown otel: %v\n", err)
	}()

	configFlag := flag.String("config", "config.yaml", "path to the config file")
	flag.Parse()
	c, err := config.LoadConfig(*configFlag)
	if err != nil {
		log.Fatalf("Erro ao carregar configurações do app: %v", err)
	}

	a := app.NewUser(&c.Users)
	a.RegisterRoutes(http.DefaultServeMux)
	log.Printf("Iniciando users na porta %s", c.Server.Endpoint.HTTP)
	_ = http.ListenAndServe(c.Server.Endpoint.HTTP, http.DefaultServeMux)
}

func setupOTel(ctx context.Context) (func(context.Context) error, error) {
	confFlag := flag.String("config", "config-otel.yaml", "otel config file to parse")
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
