package analytics

import (
	"errors"
	"go-pet-shop/internal/models"
	"go-pet-shop/internal/storage"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/mock"

	"go-pet-shop/internal/handlers/analytics/mocks"
)

// =======================
// Get User Order History
// =======================

func TestGetUserOrderHistory_Success(t *testing.T) {
	analyticsMock := mocks.NewAnalytics(t)

	analyticsMock.On(
		"GetUserOrderHistory",
		mock.Anything,
		"john.doe@example.com",
	).Return([]models.OrderDetail{}, nil)

	req := httptest.NewRequest(
		http.MethodGet,
		"/analytics/orders/john.doe@example.com",
		nil,
	)
	w := httptest.NewRecorder()

	handler := New(slog.Default(), analyticsMock)

	router := chi.NewRouter()
	router.Get("/analytics/orders/{email}", handler.GetUserOrderHistory)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), `"history"`) {
		t.Errorf("unexpected response body: %s", w.Body.String())
	}
}

func TestGetUserOrderHistory_NotFound(t *testing.T) {
	analyticsMock := mocks.NewAnalytics(t)

	analyticsMock.On(
		"GetUserOrderHistory",
		mock.Anything,
		"john.doe@example.com",
	).Return(nil, storage.ErrNotFound)

	req := httptest.NewRequest(
		http.MethodGet,
		"/analytics/orders/john.doe@example.com",
		nil,
	)
	w := httptest.NewRecorder()

	handler := New(slog.Default(), analyticsMock)

	router := chi.NewRouter()
	router.Get("/analytics/orders/{email}", handler.GetUserOrderHistory)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "User not found") {
		t.Errorf("unexpected response body: %s", w.Body.String())
	}
}

func TestGetUserOrderHistory_Fail(t *testing.T) {
	analyticsMock := mocks.NewAnalytics(t)

	analyticsMock.On(
		"GetUserOrderHistory",
		mock.Anything,
		"john.doe@example.com",
	).Return(nil, errors.New("DB error"))

	req := httptest.NewRequest(
		http.MethodGet,
		"/analytics/orders/john.doe@example.com",
		nil,
	)
	w := httptest.NewRecorder()

	handler := New(slog.Default(), analyticsMock)

	router := chi.NewRouter()
	router.Get("/analytics/orders/{email}", handler.GetUserOrderHistory)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "Failed to get order history") {
		t.Errorf("unexpected response body: %s", w.Body.String())
	}
}

func TestGetUserOrderHistory_EmptyHistory(t *testing.T) {
	analyticsMock := mocks.NewAnalytics(t)

	analyticsMock.On(
		"GetUserOrderHistory",
		mock.Anything,
		"john.doe@example.com",
	).Return(nil, nil)

	req := httptest.NewRequest(
		http.MethodGet,
		"/analytics/orders/john.doe@example.com",
		nil,
	)
	w := httptest.NewRecorder()

	handler := New(slog.Default(), analyticsMock)

	router := chi.NewRouter()
	router.Get("/analytics/orders/{email}", handler.GetUserOrderHistory)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), `"history":[]`) {
		t.Errorf("expected empty history, got: %s", w.Body.String())
	}
}

// =======================
// Get Popular Products
// =======================

func TestGetPopularProducts_Success(t *testing.T) {
	analyticsMock := mocks.NewAnalytics(t)

	analyticsMock.On(
		"GetPopularProducts",
		mock.Anything,
	).Return([]models.PopularProduct{}, nil)

	req := httptest.NewRequest(
		http.MethodGet,
		"/analytics/popular-products",
		nil,
	)
	w := httptest.NewRecorder()

	handler := New(slog.Default(), analyticsMock)
	handler.GetPopularProducts(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), `"products"`) {
		t.Errorf("unexpected response body: %s", w.Body.String())
	}
}

func TestGetPopularProducts_EmptyProducts(t *testing.T) {
	analyticsMock := mocks.NewAnalytics(t)

	analyticsMock.On(
		"GetPopularProducts",
		mock.Anything,
	).Return(nil, nil)

	req := httptest.NewRequest(
		http.MethodGet,
		"/analytics/popular-products",
		nil,
	)
	w := httptest.NewRecorder()

	handler := New(slog.Default(), analyticsMock)
	handler.GetPopularProducts(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), `"products":[]`) {
		t.Errorf("expected empty products, got: %s", w.Body.String())
	}
}

func TestGetPopularProducts_Fail(t *testing.T) {
	analyticsMock := mocks.NewAnalytics(t)

	analyticsMock.On(
		"GetPopularProducts",
		mock.Anything,
	).Return(nil, errors.New("DB error"))

	req := httptest.NewRequest(
		http.MethodGet,
		"/analytics/popular-products",
		nil,
	)
	w := httptest.NewRecorder()

	handler := New(slog.Default(), analyticsMock)
	handler.GetPopularProducts(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "Failed to get popular products") {
		t.Errorf("unexpected response body: %s", w.Body.String())
	}
}
