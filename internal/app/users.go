// Copyright Dose de Telemetria GmbH
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"net/http"
	"time"

	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/config"
	userhttp "github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/handler/http"
	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/store"
	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/store/memory"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

const name = "internal/app/users"

var (
	tracer = otel.Tracer(name)
	meter  = otel.Meter(name)
	logger = otelslog.NewLogger(name)
	reqCnt metric.Int64Counter     // contador de requisições HTTP
	reqDur metric.Float64Histogram // histograma para tempo de resposta
)

func init() {
	var err error
	reqCnt, err = meter.Int64Counter("http.server.requests", metric.WithDescription("Numero de requisicoes HTTP recebidas"))
	if err != nil {
		logger.Error("Erro ao criar contador de requisições", "error", err)
	}

	reqDur, err = meter.Float64Histogram("http.server.duration", metric.WithDescription("Duracao das requisicoes HTTP"))
	if err != nil {
		logger.Error("Erro ao criar histograma de duração", "error", err)
	}
	logger.Info("Métricas inicializadas com sucesso")
}

type User struct {
	Handler *userhttp.UserHandler
	Store   store.User
}

func NewUser(*config.Users) *User {
	store := memory.NewUserStore()
	return &User{
		Handler: userhttp.NewUserHandler(store),
		Store:   store,
	}
}

func (a *User) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /users", a.instrumentedHandler(a.Handler.List, "GET /users"))
	mux.HandleFunc("POST /users", a.instrumentedHandler(a.Handler.Create, "POST /users"))
	mux.HandleFunc("GET /users/{id}", a.instrumentedHandler(a.Handler.Get, "Get User"))
	mux.HandleFunc("PUT /users/{id}", a.instrumentedHandler(a.Handler.Update, "Update User"))
	mux.HandleFunc("DELETE /users/{id}", a.instrumentedHandler(a.Handler.Delete, "Delete User"))
}

func (a *User) instrumentedHandler(next http.HandlerFunc, operation string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), operation, trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()

		span.SetAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.url", r.URL.Path),
		)

		logger.InfoContext(ctx, "Handling API request", "method", r.Method, "url", r.URL.Path, "trace_id", span.SpanContext().TraceID().String())

		start := time.Now()
		next.ServeHTTP(w, r.WithContext(ctx))
		duration := time.Since(start).Seconds()

		reqCnt.Add(ctx, 1)
		reqDur.Record(ctx, duration)
	}
}
