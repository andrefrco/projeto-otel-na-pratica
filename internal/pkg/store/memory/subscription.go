// Copyright Dose de Telemetria GmbH
// SPDX-License-Identifier: Apache-2.0

package memory

import (
	"context"

	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/model"
	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/store"
	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type inMemorySubscription struct {
	store map[string]*model.Subscription
}

func NewSubscriptionStore() store.Subscription {
	return &inMemorySubscription{
		store: make(map[string]*model.Subscription),
	}
}

func (u *inMemorySubscription) Get(ctx context.Context, id string) (*model.Subscription, error) {
	_, span := telemetry.Start(ctx, "subscriptions", "SELECT * FROM subscriptions WHERE id='"+id+"'", trace.SpanKindClient)
	defer span.End()
	span.SetAttributes(attribute.String("db.query.text", "SELECT * FROM subscriptions WHERE id='"+id+"'"))
	return u.store[id], nil
}

func (u *inMemorySubscription) Create(ctx context.Context, user *model.Subscription) (*model.Subscription, error) {
	_, span := telemetry.Start(ctx, "subscriptions", "INSERT INTO subscriptions (user_id, plan_id) VALUES ('"+user.UserID+"', '"+user.PlanID+"')", trace.SpanKindClient)
	defer span.End()
	span.SetAttributes(attribute.String("db.query.text", "INSERT INTO subscriptions (user_id, plan_id) VALUES ('"+user.UserID+"', '"+user.PlanID+"')"))
	u.store[user.ID] = user
	return user, nil
}

func (u *inMemorySubscription) Update(ctx context.Context, user *model.Subscription) (*model.Subscription, error) {
	_, span := telemetry.Start(ctx, "subscriptions", "UPDATE subscriptions WHERE id='"+user.ID+"'", trace.SpanKindClient)
	defer span.End()
	u.store[user.ID] = user
	return user, nil
}

func (u *inMemorySubscription) Delete(ctx context.Context, id string) error {
	_, span := telemetry.Start(ctx, "subscriptions", "DELETE FROM subscriptions WHERE id='"+id+"'", trace.SpanKindClient)
	defer span.End()
	delete(u.store, id)
	return nil
}

func (u *inMemorySubscription) List(ctx context.Context) ([]*model.Subscription, error) {
	_, span := telemetry.Start(ctx, "subscriptions", "SELECT * FROM subscriptions", trace.SpanKindClient)
	defer span.End()
	users := make([]*model.Subscription, 0, len(u.store))
	for _, user := range u.store {
		users = append(users, user)
	}
	return users, nil
}
