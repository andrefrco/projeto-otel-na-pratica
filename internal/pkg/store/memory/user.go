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

type inMemoryUser struct {
	store map[string]*model.User
}

func NewUserStore() store.User {
	return &inMemoryUser{
		store: make(map[string]*model.User),
	}
}

func (u *inMemoryUser) Get(ctx context.Context, id string) (*model.User, error) {
	_, span := telemetry.Start(ctx, "users", "SELECT * FROM users WHERE id='"+id+"'", trace.SpanKindClient)
	defer span.End()
	span.SetAttributes(attribute.String("db.query.text", "SELECT * FROM users WHERE id='"+id+"'"))
	return u.store[id], nil
}

func (u *inMemoryUser) Create(ctx context.Context, user *model.User) (*model.User, error) {
	_, span := telemetry.Start(ctx, "users", "INSERT INTO users", trace.SpanKindClient)
	defer span.End()
	span.SetAttributes(attribute.String("db.query.text", "INSERT INTO users (email, address) VALUES ('"+user.Email+"', '"+user.Address+"')"))
	u.store[user.ID] = user
	return user, nil
}

func (u *inMemoryUser) Update(ctx context.Context, user *model.User) (*model.User, error) {
	_, span := telemetry.Start(ctx, "users", "UPDATE users SET email='"+user.Email+"' WHERE id='"+user.ID+"'", trace.SpanKindClient)
	defer span.End()
	span.SetAttributes(attribute.String("db.query.text", "UPDATE users SET email='"+user.Email+"', address='"+user.Address+"' WHERE id='"+user.ID+"'"))
	u.store[user.ID] = user
	return user, nil
}

func (u *inMemoryUser) Delete(ctx context.Context, id string) error {
	_, span := telemetry.Start(ctx, "users", "DELETE FROM users WHERE id='"+id+"'", trace.SpanKindClient)
	defer span.End()
	delete(u.store, id)
	return nil
}

func (u *inMemoryUser) List(ctx context.Context) ([]*model.User, error) {
	_, span := telemetry.Start(ctx, "users", "SELECT * FROM users", trace.SpanKindClient)
	defer span.End()
	users := make([]*model.User, 0, len(u.store))
	for _, user := range u.store {
		users = append(users, user)
	}
	return users, nil
}
