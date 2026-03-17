package repository

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/yourname/harborlink/internal/model"
	"github.com/yourname/harborlink/pkg/config"
)

// mockDB creates an in-memory sqlite database for testing
func mockDB(t *testing.T) *Database {
	t.Helper()
	// Use file::memory:?cache=shared mode to share in-memory db
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open mock database: %v", err)
	}

	// Auto migrate schemas
	if err := db.AutoMigrate(&model.BookingRecord{}, &model.APIKeyRecord{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	return &Database{db}
}

// uniqueRef generates a unique reference for testing
func uniqueRef(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

func TestDatabaseDSN(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		Name:     "harborlink",
		User:     "postgres",
		Password: "secret",
		SSLMode:  "disable",
	}

	dsn := cfg.DSN()
	expected := "host=localhost port=5432 user=postgres password=secret dbname=harborlink sslmode=disable"
	if dsn != expected {
		t.Errorf("expected DSN '%s', got '%s'", expected, dsn)
	}
}

func TestBookingRepository_Create(t *testing.T) {
	db := mockDB(t)
	repo := NewBookingRepository(db)

	ref := uniqueRef("test-create")
	booking := &model.BookingRecord{
		CarrierBookingRequestReference: ref,
		BookingStatus:                 model.BookingStatusReceived,
		ReceiptTypeAtOrigin:           model.ReceiptDeliveryTypeCY,
	}

	err := repo.Create(context.Background(), booking)
	if err != nil {
		t.Errorf("failed to create booking: %v", err)
	}

	if booking.ID == 0 {
		t.Error("expected booking ID to be set after creation")
	}
}

func TestBookingRepository_GetByReference(t *testing.T) {
	db := mockDB(t)
	repo := NewBookingRepository(db)

	ref := uniqueRef("test-get")
	// Create test booking
	booking := &model.BookingRecord{
		CarrierBookingRequestReference: ref,
		BookingStatus:                 model.BookingStatusReceived,
	}
	repo.Create(context.Background(), booking)

	// Retrieve booking
	found, err := repo.GetByReference(context.Background(), ref)
	if err != nil {
		t.Errorf("failed to get booking: %v", err)
	}

	if found.CarrierBookingRequestReference != ref {
		t.Errorf("expected carrierBookingRequestReference %s, got %s", ref, found.CarrierBookingRequestReference)
	}

	// Test not found
	_, err = repo.GetByReference(context.Background(), "non-existent-ref-"+time.Now().Format("20060102150405"))
	if !errors.Is(err, ErrBookingNotFound) {
		t.Errorf("expected ErrBookingNotFound, got %v", err)
	}
}

func TestBookingRepository_UpdateStatus(t *testing.T) {
	db := mockDB(t)
	repo := NewBookingRepository(db)

	ref := uniqueRef("test-update")
	// Create test booking
	booking := &model.BookingRecord{
		CarrierBookingRequestReference: ref,
		BookingStatus:                 model.BookingStatusReceived,
	}
	repo.Create(context.Background(), booking)

	// Update status
	err := repo.UpdateStatus(context.Background(), ref, model.BookingStatusConfirmed)
	if err != nil {
		t.Errorf("failed to update status: %v", err)
	}

	// Verify update
	found, err := repo.GetByReference(context.Background(), ref)
	if err != nil {
		t.Errorf("failed to get updated booking: %v", err)
	}

	if found.BookingStatus != model.BookingStatusConfirmed {
		t.Errorf("expected status %s, got %s", model.BookingStatusConfirmed, found.BookingStatus)
	}
}

func TestBookingFilter_Pagination(t *testing.T) {
	// Use unique database for this test
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open mock database: %v", err)
	}
	if err := db.AutoMigrate(&model.BookingRecord{}, &model.APIKeyRecord{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	database := &Database{db}
	repo := NewBookingRepository(database)

	// Create multiple bookings with unique references
	createdCount := 15
	for i := 0; i < createdCount; i++ {
		booking := &model.BookingRecord{
			CarrierBookingRequestReference: uniqueRef(fmt.Sprintf("test-pagination-%d", i)),
			BookingStatus:                 model.BookingStatusReceived,
		}
		err := repo.Create(context.Background(), booking)
		if err != nil {
			t.Fatalf("failed to create booking %d: %v", i, err)
		}
	}

	// Test pagination
	filter := &BookingFilter{
		Page:     1,
		PageSize: 10,
	}

	bookings, total, err := repo.List(context.Background(), filter)
	if err != nil {
		t.Errorf("failed to list bookings: %v", err)
	}

	// Verify at least createdCount records exist
	if total < int64(createdCount) {
		t.Errorf("expected at least %d total, got %d", createdCount, total)
	}

	// Verify page size is respected
	if len(bookings) > 10 {
		t.Errorf("expected at most 10 bookings, got %d", len(bookings))
	}
}

func TestAPIKeyRepository_GetByKey(t *testing.T) {
	db := mockDB(t)
	repo := NewAPIKeyRepository(db)

	key := uniqueRef("test-apikey")
	// Create test API key
	apiKey := &model.APIKeyRecord{
		Key:      key,
		Name:     "Test Key",
		TenantID: "tenant-001",
		Active:   true,
	}
	repo.Create(context.Background(), apiKey)

	// Retrieve API key
	found, err := repo.GetByKey(context.Background(), key)
	if err != nil {
		t.Errorf("failed to get API key: %v", err)
	}

	if found.Key != key {
		t.Error("expected key to match")
	}

	// Test not found
	_, err = repo.GetByKey(context.Background(), "non-existent-key-"+time.Now().Format("20060102150405"))
	if !errors.Is(err, ErrAPIKeyNotFound) {
		t.Errorf("expected ErrAPIKeyNotFound, got %v", err)
	}
}

func TestAPIKeyRepository_Deactivate(t *testing.T) {
	db := mockDB(t)
	repo := NewAPIKeyRepository(db)

	key := uniqueRef("test-deactivate")
	// Create test API key
	apiKey := &model.APIKeyRecord{
		Key:      key,
		Name:     "Test Key",
		TenantID: "tenant-001",
		Active:   true,
	}
	repo.Create(context.Background(), apiKey)

	// Deactivate
	err := repo.Deactivate(context.Background(), key)
	if err != nil {
		t.Errorf("failed to deactivate API key: %v", err)
	}

	// Verify deactivated
	_, err = repo.GetByKey(context.Background(), key)
	if !errors.Is(err, ErrAPIKeyInactive) {
		t.Errorf("expected ErrAPIKeyInactive, got %v", err)
	}
}
