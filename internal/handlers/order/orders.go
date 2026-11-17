package order

import (
	"context"
	"errors"
	"go-pet-shop/internal/models"
	"go-pet-shop/internal/storage"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

//go:generate go run github.com/vektra/mockery/v2 --name=Orders

type Orders interface {
	CreateOrder(ctx context.Context, order models.Order) (int, error)
	AddOrderItem(ctx context.Context, orderItem models.OrderItem) (int, error)
	GetOrderByID(ctx context.Context, id int) (models.Order, error)
	GetOrdersByUserEmail(ctx context.Context, email string) ([]models.Order, error)
	GetOrderItemsByOrderID(ctx context.Context, orderID int) ([]models.OrderItem, error)
}

type Handler struct {
	log     *slog.Logger
	storage Orders
}

func New(log *slog.Logger, storage Orders) *Handler {
	return &Handler{
		log:     log,
		storage: storage,
	}
}

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	const fn = "handlers.orders.CreateOrder"

	log := h.log.With(
		slog.String("fn", fn),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	log.Info("Creating new order", slog.String("url", r.URL.String()))

	var order models.Order

	if err := render.DecodeJSON(r.Body, &order); err != nil {
		log.Error("failed to decode request body", slog.Any("error", err))

		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error":   "Bad request",
			"message": "Invalid JSON payload",
		})
		return
	}

	// Валидация
	if order.UserID <= 0 {
		log.Error("invalid user_id", slog.Int("user_id", order.UserID))
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error":   "Bad request",
			"message": "Invalid user id",
		})
		return
	}

	// Создаем заказ
	orderID, err := h.storage.CreateOrder(r.Context(), order)
	if err != nil {
		log.Error("failed to create order", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"error":   "Internal server error",
			"message": "Failed to create order",
		})
		return
	}

	log.Info("Order created successfully",
		slog.Int("id", orderID),
		slog.Int("user_id", order.UserID),
		slog.String("url", r.URL.String()),
	)

	// Возвращаем созданный пользователя
	order.ID = orderID
	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, map[string]interface{}{
		"status": "Order created successfully",
		"order":  order,
	})
}

func (h *Handler) AddOrderItem(w http.ResponseWriter, r *http.Request) {
	const fn = "handlers.orders.AddOrderItem"

	log := h.log.With(
		slog.String("fn", fn),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	log.Info("Adding item to order", slog.String("url", r.URL.String()))

	orderID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || orderID <= 0 {
		log.Error("invalid orderID in URL", slog.String("id", chi.URLParam(r, "id")))
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error":   "Bad request",
			"message": "Invalid order ID",
		})
		return
	}

	var item models.OrderItem

	if err := render.DecodeJSON(r.Body, &item); err != nil {
		log.Error("failed to decode request body", slog.Any("error", err))
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error":   "Bad request",
			"message": "Invalid JSON payload",
		})
		return
	}

	// Валидация
	if item.ProductID <= 0 {
		log.Error("invalid product_id", slog.Int("product_id", item.ProductID))
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error":   "Bad request",
			"message": "Invalid product ID",
		})
		return
	}

	if item.Quantity <= 0 {
		log.Error("invalid quantity", slog.Int("quantity", item.Quantity))
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error":   "Bad request",
			"message": "Invalid quantity",
		})
		return
	}

	item.OrderID = orderID

	itemID, err := h.storage.AddOrderItem(r.Context(), item)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			log.Error("order or product not found", slog.Int("order_id", orderID), slog.Int("product_id", item.ProductID))
			w.WriteHeader(http.StatusNotFound)
			render.JSON(w, r, map[string]string{
				"error":   "Not found",
				"message": "Order or product not found",
			})
			return
		}

		log.Error("failed to add order item", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"error":   "Internal server error",
			"message": "Failed to add order item",
		})
		return
	}

	log.Info("Order item added successfully",
		slog.Int("order_id", orderID),
		slog.Int("product_id", item.ProductID),
		slog.Int("quantity", item.Quantity),
		slog.String("url", r.URL.String()),
	)

	item.ID = itemID
	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, map[string]interface{}{
		"status": "Order item added successfully",
		"item":   item,
	})
}

func (h *Handler) GetOrderByID(w http.ResponseWriter, r *http.Request) {
	const fn = "handlers.orders.GetOrderByID"

	log := h.log.With(
		slog.String("fn", fn),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	log.Info("Getting order by ID", slog.String("url", r.URL.String()))

	// Получаем ID из URL
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		log.Error("invalid order ID")
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error":   "Bad request",
			"message": "Invalid order ID",
		})
		return
	}

	// Получаем сам заказ
	order, err := h.storage.GetOrderByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			log.Warn("order not found", slog.Int("id", id))
			w.WriteHeader(http.StatusNotFound)
			render.JSON(w, r, map[string]string{
				"error":   "Not found",
				"message": "Order not found",
			})
			return
		}

		log.Error("failed to get order", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"error":   "Internal server error",
			"message": "Failed to retrieve order",
		})
		return
	}

	// Получаем товары этого заказа
	items, err := h.storage.GetOrderItemsByOrderID(r.Context(), id)
	if err != nil {
		log.Error("failed to get order items", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"error":   "Internal server error",
			"message": "Failed to retrieve order items",
		})
		return
	}

	if items == nil {
		items = []models.OrderItem{}
	}

	log.Info("Retrieved order successfully",
		slog.Int("id", id),
		slog.Int("items_count", len(items)),
	)

	render.JSON(w, r, map[string]interface{}{
		"order": order,
		"items": items,
	})
}

func (h *Handler) GetOrdersByUserEmail(w http.ResponseWriter, r *http.Request) {
	const fn = "handlers.orders.GetOrdersByUserEmail"

	log := h.log.With(
		slog.String("fn", fn),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	log.Info("Getting user orders", slog.String("url", r.URL.String()))

	// Получаем email из URL
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

	// Получаем заказы пользователя
	orders, err := h.storage.GetOrdersByUserEmail(r.Context(), email)
	if err != nil {
		log.Error("failed to get user orders", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"error":   "Internal server error",
			"message": "Failed to retrieve user orders",
		})
		return
	}

	if orders == nil {
		orders = []models.Order{}
	}

	log.Info("Retrieved user orders successfully",
		slog.String("email", email),
		slog.Int("count", len(orders)),
	)

	render.JSON(w, r, orders)
}
