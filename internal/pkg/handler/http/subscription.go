// Copyright Dose de Telemetria GmbH
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/model"
	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/store"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// SubscriptionHandler is an HTTP handler that performs CRUD operations for model.Subscription using a store.Subscription
type SubscriptionHandler struct {
	store         store.Subscription
	usersEndpoint string
	plansEndpoint string
}

// NewSubscriptionHandler returns a new SubscriptionHandler
func NewSubscriptionHandler(store store.Subscription, usersEndpoint string, plansEndpoint string) *SubscriptionHandler {
	return &SubscriptionHandler{
		store:         store,
		usersEndpoint: usersEndpoint,
		plansEndpoint: plansEndpoint,
	}
}

func (h *SubscriptionHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("subscriptions").Start(r.Context(), "subscription.list", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	subscriptions, err := h.store.List(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(subscriptions)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *SubscriptionHandler) Create(w http.ResponseWriter, r *http.Request) {
	subscription := &model.Subscription{}
	if err := json.NewDecoder(r.Body).Decode(subscription); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	ctx, span := otel.Tracer("subscriptions").Start(r.Context(), "subscription.create", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()
	span.SetAttributes(attribute.String("enduser.id", subscription.UserID))

	// verify the user exists
	{
		// Client span is a new root, and the name includes the user id.
		_, clientSpan := otel.Tracer("subscriptions").Start(context.Background(), "GET "+h.usersEndpoint+"/"+subscription.UserID, trace.WithSpanKind(trace.SpanKindClient))
		user, _ := http.Get(h.usersEndpoint + "/" + subscription.UserID)
		clientSpan.End()
		if user.StatusCode != http.StatusOK {
			http.Error(w, "User not found", http.StatusBadRequest)
			return
		}
		defer user.Body.Close()
	}

	// verify the plan exists
	{
		_, clientSpan := otel.Tracer("subscriptions").Start(context.Background(), "GET "+h.plansEndpoint+"/"+subscription.PlanID, trace.WithSpanKind(trace.SpanKindClient))
		plan, _ := http.Get(h.plansEndpoint + "/" + subscription.PlanID)
		clientSpan.End()
		if plan.StatusCode != http.StatusOK {
			http.Error(w, "Plan not found", http.StatusBadRequest)
			return
		}
		defer plan.Body.Close()
	}

	created, err := h.store.Create(ctx, subscription)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(created)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *SubscriptionHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ctx, span := otel.Tracer("subscriptions").Start(r.Context(), "GET /subscriptions/"+id, trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	subscription, err := h.store.Get(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if subscription == nil {
		http.Error(w, "Subscription not found", http.StatusNotFound)
		return
	}

	err = json.NewEncoder(w).Encode(subscription)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *SubscriptionHandler) Update(w http.ResponseWriter, r *http.Request) {
	subscription := &model.Subscription{}
	if err := json.NewDecoder(r.Body).Decode(subscription); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	ctx, span := otel.Tracer("subscriptions").Start(r.Context(), "subscription.update", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	updatedSubscription, err := h.store.Update(ctx, subscription)
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

func (h *SubscriptionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ctx, span := otel.Tracer("subscriptions").Start(r.Context(), "subscription.delete", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	err := h.store.Delete(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
