package http

import (
	"encoding/json"
	"net/http"

	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/model"
	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/store"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const name = "projeto-otel-na-pratica/internal/pkg/handler/http"

var (
	tracer = otel.Tracer(name)
	meter  = otel.Meter(name)
	logger = otelslog.NewLogger(name)
	reqCnt metric.Int64Counter
	reqDur metric.Float64Histogram
)

// UserHandler is an HTTP handler that performs CRUD operations for model.User using a store.User
type UserHandler struct {
	store store.User
}

// NewUserHandler returns a new UserHandler
func NewUserHandler(store store.User) *UserHandler {
	return &UserHandler{
		store: store,
	}
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	users, err := h.store.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.ErrorContext(r.Context(), "Erro ao listar usuários", "error", err)
		return
	}

	err = json.NewEncoder(w).Encode(users)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.ErrorContext(r.Context(), "Erro ao codificar resposta JSON", "error", err)
		return
	}
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "Decode Request Body for POST /users")
	defer span.End()

	user := &model.User{}
	if err := json.NewDecoder(r.Body).Decode(user); err != nil {
		span.RecordError(err)
		logger.ErrorContext(r.Context(), "Erro ao decodificar request body", "error", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	created, err := h.store.Create(ctx, user)
	if err != nil {
		span.RecordError(err)
		logger.ErrorContext(r.Context(), "Erro ao criar usuário no store", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	span.SetAttributes(attribute.String("user.id", created.ID))

	err = json.NewEncoder(w).Encode(created)
	if err != nil {
		span.RecordError(err)
		logger.ErrorContext(r.Context(), "Erro ao codificar resposta JSON", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ctx := r.Context()

	logger.DebugContext(ctx, "Iniciando requisição para obter usuário", "user_id", id)

	user, err := h.store.Get(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.ErrorContext(ctx, "Erro ao obter usuário", "error", err, "user_id", id)
		return
	}

	if user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.ErrorContext(ctx, "Erro ao codificar resposta JSON", "error", err, "user_id", id)
		return
	}
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	user := &model.User{}
	if err := json.NewDecoder(r.Body).Decode(user); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		logger.ErrorContext(r.Context(), "Erro ao decodificar request body", "error", err)
		return
	}

	updated, err := h.store.Update(r.Context(), user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.ErrorContext(r.Context(), "Erro ao atualizar usuário", "error", err)
		return
	}

	logger.InfoContext(r.Context(), "Usuário atualizado com sucesso", "user_id", updated.ID)
	err = json.NewEncoder(w).Encode(updated)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.ErrorContext(r.Context(), "Erro ao codificar resposta JSON", "error", err)
		return
	}
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	err := h.store.Delete(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.ErrorContext(r.Context(), "Erro ao deletar usuário", "error", err, "user_id", id)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	logger.InfoContext(r.Context(), "Usuário deletado com sucesso", "user_id", id)
}
