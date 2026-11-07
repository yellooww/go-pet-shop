package user

import (
	"bytes"
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

	"go-pet-shop/internal/handlers/user/mocks"
)

// =======================
// Get All Users
// =======================

func TestGetAllUsers_Success(t *testing.T) {
	usersMock := mocks.NewUsers(t)
	usersMock.On("GetAllUsers", mock.Anything).Return([]models.User{
		{ID: 1, Name: "John Doe", Email: "john.doe@example.com"},
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	w := httptest.NewRecorder()

	handler := New(slog.Default(), usersMock)

	handler.GetAllUsers(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "John Doe") {
		t.Errorf("unexpected response body: %s", w.Body.String())
	}
}
func TestGetAllUsers_Error(t *testing.T) {
	usersMock := mocks.NewUsers(t)
	usersMock.On("GetAllUsers", mock.Anything).Return(nil, errors.New("DB error"))

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	w := httptest.NewRecorder()

	handler := New(slog.Default(), usersMock)
	handler.GetAllUsers(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "Failed to retrieve users") {
		t.Errorf("unexpected response body: %s", w.Body.String())
	}
}

// =======================
// Create User
// =======================

func TestCreateUser_Success(t *testing.T) {
	usersMock := mocks.NewUsers(t)
	usersMock.On("CreateUser", mock.Anything, mock.Anything).Return(1, nil)

	req := httptest.NewRequest(
		http.MethodPost,
		"/users",
		bytes.NewBufferString(`{"name": "John Doe", "email": "john.doe@example.com"}`),
	)
	w := httptest.NewRecorder()

	handler := New(slog.Default(), usersMock)
	handler.CreateUser(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "User created successfully") {
		t.Errorf("unexpected response body: %s", w.Body.String())
	}
}

func TestCreateUser_BadRequest(t *testing.T) {
	usersMock := mocks.NewUsers(t)

	req := httptest.NewRequest(
		http.MethodPost,
		"/users",
		bytes.NewBufferString(`{"name":}`),
	)
	w := httptest.NewRecorder()

	handler := New(slog.Default(), usersMock)
	handler.CreateUser(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "Invalid JSON payload") {
		t.Errorf("unexpected response body: %s", w.Body.String())
	}
}

func TestCreateUser_Fail(t *testing.T) {
	usersMock := mocks.NewUsers(t)
	usersMock.On("CreateUser", mock.Anything, mock.Anything).Return(0, errors.New("DB error"))

	req := httptest.NewRequest(
		http.MethodPost,
		"/users",
		bytes.NewBufferString(`{"name": "John Doe", "email": "john.doe@example.com"}`),
	)
	w := httptest.NewRecorder()

	handler := New(slog.Default(), usersMock)
	handler.CreateUser(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "Failed to create user") {
		t.Errorf("unexpected response body: %s", w.Body.String())
	}
}

// =======================
// Get User By Email
// =======================

func TestGetUserByEmail_Success(t *testing.T) {
	usersMock := mocks.NewUsers(t)
	usersMock.On("GetUserByEmail", mock.Anything, "john.doe@example.com").Return(models.User{ID: 1, Name: "John Doe", Email: "john.doe@example.com"}, nil)

	req := httptest.NewRequest(http.MethodGet, "/users/john.doe@example.com", nil)
	w := httptest.NewRecorder()

	handler := New(slog.Default(), usersMock)

	router := chi.NewRouter()
	router.Get("/users/{email}", handler.GetUserByEmail)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "John Doe") {
		t.Errorf("unexpected response body: %s", w.Body.String())
	}
}

func TestGetUserByEmail_NotFound(t *testing.T) {
	usersMock := mocks.NewUsers(t)
	usersMock.On("GetUserByEmail", mock.Anything, "john@example.com").Return(models.User{}, storage.ErrNotFound)

	req := httptest.NewRequest(http.MethodGet, "/users/john@example.com", nil)
	w := httptest.NewRecorder()

	handler := New(slog.Default(), usersMock)

	router := chi.NewRouter()
	router.Get("/users/{email}", handler.GetUserByEmail)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "User not found") {
		t.Errorf("unexpected response body: %s", w.Body.String())
	}
}

func TestGetUserByEmail_Fail(t *testing.T) {
	usersMock := mocks.NewUsers(t)
	usersMock.On("GetUserByEmail", mock.Anything, "john.doe@example.com").Return(models.User{}, errors.New("DB error"))

	req := httptest.NewRequest(http.MethodGet, "/users/john.doe@example.com", nil)
	w := httptest.NewRecorder()

	handler := New(slog.Default(), usersMock)

	router := chi.NewRouter()
	router.Get("/users/{email}", handler.GetUserByEmail)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "Failed to retrieve user") {
		t.Errorf("unexpected response body: %s", w.Body.String())
	}
}
