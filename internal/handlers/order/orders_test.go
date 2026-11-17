package order

import (
	"bytes"
	"errors"
	"go-pet-shop/internal/handlers/order/mocks"
	"go-pet-shop/internal/models"
	"go-pet-shop/internal/storage"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/mock"
)

// =======================
// Create Order
// =======================

func TestCreateOrder_Success(t *testing.T) {
	ordersMock := mocks.NewOrders(t)
	ordersMock.On("CreateOrder", mock.Anything, mock.Anything).Return(1, nil)

	req := httptest.NewRequest(
		http.MethodPost,
		"/orders",
		bytes.NewBufferString(`{"user_id": 1}`),
	)
	w := httptest.NewRecorder()

	handler := New(slog.Default(), ordersMock)
	handler.CreateOrder(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "Order created successfully") {
		t.Errorf("unexpected response body: %s", w.Body.String())
	}
}

func TestCreateOrder_BadRequest(t *testing.T) {
	ordersMock := mocks.NewOrders(t)

	req := httptest.NewRequest(
		http.MethodPost,
		"/orders",
		bytes.NewBufferString(`{"user_id":}`),
	)
	w := httptest.NewRecorder()

	handler := New(slog.Default(), ordersMock)
	handler.CreateOrder(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "Invalid JSON payload") {
		t.Errorf("unexpected response body: %s", w.Body.String())
	}
}

func TestCreateOrder_InvalidUserID(t *testing.T) {
	ordersMock := mocks.NewOrders(t)

	req := httptest.NewRequest(
		http.MethodPost,
		"/orders",
		bytes.NewBufferString(`{"user_id": 0}`),
	)
	w := httptest.NewRecorder()

	handler := New(slog.Default(), ordersMock)
	handler.CreateOrder(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "Invalid user id") {
		t.Errorf("unexpected response body: %s", w.Body.String())
	}
}

func TestCreateOrder_Fail(t *testing.T) {
	ordersMock := mocks.NewOrders(t)
	ordersMock.On("CreateOrder", mock.Anything, mock.Anything).Return(0, errors.New("DB error"))

	req := httptest.NewRequest(
		http.MethodPost,
		"/orders",
		bytes.NewBufferString(`{"user_id": 1}`),
	)
	w := httptest.NewRecorder()

	handler := New(slog.Default(), ordersMock)
	handler.CreateOrder(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "Failed to create order") {
		t.Errorf("unexpected response body: %s", w.Body.String())
	}
}

// =======================
// Add Order Item
// =======================

func TestAddOrderItem_Success(t *testing.T) {
	ordersMock := mocks.NewOrders(t)
	ordersMock.On("AddOrderItem", mock.Anything, mock.Anything).Return(1, nil)

	req := httptest.NewRequest(
		http.MethodPost,
		"/orders/1/items",
		bytes.NewBufferString(`{"product_id": 1, "quantity": 2}`),
	)
	w := httptest.NewRecorder()

	handler := New(slog.Default(), ordersMock)

	router := chi.NewRouter()
	router.Post("/orders/{id}/items", handler.AddOrderItem)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "Order item added successfully") {
		t.Errorf("unexpected response body: %s", w.Body.String())
	}
}

func TestAddOrderItem_InvalidOrderID(t *testing.T) {
	ordersMock := mocks.NewOrders(t)

	req := httptest.NewRequest(
		http.MethodPost,
		"/orders/abc/items",
		bytes.NewBufferString(`{"product_id": 1, "quantity": 2}`),
	)
	w := httptest.NewRecorder()

	handler := New(slog.Default(), ordersMock)

	router := chi.NewRouter()
	router.Post("/orders/{id}/items", handler.AddOrderItem)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "Invalid order ID") {
		t.Errorf("unexpected response body: %s", w.Body.String())
	}
}

func TestAddOrderItem_BadRequest(t *testing.T) {
	ordersMock := mocks.NewOrders(t)

	req := httptest.NewRequest(
		http.MethodPost,
		"/orders/1/items",
		bytes.NewBufferString(`{"product_id":}`),
	)
	w := httptest.NewRecorder()

	handler := New(slog.Default(), ordersMock)

	router := chi.NewRouter()
	router.Post("/orders/{id}/items", handler.AddOrderItem)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "Invalid JSON payload") {
		t.Errorf("unexpected response body: %s", w.Body.String())
	}
}

func TestAddOrderItem_InvalidProductID(t *testing.T) {
	ordersMock := mocks.NewOrders(t)

	req := httptest.NewRequest(
		http.MethodPost,
		"/orders/1/items",
		bytes.NewBufferString(`{"product_id": 0, "quantity": 2}`),
	)
	w := httptest.NewRecorder()

	handler := New(slog.Default(), ordersMock)

	router := chi.NewRouter()
	router.Post("/orders/{id}/items", handler.AddOrderItem)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "Invalid product ID") {
		t.Errorf("unexpected response body: %s", w.Body.String())
	}
}

func TestAddOrderItem_InvalidQuantity(t *testing.T) {
	ordersMock := mocks.NewOrders(t)

	req := httptest.NewRequest(
		http.MethodPost,
		"/orders/1/items",
		bytes.NewBufferString(`{"product_id": 1, "quantity": 0}`),
	)
	w := httptest.NewRecorder()

	handler := New(slog.Default(), ordersMock)

	router := chi.NewRouter()
	router.Post("/orders/{id}/items", handler.AddOrderItem)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "Invalid quantity") {
		t.Errorf("unexpected response body: %s", w.Body.String())
	}
}

func TestAddOrderItem_NotFound(t *testing.T) {
	ordersMock := mocks.NewOrders(t)
	ordersMock.On("AddOrderItem", mock.Anything, mock.Anything).Return(0, storage.ErrNotFound)

	req := httptest.NewRequest(
		http.MethodPost,
		"/orders/1/items",
		bytes.NewBufferString(`{"product_id": 1, "quantity": 2}`),
	)
	w := httptest.NewRecorder()

	handler := New(slog.Default(), ordersMock)

	router := chi.NewRouter()
	router.Post("/orders/{id}/items", handler.AddOrderItem)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "Order or product not found") {
		t.Errorf("unexpected response body: %s", w.Body.String())
	}
}

func TestAddOrderItem_Fail(t *testing.T) {
	ordersMock := mocks.NewOrders(t)
	ordersMock.On("AddOrderItem", mock.Anything, mock.Anything).Return(0, errors.New("DB error"))

	req := httptest.NewRequest(
		http.MethodPost,
		"/orders/1/items",
		bytes.NewBufferString(`{"product_id": 1, "quantity": 2}`),
	)
	w := httptest.NewRecorder()

	handler := New(slog.Default(), ordersMock)

	router := chi.NewRouter()
	router.Post("/orders/{id}/items", handler.AddOrderItem)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "Failed to add order item") {
		t.Errorf("unexpected response body: %s", w.Body.String())
	}
}

// =======================
// Get Order By ID
// =======================

func TestGetOrderByID_Success(t *testing.T) {
	ordersMock := mocks.NewOrders(t)
	ordersMock.On("GetOrderByID", mock.Anything, 1).Return(models.Order{
		ID:     1,
		UserID: 1,
	}, nil)
	ordersMock.On("GetOrderItemsByOrderID", mock.Anything, 1).Return([]models.OrderItem{
		{ID: 1, OrderID: 1, ProductID: 1, Quantity: 2},
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/orders/1", nil)
	w := httptest.NewRecorder()

	handler := New(slog.Default(), ordersMock)

	router := chi.NewRouter()
	router.Get("/orders/{id}", handler.GetOrderByID)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), `"order"`) {
		t.Errorf("unexpected response body: %s", w.Body.String())
	}

	if !strings.Contains(w.Body.String(), `"items"`) {
		t.Errorf("unexpected response body: %s", w.Body.String())
	}
}

func TestGetOrderByID_InvalidID(t *testing.T) {
	ordersMock := mocks.NewOrders(t)

	req := httptest.NewRequest(http.MethodGet, "/orders/abc", nil)
	w := httptest.NewRecorder()

	handler := New(slog.Default(), ordersMock)

	router := chi.NewRouter()
	router.Get("/orders/{id}", handler.GetOrderByID)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "Invalid order ID") {
		t.Errorf("unexpected response body: %s", w.Body.String())
	}
}

func TestGetOrderByID_NotFound(t *testing.T) {
	ordersMock := mocks.NewOrders(t)
	ordersMock.On("GetOrderByID", mock.Anything, 1).Return(models.Order{}, storage.ErrNotFound)

	req := httptest.NewRequest(http.MethodGet, "/orders/1", nil)
	w := httptest.NewRecorder()

	handler := New(slog.Default(), ordersMock)

	router := chi.NewRouter()
	router.Get("/orders/{id}", handler.GetOrderByID)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "Order not found") {
		t.Errorf("unexpected response body: %s", w.Body.String())
	}
}

func TestGetOrderByID_Fail(t *testing.T) {
	ordersMock := mocks.NewOrders(t)
	ordersMock.On("GetOrderByID", mock.Anything, 1).Return(models.Order{}, errors.New("DB error"))

	req := httptest.NewRequest(http.MethodGet, "/orders/1", nil)
	w := httptest.NewRecorder()

	handler := New(slog.Default(), ordersMock)

	router := chi.NewRouter()
	router.Get("/orders/{id}", handler.GetOrderByID)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "Failed to retrieve order") {
		t.Errorf("unexpected response body: %s", w.Body.String())
	}
}

func TestGetOrderByID_ItemsFail(t *testing.T) {
	ordersMock := mocks.NewOrders(t)
	ordersMock.On("GetOrderByID", mock.Anything, 1).Return(models.Order{
		ID:     1,
		UserID: 1,
	}, nil)
	ordersMock.On("GetOrderItemsByOrderID", mock.Anything, 1).Return(nil, errors.New("DB error"))

	req := httptest.NewRequest(http.MethodGet, "/orders/1", nil)
	w := httptest.NewRecorder()

	handler := New(slog.Default(), ordersMock)

	router := chi.NewRouter()
	router.Get("/orders/{id}", handler.GetOrderByID)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "Failed to retrieve order items") {
		t.Errorf("unexpected response body: %s", w.Body.String())
	}
}

// =======================
// Get Orders By User Email
// =======================

func TestGetOrdersByUserEmail_Success(t *testing.T) {
	ordersMock := mocks.NewOrders(t)
	ordersMock.On("GetOrdersByUserEmail", mock.Anything, "john@example.com").Return([]models.Order{
		{ID: 1, UserID: 1},
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/users/john@example.com/orders", nil)
	w := httptest.NewRecorder()

	handler := New(slog.Default(), ordersMock)

	router := chi.NewRouter()
	router.Get("/users/{email}/orders", handler.GetOrdersByUserEmail)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), `"id":1`) {
		t.Errorf("unexpected response body: %s", w.Body.String())
	}
}

func TestGetOrdersByUserEmail_Fail(t *testing.T) {
	ordersMock := mocks.NewOrders(t)
	ordersMock.On("GetOrdersByUserEmail", mock.Anything, "john@example.com").Return(nil, errors.New("DB error"))

	req := httptest.NewRequest(http.MethodGet, "/users/john@example.com/orders", nil)
	w := httptest.NewRecorder()

	handler := New(slog.Default(), ordersMock)

	router := chi.NewRouter()
	router.Get("/users/{email}/orders", handler.GetOrdersByUserEmail)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "Failed to retrieve user orders") {
		t.Errorf("unexpected response body: %s", w.Body.String())
	}
}
