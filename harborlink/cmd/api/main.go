package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourname/harborlink/internal/adapter"
	"github.com/yourname/harborlink/internal/booking"
	"github.com/yourname/harborlink/internal/core"
	"github.com/yourname/harborlink/internal/eventbus"
	"github.com/yourname/harborlink/internal/model"
	"github.com/yourname/harborlink/internal/notification"
	"github.com/yourname/harborlink/internal/repository"
	"github.com/yourname/harborlink/internal/scheduler"
	"github.com/yourname/harborlink/internal/service"
	"github.com/yourname/harborlink/pkg/cache"
	"github.com/yourname/harborlink/pkg/config"
)

var (
	configPath = flag.String("config", "", "Path to configuration file")
	version    = "dev"
	buildTime  = "unknown"
)

func main() {
	flag.Parse()

	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Display startup information
	displayStartupInfo(cfg)

	// Initialize components
	ctx := context.Background()

	// Initialize database
	db, err := initDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize cache
	redisClient, err := initCache(cfg)
	if err != nil {
		log.Printf("Warning: Redis not available: %v", err)
	}
	var bookingCache *cache.BookingCache
	if redisClient != nil {
		bookingCache = cache.NewBookingCache(redisClient)
	}

	// Initialize repositories
	bookingRepo := repository.NewBookingRepository(db)
	apiKeyRepo := repository.NewAPIKeyRepository(db)
	slotWatchRepo := repository.NewSlotWatchRepository(db)

	// Initialize adapter registry
	registry := initAdapterRegistry(cfg, bookingCache)

	// Initialize services
	carrierRouter := service.NewCarrierRouter(registry)
	bookingService := service.NewBookingService(bookingRepo, bookingCache, carrierRouter)

	// Initialize event bus
	eventBus := eventbus.NewEventBus(100)

	// Initialize WebSocket hub
	wsHub := notification.NewWebSocketHub()

	// Initialize scheduler
	sched := scheduler.NewScheduler(registry, eventBus, slotWatchRepo, cfg)

	// Initialize notifier service
	notifier := notification.NewNotifierService(wsHub)

	// Initialize quick lock handler
	notifyTimeout := cfg.SlotWatch.NotifyTimeout
	if notifyTimeout == 0 {
		notifyTimeout = 30 * time.Second
	}
	lockTimeout := cfg.SlotWatch.LockTimeout
	if lockTimeout == 0 {
		lockTimeout = 10 * time.Second
	}
	quickLockHandler := booking.NewQuickLockHandler(bookingService, slotWatchRepo, eventBus, notifier, notifyTimeout, lockTimeout)

	// Create server dependencies
	deps := &core.ServerDependencies{
		BookingService:   bookingService,
		BookingCache:     bookingCache,
		APIKeyRepo:       apiKeyRepo,
		Registry:         registry,
		EventBus:         eventBus,
		WsHub:            wsHub,
		Scheduler:        sched,
		SlotWatchRepo:    slotWatchRepo,
		QuickLockHandler: quickLockHandler,
	}

	// Create server with full dependencies
	server := core.NewServerWithDeps(cfg, deps)

	// Handle graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start server in goroutine
	go func() {
		log.Println("Starting HTTP server...")
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server stopped: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-quit
	log.Println("Shutting down...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
	log.Println("Server stopped")
}

func loadConfig() (*config.Config, error) {
	if *configPath != "" {
		return config.LoadFromPath(*configPath)
	}
	return config.Load("")
}

func displayStartupInfo(cfg *config.Config) {
	fmt.Println("========================================")
	fmt.Println("  HarborLink - DCSA API Aggregator")
	fmt.Printf("  Version: %s\n", version)
	fmt.Printf("  Build Time: %s\n", buildTime)
	fmt.Println("========================================")
	fmt.Printf("  Server: %s\n", cfg.Server.Addr())
	fmt.Printf("  Mode: %s\n", cfg.Server.Mode)
	fmt.Printf("  Database: %s:%d/%s\n", cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
	fmt.Printf("  Redis: %s\n", cfg.Redis.Addr())
	fmt.Printf("  Enabled Carriers: %d\n", countEnabledCarriers(cfg.Carriers))
	fmt.Println("========================================")
}

func countEnabledCarriers(carriers []config.CarrierConfig) int {
	count := 0
	for _, c := range carriers {
		if c.Enabled {
			count++
		}
	}
	return count
}

func initDatabase(cfg *config.Config) (*repository.Database, error) {
	db, err := repository.NewDatabase(&cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto migrate all models
	if err := db.Migrate(
		&model.BookingRecord{},
		&model.CarrierConfigRecord{},
		&model.APIKeyRecord{},
		&model.BookingAuditLog{},
		&model.SlotWatch{},
	); err != nil {
		log.Printf("Warning: Migration failed: %v", err)
	}

	return db, nil
}

func initCache(cfg *config.Config) (*cache.Client, error) {
	client, err := cache.NewClient(&cfg.Redis)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func initAdapterRegistry(cfg *config.Config, bookingCache *cache.BookingCache) *adapter.Registry {
	registry := adapter.NewRegistry(bookingCache)

	// Initialize adapters from configuration
	for _, carrierCfg := range cfg.Carriers {
		if carrierCfg.Enabled {
			// Use mock adapter for now (real adapters to be implemented later)
			baseAdapter := adapter.NewBaseAdapter(&carrierCfg, bookingCache)
			mockAdapter := adapter.NewMockAdapter(baseAdapter)
			registry.Register(mockAdapter)
			log.Printf("Registered adapter: %s (%s)", carrierCfg.Code, carrierCfg.Name)
		}
	}

	return registry
}
