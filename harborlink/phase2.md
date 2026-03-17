# Phase 2: Infrastructure Layer

**Status**: Completed
**Date**: 2026-03-02

---

## Overview

Phase 2 建立了 HarborLink 的核心基础设施层。

## Project Structure

```
/harborlink
├── cmd/api/main.go              # Entry point
├── config.yaml                  # Configuration
├── api/DCSA-2.0.yaml           # DCSA OpenAPI spec
├── internal/
│   ├── adapter/                # Carrier adapters
│   │   ├── interface.go        # CarrierAdapter interface
│   │   ├── base.go             # BaseAdapter
│   │   ├── mock.go             # MockAdapter (testing)
│   │   └── registry.go         # Adapter registry
│   ├── core/server.go          # Gin HTTP server
│   ├── model/                  # DCSA data models
│   │   ├── booking.go, party.go, equipment.go
│   │   ├── location.go, commodity.go, transport.go
│   │   ├── enums.go, db_models.go
│   └── repository/             # Database layer
│       ├── database.go
│       ├── booking_repository.go
│       └── apikey_repository.go
└── pkg/
    ├── cache/redis.go          # Redis client
    └── config/config.go        # Viper config
```

## Components

### 1. Configuration (`pkg/config/`)
- YAML config + environment variable override
- Prefix: `HARBORLINK_`
- Default values built-in

### 2. Data Models (`internal/model/`)
- DCSA 2.0 compliant entities
- GORM database models
- Enum types (BookingStatus, ReceiptDeliveryType, etc.)

### 3. Repository (`internal/repository/`)
- BookingRepository: CRUD + filtering
- APIKeyRepository: multi-tenant key management
- Soft delete support

### 4. Cache (`pkg/cache/`)
- Redis client wrapper
- BookingCache: get/set/status operations
- Distributed locking for carrier API calls
- Rate limiting support

### 5. HTTP Server (`internal/core/`)
- Gin framework
- Routes: `/v2/health`, `/v2/bookings/*` (placeholders)
- Middleware: recovery, logging, CORS

### 6. Adapter Framework (`internal/adapter/`)
- CarrierAdapter interface
- BaseAdapter with common functionality
- MockAdapter for testing
- Registry for adapter management

## Dependencies

| Package | Purpose |
|---------|---------|
| gin-gonic/gin | HTTP framework |
| spf13/viper | Configuration |
| gorm.io/gorm | ORM |
| redis/go-redis | Cache |

## Tests

```bash
go test ./...
# All packages pass
```

## Next Steps (Phase 3)

- [ ] Service layer (booking_service, carrier_router)
- [ ] HTTP handlers (booking_handler, webhook_handler)
- [ ] Middleware (auth, rate limiting)
- [ ] Request validation
- [ ] Real carrier adapters (Maersk, MSC, CMA CGM)
