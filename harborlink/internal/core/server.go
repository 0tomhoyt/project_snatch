package core

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/yourname/harborlink/internal/adapter"
	"github.com/yourname/harborlink/internal/booking"
	"github.com/yourname/harborlink/internal/eventbus"
	"github.com/yourname/harborlink/internal/handler"
	"github.com/yourname/harborlink/internal/middleware"
	"github.com/yourname/harborlink/internal/notification"
	"github.com/yourname/harborlink/internal/repository"
	"github.com/yourname/harborlink/internal/scheduler"
	"github.com/yourname/harborlink/internal/service"
	"github.com/yourname/harborlink/pkg/cache"
	"github.com/yourname/harborlink/pkg/config"
)

// Server represents the HTTP server
type Server struct {
	cfg               *config.Config
	router            *gin.Engine
	server            *http.Server
	bookingHandler    *handler.BookingHandler
	slotWatchHandler  *handler.SlotWatchHandler
	scheduler         *scheduler.Scheduler
	quickLockHandler  *booking.QuickLockHandler
	wsHub             *notification.WebSocketHub
}

// ServerDependencies holds all dependencies needed by the server
type ServerDependencies struct {
	BookingService   *service.BookingService
	BookingCache     *cache.BookingCache
	APIKeyRepo       repository.APIKeyRepository
	Registry         *adapter.Registry
	EventBus         *eventbus.EventBus
	WsHub            *notification.WebSocketHub
	Scheduler        *scheduler.Scheduler
	SlotWatchRepo    repository.SlotWatchRepository
	QuickLockHandler *booking.QuickLockHandler
}

// NewServer creates a new HTTP server
func NewServer(cfg *config.Config) *Server {
	// Set Gin mode
	switch cfg.Server.Mode {
	case "release":
		gin.SetMode(gin.ReleaseMode)
	case "test":
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.DebugMode)
	}

	router := gin.New()

	s := &Server{
		cfg:    cfg,
		router: router,
	}

	// Setup basic middlewares
	router.Use(gin.Recovery())
	router.Use(middleware.RequestLogger())
	router.Use(middleware.CorrelationID())
	router.Use(corsMiddleware())
	router.Use(apiVersionMiddleware())

	// Setup routes
	s.setupRoutes()

	return s
}

// NewServerWithDeps creates a new HTTP server with full dependencies
func NewServerWithDeps(cfg *config.Config, deps *ServerDependencies) *Server {
	// Set Gin mode
	switch cfg.Server.Mode {
	case "release":
		gin.SetMode(gin.ReleaseMode)
	case "test":
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.DebugMode)
	}

	router := gin.New()

	s := &Server{
		cfg:              cfg,
		router:           router,
		bookingHandler:   handler.NewBookingHandler(deps.BookingService),
		scheduler:        deps.Scheduler,
		quickLockHandler: deps.QuickLockHandler,
		wsHub:            deps.WsHub,
	}

	// Setup slot watch handler if dependencies are available
	if deps.SlotWatchRepo != nil && deps.Scheduler != nil {
		slotWatchSvc := service.NewSlotWatchService(deps.SlotWatchRepo, deps.Scheduler, &cfg.SlotWatch)
		s.slotWatchHandler = handler.NewSlotWatchHandler(slotWatchSvc, deps.WsHub)
	}

	// Setup middlewares
	router.Use(gin.Recovery())
	router.Use(middleware.RecoveryLogger())
	router.Use(middleware.RequestLogger())
	router.Use(middleware.CorrelationID())
	router.Use(corsMiddleware())
	router.Use(apiVersionMiddleware())

	// Add rate limiting if cache is available
	if deps.BookingCache != nil {
		router.Use(middleware.RateLimiter(deps.BookingCache, middleware.DefaultRateLimiterConfig()))
	}

	// Setup routes
	s.setupRoutesWithAuth(deps)

	return s
}

// setupRoutes configures all routes (basic version without auth)
func (s *Server) setupRoutes() {
	// API v2 routes
	v2 := s.router.Group("/v2")
	{
		// Health check (public)
		v2.GET("/health", s.healthCheck)

		// Booking endpoints (placeholder handlers)
		bookings := v2.Group("/bookings")
		{
			bookings.GET("", s.listBookings)              // List bookings
			bookings.POST("", s.createBooking)            // Create booking
			bookings.GET("/:reference", s.getBooking)     // Get booking
			bookings.PUT("/:reference", s.updateBooking)  // Update booking
			bookings.DELETE("/:reference", s.cancelBooking) // Cancel booking
		}
	}

	// Root endpoint
	s.router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"name":        "HarborLink",
			"version":     "2.0.0",
			"description": "DCSA API Aggregator Gateway",
		})
	})
}

// setupRoutesWithAuth configures all routes with authentication
func (s *Server) setupRoutesWithAuth(deps *ServerDependencies) {
	// Health check (public, no auth required)
	s.router.GET("/v2/health", s.healthCheck)

	// Root endpoint (public)
	s.router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"name":        "HarborLink",
			"version":     "2.0.0",
			"description": "DCSA API Aggregator Gateway",
		})
	})

	// API v2 routes with authentication
	v2 := s.router.Group("/v2")

	// Optional: Add API key auth for protected endpoints
	// For now, we'll make it optional to allow easier testing
	if deps.APIKeyRepo != nil {
		v2.Use(middleware.OptionalAPIKey(deps.APIKeyRepo))
	}

	// Register booking routes
	if s.bookingHandler != nil {
		s.bookingHandler.RegisterRoutes(v2)
	}

	// Register slot watch routes
	if s.slotWatchHandler != nil {
		s.slotWatchHandler.RegisterRoutes(v2)
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	ctx := context.Background()

	// Start WebSocket hub
	if s.wsHub != nil {
		go s.wsHub.Run()
		log.Println("[INFO] WebSocket hub started")
	}

	// Start scheduler
	if s.scheduler != nil {
		if err := s.scheduler.Start(ctx); err != nil {
			log.Printf("[ERROR] Failed to start scheduler: %v", err)
		} else {
			log.Println("[INFO] Scheduler started")
		}
	}

	// Start quick lock handler
	if s.quickLockHandler != nil {
		go s.quickLockHandler.Start(ctx)
		log.Println("[INFO] Quick lock handler started")
	}

	s.server = &http.Server{
		Addr:         s.cfg.Server.Addr(),
		Handler:      s.router,
		ReadTimeout:  s.cfg.Server.ReadTimeout,
		WriteTimeout: s.cfg.Server.WriteTimeout,
		IdleTimeout:  s.cfg.Server.IdleTimeout,
	}

	log.Printf("Starting server on %s", s.cfg.Server.Addr())
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	// Stop quick lock handler
	if s.quickLockHandler != nil {
		s.quickLockHandler.Stop()
		log.Println("[INFO] Quick lock handler stopped")
	}

	// Stop scheduler
	if s.scheduler != nil {
		if err := s.scheduler.Stop(); err != nil {
			log.Printf("[ERROR] Failed to stop scheduler: %v", err)
		} else {
			log.Println("[INFO] Scheduler stopped")
		}
	}

	if s.server == nil {
		return nil
	}
	log.Println("Shutting down server...")
	return s.server.Shutdown(ctx)
}

// Router returns the gin router (useful for testing)
func (s *Server) Router() *gin.Engine {
	return s.router
}

// corsMiddleware creates a CORS middleware
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, API-Version, X-API-Key, X-Carrier-Code, X-Correlation-ID")
		c.Header("Access-Control-Expose-Headers", "API-Version, X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset, X-Correlation-ID")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// apiVersionMiddleware adds API-Version header to responses
func apiVersionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("API-Version", "2.0.3")
		c.Next()
	}
}

// healthCheck handles health check requests
func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version":   "2.0.0",
	})
}

// Placeholder handlers for backward compatibility
func (s *Server) listBookings(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "List bookings - service not initialized",
	})
}

func (s *Server) createBooking(c *gin.Context) {
	c.JSON(http.StatusAccepted, gin.H{
		"message": "Create booking - service not initialized",
	})
}

func (s *Server) getBooking(c *gin.Context) {
	reference := c.Param("reference")
	c.JSON(http.StatusOK, gin.H{
		"message":   "Get booking - service not initialized",
		"reference": reference,
	})
}

func (s *Server) updateBooking(c *gin.Context) {
	reference := c.Param("reference")
	c.JSON(http.StatusOK, gin.H{
		"message":   "Update booking - service not initialized",
		"reference": reference,
	})
}

func (s *Server) cancelBooking(c *gin.Context) {
	reference := c.Param("reference")
	c.JSON(http.StatusOK, gin.H{
		"message":   "Cancel booking - service not initialized",
		"reference": reference,
	})
}
