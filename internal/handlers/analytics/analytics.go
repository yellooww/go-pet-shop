package analytics

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

//go:generate go run github.com/vektra/mockery/v2 --name=Analytics

type Analytics interface {
	GetUserOrderHistory(ctx context.Context, email string) ([]models.OrderDetail, error)
	GetPopularProducts(ctx context.Context) ([]models.PopularProduct, error)
}

type Handler struct {
	log     *slog.Logger
	storage Analytics
}

func New(log *slog.Logger, storage Analytics) *Handler {
	return &Handler{
		log:     log,
		storage: storage,
	}
}

func (h *Handler) GetUserOrderHistory(w http.ResponseWriter, r *http.Request) {
	const fn = "handlers.analytics.GetUserOrderHistory"

	log := h.log.With(
		slog.String("fn", fn),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	email := chi.URLParam(r, "email")

	log.Info("Getting user order history", slog.String("email", email))

	if email == "" {
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error":   "Bad request",
			"message": "User email is required",
		})
		return
	}

	history, err := h.storage.GetUserOrderHistory(r.Context(), email)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			render.JSON(w, r, map[string]string{
				"error":   "Not found",
				"message": "User not found",
			})
			return
		}

		log.Error("failed to get user order history", slog.Any("error", err))

		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"error":   "Internal server error",
			"message": "Failed to get order history",
		})
		return
	}

	if history == nil {
		history = make([]models.OrderDetail, 0)
	}

	render.JSON(w, r, map[string]interface{}{
		"history": history,
	})
}

func (h *Handler) GetPopularProducts(w http.ResponseWriter, r *http.Request) {
	const fn = "handlers.analytics.GetPopularProducts"

	log := h.log.With(
		slog.String("fn", fn),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	log.Info("Getting popular products")

	products, err := h.storage.GetPopularProducts(r.Context())
	if err != nil {
		log.Error("failed to get popular products", slog.Any("error", err))

		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"error":   "Internal server error",
			"message": "Failed to get popular products",
		})
		return
	}

	if products == nil {
		products = make([]models.PopularProduct, 0)
	}

	render.JSON(w, r, map[string]interface{}{
		"products": products,
	})
}
