// Copyright Dose de Telemetria GmbH
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"encoding/json"
	"net/http"

	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/model"
	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/store"
	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/telemetry"
	"go.opentelemetry.io/otel/trace"
)

// PlanHandler is an HTTP handler that performs CRUD operations for model.Plan using a store.Plan
type PlanHandler struct {
	store store.Plan
}

// NewPlanHandler returns a new PlanHandler
func NewPlanHandler(store store.Plan) *PlanHandler {
	return &PlanHandler{
		store: store,
	}
}

func (h *PlanHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx, span := telemetry.Start(r.Context(), "plans", "plan.list", trace.SpanKindServer)
	defer span.End()
	// More than 10 internal spans in one trace.
	for i := 0; i < 11; i++ {
		_, child := telemetry.Start(ctx, "plans", "plan.list.internal", trace.SpanKindInternal)
		child.End()
	}

	plans, err := h.store.List(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(plans)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *PlanHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx, span := telemetry.Start(r.Context(), "plans", "plan.create", trace.SpanKindServer)
	defer span.End()

	plan := &model.Plan{}
	if err := json.NewDecoder(r.Body).Decode(plan); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	created, err := h.store.Create(ctx, plan)
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

func (h *PlanHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	// Span name includes the id.
	ctx, span := telemetry.Start(r.Context(), "plans", "GET /plans/"+id, trace.SpanKindServer)
	defer span.End()

	plan, err := h.store.Get(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(plan)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *PlanHandler) Update(w http.ResponseWriter, r *http.Request) {
	plan := &model.Plan{}
	if err := json.NewDecoder(r.Body).Decode(plan); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	ctx, span := telemetry.Start(r.Context(), "plans", "plan.update", trace.SpanKindServer)
	defer span.End()

	updated, err := h.store.Update(ctx, plan)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(updated)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *PlanHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ctx, span := telemetry.Start(r.Context(), "plans", "plan.delete", trace.SpanKindServer)
	defer span.End()

	err := h.store.Delete(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
