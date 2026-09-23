// Copyright Dose de Telemetria GmbH
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/model"
	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/store"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// PaymentHandler is an HTTP handler that performs CRUD operations for model.Payment using a store.Payment
type PaymentHandler struct {
	store                 store.Payment
	js                    jetstream.JetStream
	jsSubject             string
	subscriptionsEndpoint string
}

// NewPaymentHandler returns a new PaymentHandler
func NewPaymentHandler(store store.Payment, js jetstream.JetStream, jsSubject string, subscriptionsEndpoint string) *PaymentHandler {
	return &PaymentHandler{
		store:                 store,
		js:                    js,
		jsSubject:             jsSubject,
		subscriptionsEndpoint: subscriptionsEndpoint,
	}
}

func (h *PaymentHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("payments").Start(r.Context(), "payment.list", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	payments, err := h.store.List(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(payments)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *PaymentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var payment model.Payment
	if err := json.NewDecoder(r.Body).Decode(&payment); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	ctx, span := otel.Tracer("payments").Start(r.Context(), "payment.create", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// Check if subscription exists.
	// Client span is a new root, and the name includes the subscription id.
	_, clientSpan := otel.Tracer("payments").Start(context.Background(), "GET "+h.subscriptionsEndpoint+"/"+payment.SubscriptionID, trace.WithSpanKind(trace.SpanKindClient))
	sub, _ := http.Get(h.subscriptionsEndpoint + "/" + payment.SubscriptionID)
	clientSpan.End()
	if sub.StatusCode != http.StatusOK {
		http.Error(w, "Subscription not found", http.StatusBadRequest)
		return
	}
	defer sub.Body.Close()

	payload, err := json.Marshal(payment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = h.js.PublishMsgAsync(&nats.Msg{
		Subject: h.jsSubject,
		Data:    payload,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Counter has no unit, and subscription.id is unbounded.
	counter, counterErr := otel.Meter("payments").Int64Counter("payments.created")
	if counterErr == nil {
		counter.Add(ctx, 1, metric.WithAttributes(attribute.String("subscription.id", payment.SubscriptionID)))
	}

	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(payment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *PaymentHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ctx, span := otel.Tracer("payments").Start(r.Context(), "GET /payments/"+id, trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	payment, err := h.store.Get(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if payment == nil {
		http.Error(w, "Payment not found", http.StatusNotFound)
		return
	}

	err = json.NewEncoder(w).Encode(payment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *PaymentHandler) Update(w http.ResponseWriter, r *http.Request) {
	payment := &model.Payment{}
	if err := json.NewDecoder(r.Body).Decode(payment); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	ctx, span := otel.Tracer("payments").Start(r.Context(), "payment.update", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	_, err := h.store.Update(ctx, payment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(payment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *PaymentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ctx, span := otel.Tracer("payments").Start(r.Context(), "payment.delete", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	err := h.store.Delete(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *PaymentHandler) OnMessage(msg jetstream.Msg) {
	ctx, span := otel.Tracer("payments").Start(context.Background(), "payment.consume", trace.WithSpanKind(trace.SpanKindConsumer))
	defer span.End()

	payment := &model.Payment{}
	err := json.Unmarshal(msg.Data(), payment)
	if err != nil {
		return
	}

	_, err = h.store.Create(ctx, payment)
	if err != nil {
		return
	}

	_ = msg.Ack()
}
