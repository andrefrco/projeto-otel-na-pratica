// Copyright Dose de Telemetria GmbH
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/app"
	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/config"
)

func main() {
	log.Printf("Carregando arquivo de configuração")
	configFlag := flag.String("config", "", "path to the config file")
	flag.Parse()

	c, _ := config.LoadConfig(*configFlag)
	a := app.NewSubscription(&c.Subscriptions)
	a.RegisterRoutes(http.DefaultServeMux)
	log.Printf("Iniciando users na porta %s", c.Server.Endpoint.HTTP)
	_ = http.ListenAndServe(c.Server.Endpoint.HTTP, http.DefaultServeMux)
}
