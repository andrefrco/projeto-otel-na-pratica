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

type inMemoryPlan struct {
	store map[string]*model.Plan
}

func NewPlanStore() store.Plan {
	return &inMemoryPlan{
		store: make(map[string]*model.Plan),
	}
}

func (u *inMemoryPlan) Get(ctx context.Context, id string) (*model.Plan, error) {
	_, span := telemetry.Start(ctx, "plans", "SELECT * FROM plans WHERE id='"+id+"'", trace.SpanKindClient)
	defer span.End()
	span.SetAttributes(attribute.String("db.query.text", "SELECT * FROM plans WHERE id='"+id+"'"))
	return u.store[id], nil
}

func (u *inMemoryPlan) Create(ctx context.Context, plan *model.Plan) (*model.Plan, error) {
	_, span := telemetry.Start(ctx, "plans", "INSERT INTO plans", trace.SpanKindClient)
	defer span.End()
	u.store[plan.ID] = plan
	return plan, nil
}

func (u *inMemoryPlan) Update(ctx context.Context, plan *model.Plan) (*model.Plan, error) {
	_, span := telemetry.Start(ctx, "plans", "UPDATE plans WHERE id='"+plan.ID+"'", trace.SpanKindClient)
	defer span.End()
	u.store[plan.ID] = plan
	return plan, nil
}

func (u *inMemoryPlan) Delete(ctx context.Context, id string) error {
	_, span := telemetry.Start(ctx, "plans", "DELETE FROM plans WHERE id='"+id+"'", trace.SpanKindClient)
	defer span.End()
	delete(u.store, id)
	return nil
}

func (u *inMemoryPlan) List(ctx context.Context) ([]*model.Plan, error) {
	_, span := telemetry.Start(ctx, "plans", "SELECT * FROM plans", trace.SpanKindClient)
	defer span.End()
	plans := make([]*model.Plan, 0, len(u.store))
	for _, plan := range u.store {
		plans = append(plans, plan)
	}
	return plans, nil
}
