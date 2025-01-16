// Copyright Dose de Telemetria GmbH
// SPDX-License-Identifier: Apache-2.0

package memory

import (
	"context"

	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/model"
	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/store"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const name = "projeto-otel-na-pratica/internal/pkg/store/memory"

var (
	tracer = otel.Tracer(name)
	logger = otelslog.NewLogger(name)
)

type inMemoryUser struct {
	store map[string]*model.User
}

func NewUserStore() store.User {
	return &inMemoryUser{
		store: make(map[string]*model.User),
	}
}

func (u *inMemoryUser) Get(ctx context.Context, id string) (*model.User, error) {
	ctx, span := tracer.Start(ctx, "Get User from Memory Store", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	span.SetAttributes(attribute.String("user.id", id))
	span.AddEvent("Starting user finding")

	user, exists := u.store[id]
	if !exists {
		span.SetStatus(codes.Unset, "User not found")
		span.AddEvent("User not found", trace.WithAttributes(attribute.String("user.id", id)))

		logger.WarnContext(ctx, "User not found", "user_id", id)
		return nil, nil
	}

	logger.InfoContext(ctx, "User found successfully", "user_id", id)
	span.SetStatus(codes.Ok, "User found successfully")
	return user, nil
}

func (u *inMemoryUser) Create(ctx context.Context, user *model.User) (*model.User, error) {
	u.store[user.ID] = user
	logger.InfoContext(ctx, "User created successfully", "user_id", user.ID)
	return user, nil
}

func (u *inMemoryUser) Update(_ context.Context, user *model.User) (*model.User, error) {
	u.store[user.ID] = user
	return user, nil
}

func (u *inMemoryUser) Delete(_ context.Context, id string) error {
	delete(u.store, id)
	return nil
}

func (u *inMemoryUser) List(ctx context.Context) ([]*model.User, error) {
	ctx, span := tracer.Start(ctx, "List Users from Memory Store", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	span.AddEvent("Starting user listing")

	users := make([]*model.User, 0, len(u.store))
	for _, user := range u.store {
		users = append(users, user)
	}

	span.SetAttributes(
		attribute.Int("users.count", len(users)),
		attribute.String("operation.name", "select"),
	)

	span.AddEvent("User listing completed")
	logger.InfoContext(ctx, "User listing successfully completed", "total_users", len(users))

	return users, nil
}
