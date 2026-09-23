// Copyright Dose de Telemetria GmbH
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"context"
	"time"

	"github.com/dosedetelemetria/projeto-otel-na-pratica/api"
	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/model"
	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/store"
	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/telemetry"
	"go.opentelemetry.io/otel/trace"
)

type planServer struct {
	api.UnimplementedPlanServiceServer

	store store.Plan
}

func NewPlanServer(store store.Plan) api.PlanServiceServer {
	return &planServer{
		store: store,
	}
}

func (s *planServer) Get(ctx context.Context, req *api.GetRequest) (*api.GetResponse, error) {
	ctx, span := telemetry.Start(ctx, "plans", "GET /plans/"+req.Id, trace.SpanKindServer)
	defer span.End()

	plan, err := s.store.Get(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	resp := &api.GetResponse{
		Plan: &api.Plan{
			Id:          plan.ID,
			Name:        plan.Name,
			Description: plan.Description,
			Price:       plan.Price,
			Version:     plan.Version,
			CreatedAt:   plan.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   plan.UpdatedAt.Format(time.RFC3339),
			DeletedAt:   plan.DeletedAt.Format(time.RFC3339),
		},
	}
	return resp, nil
}

func (s *planServer) Create(ctx context.Context, req *api.CreateRequest) (*api.CreateResponse, error) {
	ctx, span := telemetry.Start(ctx, "plans", "plan.create", trace.SpanKindServer)
	defer span.End()

	plan, err := s.store.Create(ctx, &model.Plan{
		ID:          req.Plan.Id,
		Name:        req.Plan.Name,
		Description: req.Plan.Description,
		Price:       req.Plan.Price,
		Version:     req.Plan.Version,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	})
	if err != nil {
		return nil, err
	}

	resp := &api.CreateResponse{
		Plan: &api.Plan{
			Id:          plan.ID,
			Name:        plan.Name,
			Description: plan.Description,
			Price:       plan.Price,
			Version:     plan.Version,
			CreatedAt:   plan.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   plan.UpdatedAt.Format(time.RFC3339),
		},
	}
	return resp, nil
}

func (s *planServer) Update(ctx context.Context, req *api.UpdateRequest) (*api.UpdateResponse, error) {
	ctx, span := telemetry.Start(ctx, "plans", "plan.update", trace.SpanKindServer)
	defer span.End()

	plan, err := s.store.Update(ctx, &model.Plan{
		ID:          req.Plan.Id,
		Name:        req.Plan.Name,
		Description: req.Plan.Description,
		Price:       req.Plan.Price,
		Version:     req.Plan.Version,
		UpdatedAt:   time.Now(),
	})
	if err != nil {
		return nil, err
	}

	resp := &api.UpdateResponse{
		Plan: &api.Plan{
			Id:          plan.ID,
			Name:        plan.Name,
			Description: plan.Description,
			Price:       plan.Price,
			Version:     plan.Version,
			CreatedAt:   plan.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   plan.UpdatedAt.Format(time.RFC3339),
		},
	}
	return resp, nil
}

func (s *planServer) Delete(ctx context.Context, req *api.DeleteRequest) (*api.DeleteResponse, error) {
	ctx, span := telemetry.Start(ctx, "plans", "plan.delete", trace.SpanKindServer)
	defer span.End()

	err := s.store.Delete(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &api.DeleteResponse{}, nil
}

func (s *planServer) List(ctx context.Context, req *api.ListRequest) (*api.ListResponse, error) {
	ctx, span := telemetry.Start(ctx, "plans", "plan.list", trace.SpanKindServer)
	defer span.End()

	plans, err := s.store.List(ctx)
	if err != nil {
		return nil, err
	}

	resp := &api.ListResponse{
		Plans: make([]*api.Plan, len(plans)),
	}

	for i, plan := range plans {
		resp.Plans[i] = &api.Plan{
			Id:          plan.ID,
			Name:        plan.Name,
			Description: plan.Description,
			Price:       plan.Price,
			Version:     plan.Version,
			CreatedAt:   plan.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   plan.UpdatedAt.Format(time.RFC3339),
		}
	}
	return resp, nil
}
