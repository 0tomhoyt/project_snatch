package core

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourname/harborlink/pkg/config"
)

func TestHealthCheck(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: 8080,
			Host: "127.0.0.1",
			Mode: "test",
		},
	}

	server := NewServer(cfg)
	router := server.Router()

	req := httptest.NewRequest(http.MethodGet, "/v2/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestRootEndpoint(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: 8080,
			Host: "127.0.0.1",
			Mode: "test",
		},
	}

	server := NewServer(cfg)
	router := server.Router()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestCreateBooking(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: 8080,
			Host: "127.0.0.1",
			Mode: "test",
		},
	}

	server := NewServer(cfg)
	router := server.Router()

	req := httptest.NewRequest(http.MethodPost, "/v2/bookings", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("expected status 202, got %d", w.Code)
	}
}

func TestGetBooking(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: 8080,
			Host: "127.0.0.1",
			Mode: "test",
		},
	}

	server := NewServer(cfg)
	router := server.Router()

	req := httptest.NewRequest(http.MethodGet, "/v2/bookings/test-ref-123", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestUpdateBooking(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: 8080,
			Host: "127.0.0.1",
			Mode: "test",
		},
	}

	server := NewServer(cfg)
	router := server.Router()

	req := httptest.NewRequest(http.MethodPut, "/v2/bookings/test-ref-123", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}
