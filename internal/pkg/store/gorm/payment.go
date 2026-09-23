// Copyright Dose de Telemetria GmbH
// SPDX-License-Identifier: Apache-2.0

package gorm

import (
	"context"
	"strconv"

	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/model"
	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/store"
	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

type Payment struct {
	db *gorm.DB
}

func NewPaymentStore(db *gorm.DB) store.Payment {
	return &Payment{db: db}
}

func (p *Payment) Get(ctx context.Context, id string) (*model.Payment, error) {
	_, span := telemetry.Start(ctx, "payments", "SELECT * FROM payments WHERE id='"+id+"'", trace.SpanKindClient)
	defer span.End()
	span.SetAttributes(attribute.String("db.query.text", "SELECT * FROM payments WHERE id='"+id+"'"))
	ret := &model.Payment{}
	_ = p.db.WithContext(ctx).Model(ret).First(&ret, "id = ?", id)
	return ret, nil
}

func (p *Payment) Create(ctx context.Context, payment *model.Payment) (*model.Payment, error) {
	query := "INSERT INTO payments (id, subscription_id, amount) VALUES ('" + payment.ID + "', '" + payment.SubscriptionID + "', " + strconv.FormatFloat(payment.Amount, 'f', -1, 64) + ")"
	_, span := telemetry.Start(ctx, "payments", query, trace.SpanKindClient)
	defer span.End()
	span.SetAttributes(attribute.String("db.query.text", query))
	res := p.db.WithContext(ctx).Create(&payment)
	return payment, res.Error
}

func (p *Payment) Update(ctx context.Context, payment *model.Payment) (*model.Payment, error) {
	_, span := telemetry.Start(ctx, "payments", "UPDATE payments WHERE id='"+payment.ID+"'", trace.SpanKindClient)
	defer span.End()
	res := p.db.WithContext(ctx).Save(&payment)
	return payment, res.Error
}

func (p *Payment) Delete(ctx context.Context, id string) error {
	_, span := telemetry.Start(ctx, "payments", "DELETE FROM payments WHERE id='"+id+"'", trace.SpanKindClient)
	defer span.End()
	_ = p.db.WithContext(ctx).Delete(&model.Payment{}, "id = ?", id)
	return nil
}

func (p *Payment) List(ctx context.Context) ([]*model.Payment, error) {
	_, span := telemetry.Start(ctx, "payments", "SELECT * FROM payments", trace.SpanKindClient)
	defer span.End()
	var ret []*model.Payment
	_ = p.db.WithContext(ctx).Find(&ret)
	return ret, nil
}
