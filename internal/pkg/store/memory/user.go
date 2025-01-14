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

	user, exists := u.store[id]
	if !exists {
		span.SetStatus(codes.Unset, "User not found")
		span.AddEvent("User not found", trace.WithAttributes(attribute.String("user.id", id)))

		logger.WarnContext(ctx, "Usuário não encontrado", "user_id", id)
		return nil, nil
	}

	logger.InfoContext(ctx, "Usuário encontrado com sucesso", "user_id", id)
	return user, nil
}

func (u *inMemoryUser) Create(ctx context.Context, user *model.User) (*model.User, error) {
	ctx, span := tracer.Start(ctx, "Store Create Operation", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()
	u.store[user.ID] = user

	span.SetAttributes(
		attribute.String("user.id", user.ID),
		attribute.String("operation.name", "Insert"),
	)
	logger.InfoContext(ctx, "Usuário criado com sucesso", "user_id", user.ID)
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

	span.AddEvent("Iniciando listagem de usuários")

	users := make([]*model.User, 0, len(u.store))
	for _, user := range u.store {
		users = append(users, user)
	}

	span.SetAttributes(
		attribute.Int("users.count", len(users)),
	)

	span.AddEvent("Listagem de usuários concluída")
	logger.InfoContext(ctx, "Listagem de usuários realizada com sucesso", "total_users", len(users))

	return users, nil
}
