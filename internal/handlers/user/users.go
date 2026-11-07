package user

import (
	"context"
	"errors"
	"go-pet-shop/internal/models"
	"go-pet-shop/internal/storage"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

//go:generate go run github.com/vektra/mockery/v2 --name=Users

type Users interface {
	CreateUser(ctx context.Context, user models.User) (int, error)
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
	GetAllUsers(ctx context.Context) ([]models.User, error)
}

type Handler struct {
	log     *slog.Logger
	storage Users
}

func New(log *slog.Logger, storage Users) *Handler {
	return &Handler{
		log:     log,
		storage: storage,
	}
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	const fn = "handlers.users.CreateUser"

	log := h.log.With(
		slog.String("fn", fn),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	log.Info("Creating new user", slog.String("url", r.URL.String()))

	var user models.User
	if err := render.DecodeJSON(r.Body, &user); err != nil {
		log.Error("failed to decode request body", slog.Any("error", err))
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error":   "Bad request",
			"message": "Invalid JSON payload",
		})
		return
	}

	// Валидация
	if user.Name == "" {
		log.Error("user name is empty")
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error":   "Bad request",
			"message": "User name is required",
		})
		return
	}

	if user.Email == "" {
		log.Error("user email is empty")
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error":   "Bad request",
			"message": "User email is required",
		})
		return
	}

	// Создаем пользователя
	userID, err := h.storage.CreateUser(r.Context(), user)
	if err != nil {
		log.Error("failed to create user", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"error":   "Internal server error",
			"message": "Failed to create user",
		})
		return
	}

	log.Info("User created successfully",
		slog.Int("id", userID),
		slog.String("name", user.Name),
		slog.String("url", r.URL.String()),
	)

	// Возвращаем созданный пользователя
	user.ID = userID
	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, map[string]interface{}{
		"status": "User created successfully",
		"user":   user,
	})
}

func (h *Handler) GetUserByEmail(w http.ResponseWriter, r *http.Request) {
	const fn = "handlers.users.GetUserByEmail"

	log := h.log.With(
		slog.String("fn", fn),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	log.Info("Getting user by email", slog.String("url", r.URL.String()))

	email := chi.URLParam(r, "email")
	if email == "" {
		log.Error("empty email in URL")
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error":   "Bad request",
			"message": "User email is required",
		})
		return
	}

	user, err := h.storage.GetUserByEmail(r.Context(), email)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			log.Warn("user not found", slog.String("email", email))
			w.WriteHeader(http.StatusNotFound)
			render.JSON(w, r, map[string]interface{}{
				"error":   "Not found",
				"message": "User not found",
			})
			return
		}

		log.Error("failed to get user", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"error":   "Internal server error",
			"message": "Failed to retrieve user",
		})
		return
	}

	log.Info("Retrieved user successfully",
		slog.Int("id", user.ID),
		slog.String("name", user.Name),
		slog.String("url", r.URL.String()),
	)

	render.JSON(w, r, user)
}

func (h *Handler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	const fn = "handlers.users.GetAllUsers"

	log := h.log.With(
		slog.String("fn", fn),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	items, err := h.storage.GetAllUsers(r.Context())

	if err != nil {
		log.Error("failed to get users", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"error":   "Internal server error",
			"message": "Failed to retrieve users",
		})
		return
	}

	log.Info("Retrieved users successfully",
		slog.String("url", r.URL.String()),
		slog.Int("count", len(items)),
	)

	render.JSON(w, r, items)
}
