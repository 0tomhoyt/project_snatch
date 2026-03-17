# HarborLink - DCSA API Aggregator

API gateway that aggregates DCSA booking interfaces from multiple shipping carriers.

## Tech Stack

| Component | Technology |
|-----------|------------|
| HTTP | Gin |
| Database | PostgreSQL + GORM |
| Cache | Redis |
| Config | Viper |

## Project Structure

```
/harborlink
├── cmd/api/main.go           # Entry point
├── config.yaml               # Configuration
├── api/DCSA-2.0.yaml        # DCSA OpenAPI spec
├── internal/
│   ├── adapter/             # Carrier adapters
│   ├── core/server.go       # HTTP server
│   ├── model/               # Data models
│   └── repository/          # Database access
└── pkg/
    ├── cache/               # Redis client
    └── config/              # Configuration
```

## Progress

- [x] Phase 1: DCSA domain modeling
- [x] Phase 2: Infrastructure (config, db, cache, server, adapters)
- [ ] Phase 3: Business logic (services, handlers, middleware)

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v2/health` | Health check |
| POST | `/v2/bookings` | Create booking |
| GET | `/v2/bookings/{ref}` | Get booking |
| PUT | `/v2/bookings/{ref}` | Update booking |
| DELETE | `/v2/bookings/{ref}` | Cancel booking |

## Booking Status

`RECEIVED` → `PENDING_UPDATE` → `CONFIRMED` → `COMPLETED`
                     ↓                ↓
              `REJECTED`         `DECLINED`
                                 `CANCELLED`

## Key Interfaces

### CarrierAdapter
```go
type CarrierAdapter interface {
    CreateBooking(ctx, *model.CreateBooking) (*model.CreateBookingResponse, error)
    GetBooking(ctx, reference string) (*model.Booking, error)
    UpdateBooking(ctx, ref string, *model.UpdateBooking) (*model.Booking, error)
    CancelBooking(ctx, ref string, *model.CancelBookingRequest) error
    GetCarrierCode() string
    GetCarrierName() string
    IsEnabled() bool
    HealthCheck(ctx) error
}
```

### BookingRepository
```go
type BookingRepository interface {
    Create(ctx, *model.BookingRecord) error
    GetByReference(ctx, ref string) (*model.BookingRecord, error)
    Update(ctx, *model.BookingRecord) error
    UpdateStatus(ctx, ref string, status model.BookingStatus) error
    List(ctx, *BookingFilter) ([]model.BookingRecord, int64, error)
}
```

### BookingCache
```go
// Cache operations
Get(ctx, ref string, dest interface{}) error
Set(ctx, ref string, value interface{}, ttl time.Duration) error

// Status caching
GetStatus(ctx, ref string) (string, error)
SetStatus(ctx, ref string, status string, ttl time.Duration) error

// Distributed locking
AcquireCarrierLock(ctx, code string, ttl time.Duration) (bool, error)
ReleaseCarrierLock(ctx, code string) error
```

## Configuration

```yaml
# config.yaml
server:
  port: 8080
  mode: debug

database:
  host: localhost
  port: 5432
  name: harborlink
  user: postgres

redis:
  host: localhost
  port: 6379

carriers:
  - name: maersk
    code: MAEU
    enabled: true
    rate_limit: 100
```

Environment variables: `HARBORLINK_SERVER_PORT`, `HARBORLINK_DATABASE_HOST`, etc.

## Commands

```bash
# Run
go run ./cmd/api/main.go

# Build
go build -o bin/harborlink ./cmd/api

# Test
go test ./...

# Test with coverage
go test -cover ./...
```

## Coding Conventions

- Use `fmt.Errorf("...: %w", err)` for error wrapping
- Use `context.Context` for cancellation
- Group imports: stdlib → external → internal
- Use `AdapterError` for carrier-specific errors

## References

- [DCSA Booking API 2.0](https://dcsa.org/standards/booking-process/)
- [Gin Documentation](https://gin-gonic.com/docs/)
- [GORM Documentation](https://gorm.io/docs/)
