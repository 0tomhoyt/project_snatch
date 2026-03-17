package scheduler

import (
	"context"
	"log"
	"sync"

	"github.com/yourname/harborlink/internal/adapter"
	"github.com/yourname/harborlink/internal/eventbus"
	"github.com/yourname/harborlink/internal/repository"
	"github.com/yourname/harborlink/pkg/config"
)

// Scheduler manages carrier pollers and distributes watch requests
type Scheduler struct {
	registry  *adapter.Registry
	eventBus  *eventbus.EventBus
	watchRepo repository.SlotWatchRepository
	config    *config.Config

	pollers map[string]*CarrierPoller // carrierCode -> poller
	mu      sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewScheduler creates a new scheduler
func NewScheduler(
	registry *adapter.Registry,
	eventBus *eventbus.EventBus,
	watchRepo repository.SlotWatchRepository,
	cfg *config.Config,
) *Scheduler {
	return &Scheduler{
		registry:  registry,
		eventBus:  eventBus,
		watchRepo: watchRepo,
		config:    cfg,
		pollers:   make(map[string]*CarrierPoller),
	}
}

// Start starts all enabled carrier pollers
func (s *Scheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.ctx, s.cancel = context.WithCancel(ctx)

	// Get carrier configurations
	carrierConfigs := s.getCarrierConfigMap()

	// Start pollers for enabled carriers
	for code, carrier := range s.registry.GetAll() {
		if !carrier.IsEnabled() {
			continue
		}

		carrierCfg, ok := carrierConfigs[code]
		if !ok {
			log.Printf("[WARN] Scheduler: no config found for carrier %s, using defaults", code)
			carrierCfg = &config.CarrierConfig{
				Code:         code,
				PollEnabled:  true,
				PollInterval: defaultPollInterval,
				PollTimeout:  defaultPollTimeout,
			}
		}

		if !carrierCfg.PollEnabled {
			log.Printf("[INFO] Scheduler: polling disabled for carrier %s", code)
			continue
		}

		poller := NewCarrierPoller(carrier, carrierCfg, s.eventBus, s.watchRepo)
		s.pollers[code] = poller

		s.wg.Add(1)
		go func(p *CarrierPoller) {
			defer s.wg.Done()
			p.Start(s.ctx)
		}(poller)

		log.Printf("[INFO] Scheduler: started poller for carrier %s (interval: %s)", code, carrierCfg.PollInterval)
	}

	// Load existing active watches
	if err := s.loadActiveWatches(); err != nil {
		log.Printf("[ERROR] Scheduler: failed to load active watches: %v", err)
	}

	return nil
}

// Stop stops all pollers
func (s *Scheduler) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cancel != nil {
		s.cancel()
	}

	// Stop all pollers
	for _, poller := range s.pollers {
		poller.Stop()
	}

	// Wait for all goroutines to finish
	s.wg.Wait()

	log.Println("[INFO] Scheduler: stopped all pollers")
	return nil
}

// RegisterWatch registers a new watch request to relevant pollers
func (s *Scheduler) RegisterWatch(watchID uint, carrierCodes []string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, code := range carrierCodes {
		poller, ok := s.pollers[code]
		if !ok {
			log.Printf("[WARN] Scheduler: no poller for carrier %s", code)
			continue
		}
		poller.AddWatch(watchID)
	}

	return nil
}

// UnregisterWatch removes a watch from all pollers
func (s *Scheduler) UnregisterWatch(watchID uint, carrierCodes []string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, code := range carrierCodes {
		if poller, ok := s.pollers[code]; ok {
			poller.RemoveWatch(watchID)
		}
	}

	return nil
}

// GetPollerStats returns statistics for all pollers
func (s *Scheduler) GetPollerStats() map[string]PollerStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := make(map[string]PollerStats)
	for code, poller := range s.pollers {
		stats[code] = poller.GetStats()
	}
	return stats
}

// loadActiveWatches loads all active watches into pollers
func (s *Scheduler) loadActiveWatches() error {
	watches, err := s.watchRepo.ListActive(s.ctx)
	if err != nil {
		return err
	}

	for _, watch := range watches {
		for _, code := range watch.CarrierCodes {
			if poller, ok := s.pollers[code]; ok {
				poller.AddWatch(watch.ID)
			}
		}
	}

	log.Printf("[INFO] Scheduler: loaded %d active watches", len(watches))
	return nil
}

// getCarrierConfigMap creates a map of carrier code to config
func (s *Scheduler) getCarrierConfigMap() map[string]*config.CarrierConfig {
	configMap := make(map[string]*config.CarrierConfig)
	for i := range s.config.Carriers {
		cfg := &s.config.Carriers[i]
		configMap[cfg.Code] = cfg
	}
	return configMap
}

// TriggerPoll triggers an immediate poll for a specific carrier (for testing)
func (s *Scheduler) TriggerPoll(carrierCode string) error {
	s.mu.RLock()
	poller, ok := s.pollers[carrierCode]
	s.mu.RUnlock()

	if !ok {
		return ErrPollerNotFound
	}

	return poller.TriggerPoll()
}

// CleanupExpired runs cleanup of expired watches
func (s *Scheduler) CleanupExpired(ctx context.Context) (int64, error) {
	return s.watchRepo.CleanupExpired(ctx)
}
