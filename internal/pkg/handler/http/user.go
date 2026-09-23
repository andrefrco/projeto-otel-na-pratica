// Copyright Dose de Telemetria GmbH
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/model"
	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/store"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// UserHandler is an HTTP handler that performs CRUD operations for model.User using a store.User
type UserHandler struct {
	store store.User
}

// NewUserHandler returns a new UserHandler
func NewUserHandler(store store.User) *UserHandler {
	return &UserHandler{
		store: store,
	}
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("users").Start(r.Context(), "user.list", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	users, err := h.store.List(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(users)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	user := &model.User{}
	if err := json.NewDecoder(r.Body).Decode(user); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	ctx, span := otel.Tracer("users").Start(r.Context(), "user.create", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()
	span.SetAttributes(
		attribute.String("user.email", user.Email),
		attribute.String("user.address", user.Address),
		attribute.String("service.name", "users"),
	)

	created, err := h.store.Create(ctx, user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	otelslog.NewLogger("users").InfoContext(ctx, "user created at "+user.Address, slog.String("email", user.Email))

	err = json.NewEncoder(w).Encode(created)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	// Span name includes the id.
	ctx, span := otel.Tracer("users").Start(r.Context(), "GET /users/"+id, trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	user, err := h.store.Get(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	user := &model.User{}
	if err := json.NewDecoder(r.Body).Decode(user); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	ctx, span := otel.Tracer("users").Start(r.Context(), "user.update", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	updatedSubscription, err := h.store.Update(ctx, user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(updatedSubscription)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ctx, span := otel.Tracer("users").Start(r.Context(), "user.delete", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	err := h.store.Delete(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
