package middleware

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestAPIKeyAuth_MissingKey(t *testing.T) {
	router := gin.New()
	router.Use(APIKeyAuth(nil))
	router.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// Should return 401 without API key
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestOptionalAPIKey_NoKey(t *testing.T) {
	router := gin.New()
	router.Use(OptionalAPIKey(nil))
	router.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// Should allow request through without API key
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestGetTenantID(t *testing.T) {
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		tenantID := GetTenantID(c)
		c.String(200, tenantID)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	if w.Body.String() != "test-tenant" {
		t.Errorf("expected tenant_id 'test-tenant', got: %s", w.Body.String())
	}
}

func TestGetRateLimit(t *testing.T) {
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		c.Set("rate_limit", 500)
		limit := GetRateLimit(c)
		c.String(200, strconv.Itoa(limit))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	if w.Body.String() != "500" {
		t.Errorf("expected rate_limit 500, got: %s", w.Body.String())
	}
}

func TestGetRateLimit_Default(t *testing.T) {
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		limit := GetRateLimit(c)
		c.String(200, strconv.Itoa(limit))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// Default should be 1000
	if w.Body.String() != "1000" {
		t.Errorf("expected default rate_limit 1000, got: %s", w.Body.String())
	}
}

func TestCorrelationID(t *testing.T) {
	router := gin.New()
	router.Use(CorrelationID())
	router.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// Should set X-Correlation-ID header
	correlationID := w.Header().Get("X-Correlation-ID")
	if correlationID == "" {
		t.Error("expected X-Correlation-ID header to be set")
	}
}

func TestCorrelationID_Existing(t *testing.T) {
	router := gin.New()
	router.Use(CorrelationID())
	router.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Correlation-ID", "existing-id")
	router.ServeHTTP(w, req)

	// Should preserve existing correlation ID
	correlationID := w.Header().Get("X-Correlation-ID")
	if correlationID != "existing-id" {
		t.Errorf("expected X-Correlation-ID 'existing-id', got: %s", correlationID)
	}
}

func TestGetCorrelationID(t *testing.T) {
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		c.Set("correlation_id", "test-correlation")
		id := GetCorrelationID(c)
		c.String(200, id)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	if w.Body.String() != "test-correlation" {
		t.Errorf("expected correlation_id 'test-correlation', got: %s", w.Body.String())
	}
}

func TestRequestLogger(t *testing.T) {
	router := gin.New()
	router.Use(RequestLogger())
	router.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// Should return 200
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}
