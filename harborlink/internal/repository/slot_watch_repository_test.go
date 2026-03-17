package repository

import (
	"context"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/yourname/harborlink/internal/model"
)

// mockSlotWatchDB creates an in-memory SQLite database for testing
func mockSlotWatchDB(t *testing.T) *Database {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}

	// Auto migrate
	if err := db.AutoMigrate(&model.SlotWatch{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	return &Database{db}
}

// uniqueWatchRef generates a unique watch reference for testing
func uniqueWatchRef(t *testing.T) string {
	t.Helper()
	return "WATCH-TEST-" + time.Now().Format("20060102150405")
}

func TestSlotWatchRepository_Create(t *testing.T) {
	db := mockSlotWatchDB(t)
	repo := NewSlotWatchRepository(db)
	ctx := context.Background()

	carrierCodes := []string{"MAEU", "MSCU"}

	watch := &model.SlotWatch{
		TenantID:        "tenant-1",
		Reference:       uniqueWatchRef(t),
		POL:             "CNSHA",
		POD:             "USLAX",
		CarrierCodes:    carrierCodes,
		LockStrategy:    model.LockStrategyAutoLock,
		Status:          model.WatchStatusActive,
		Priority:        5,
		MaxRetries:      3,
	}

	err := repo.Create(ctx, watch)
	if err != nil {
		t.Errorf("failed to create slot watch: %v", err)
	}

	// Verify the watch was created
	retrieved, err := repo.GetByReference(ctx, watch.Reference)
	if err != nil {
		t.Errorf("failed to retrieve created watch: %v", err)
	}
	if retrieved.Reference != watch.Reference {
		t.Errorf("expected reference %s, got %s", watch.Reference, retrieved.Reference)
	}
}

func TestSlotWatchRepository_GetByID(t *testing.T) {
	db := mockSlotWatchDB(t)
	repo := NewSlotWatchRepository(db)
	ctx := context.Background()

	// Test not found
	_, err := repo.GetByID(ctx, 999)
	if err != ErrSlotWatchNotFound {
		t.Errorf("expected ErrSlotWatchNotFound, got %v", err)
	}

	// Create and retrieve
	watch := &model.SlotWatch{
		TenantID:     "tenant-1",
		Reference:    uniqueWatchRef(t),
		POL:          "CNSHA",
		POD:          "USLAX",
		CarrierCodes: []string{"MAEU"},
		LockStrategy: model.LockStrategyAutoLock,
		Status:       model.WatchStatusActive,
	}
	_ = repo.Create(ctx, watch)

	retrieved, err := repo.GetByID(ctx, watch.ID)
	if err != nil {
		t.Errorf("failed to get watch by ID: %v", err)
	}
	if retrieved.ID != watch.ID {
		t.Errorf("expected ID %d, got %d", watch.ID, retrieved.ID)
	}
}

func TestSlotWatchRepository_GetByReference(t *testing.T) {
	db := mockSlotWatchDB(t)
	repo := NewSlotWatchRepository(db)
	ctx := context.Background()

	ref := uniqueWatchRef(t)

	// Test not found
	_, err := repo.GetByReference(ctx, "non-existent")
	if err != ErrSlotWatchNotFound {
		t.Errorf("expected ErrSlotWatchNotFound, got %v", err)
	}

	// Create and retrieve
	watch := &model.SlotWatch{
		TenantID:     "tenant-1",
		Reference:    ref,
		POL:          "CNSHA",
		POD:          "USLAX",
		CarrierCodes: []string{"MAEU"},
		LockStrategy: model.LockStrategyAutoLock,
		Status:       model.WatchStatusActive,
	}
	_ = repo.Create(ctx, watch)

	retrieved, err := repo.GetByReference(ctx, ref)
	if err != nil {
		t.Errorf("failed to get watch by reference: %v", err)
	}
	if retrieved.Reference != ref {
		t.Errorf("expected reference %s, got %s", ref, retrieved.Reference)
	}
}

func TestSlotWatchRepository_GetByTenantReference(t *testing.T) {
	db := mockSlotWatchDB(t)
	repo := NewSlotWatchRepository(db)
	ctx := context.Background()

	ref := uniqueWatchRef(t)

	// Create watch
	watch := &model.SlotWatch{
		TenantID:     "tenant-1",
		Reference:    ref,
		POL:          "CNSHA",
		POD:          "USLAX",
		CarrierCodes: []string{"MAEU"},
		LockStrategy: model.LockStrategyAutoLock,
		Status:       model.WatchStatusActive,
	}
	_ = repo.Create(ctx, watch)

	// Retrieve with correct tenant
	retrieved, err := repo.GetByTenantReference(ctx, "tenant-1", ref)
	if err != nil {
		t.Errorf("failed to get watch: %v", err)
	}
	if retrieved.TenantID != "tenant-1" {
		t.Errorf("expected tenantID tenant-1, got %s", retrieved.TenantID)
	}

	// Try with wrong tenant
	_, err = repo.GetByTenantReference(ctx, "tenant-2", ref)
	if err != ErrSlotWatchNotFound {
		t.Errorf("expected ErrSlotWatchNotFound for wrong tenant, got %v", err)
	}
}

func TestSlotWatchRepository_UpdateStatus(t *testing.T) {
	db := mockSlotWatchDB(t)
	repo := NewSlotWatchRepository(db)
	ctx := context.Background()

	// Create watch
	watch := &model.SlotWatch{
		TenantID:     "tenant-1",
		Reference:    uniqueWatchRef(t),
		POL:          "CNSHA",
		POD:          "USLAX",
		CarrierCodes: []string{"MAEU"},
		LockStrategy: model.LockStrategyAutoLock,
		Status:       model.WatchStatusActive,
	}
	_ = repo.Create(ctx, watch)

	// Update status
	err := repo.UpdateStatus(ctx, watch.ID, model.WatchStatusTriggered)
	if err != nil {
		t.Errorf("failed to update status: %v", err)
	}

	// Verify update
	retrieved, _ := repo.GetByID(ctx, watch.ID)
	if retrieved.Status != model.WatchStatusTriggered {
		t.Errorf("expected status TRIGGERED, got %s", retrieved.Status)
	}

	// Test update non-existent
	err = repo.UpdateStatus(ctx, 999, model.WatchStatusCancelled)
	if err != ErrSlotWatchNotFound {
		t.Errorf("expected ErrSlotWatchNotFound, got %v", err)
	}
}

func TestSlotWatchRepository_Delete(t *testing.T) {
	db := mockSlotWatchDB(t)
	repo := NewSlotWatchRepository(db)
	ctx := context.Background()

	ref := uniqueWatchRef(t)

	// Create watch
	watch := &model.SlotWatch{
		TenantID:     "tenant-1",
		Reference:    ref,
		POL:          "CNSHA",
		POD:          "USLAX",
		CarrierCodes: []string{"MAEU"},
		LockStrategy: model.LockStrategyAutoLock,
		Status:       model.WatchStatusActive,
	}
	_ = repo.Create(ctx, watch)

	// Delete
	err := repo.Delete(ctx, ref)
	if err != nil {
		t.Errorf("failed to delete watch: %v", err)
	}

	// Verify deleted
	_, err = repo.GetByReference(ctx, ref)
	if err != ErrSlotWatchNotFound {
		t.Errorf("expected ErrSlotWatchNotFound after delete, got %v", err)
	}

	// Test delete non-existent
	err = repo.Delete(ctx, "non-existent")
	if err != ErrSlotWatchNotFound {
		t.Errorf("expected ErrSlotWatchNotFound, got %v", err)
	}
}

func TestSlotWatchRepository_List(t *testing.T) {
	db := mockSlotWatchDB(t)
	repo := NewSlotWatchRepository(db)
	ctx := context.Background()

	// Create multiple watches
	for i := 0; i < 5; i++ {
		watch := &model.SlotWatch{
			TenantID:     "tenant-1",
			Reference:    uniqueWatchRef(t),
			POL:          "CNSHA",
			POD:          "USLAX",
			CarrierCodes: []string{"MAEU"},
			LockStrategy: model.LockStrategyAutoLock,
			Status:       model.WatchStatusActive,
			Priority:     5 + i,
		}
		_ = repo.Create(ctx, watch)
	}

	// Create watch for different tenant
	watch := &model.SlotWatch{
		TenantID:     "tenant-2",
		Reference:    uniqueWatchRef(t),
		POL:          "CNSHA",
		POD:          "USLAX",
		CarrierCodes: []string{"MAEU"},
		LockStrategy: model.LockStrategyAutoLock,
		Status:       model.WatchStatusActive,
	}
	_ = repo.Create(ctx, watch)

	// List with tenant filter
	filter := &SlotWatchFilter{
		TenantID: "tenant-1",
		Page:     1,
		PageSize: 10,
	}
	watches, total, err := repo.List(ctx, filter)
	if err != nil {
		t.Errorf("failed to list watches: %v", err)
	}
	if total != 5 {
		t.Errorf("expected 5 watches, got %d", total)
	}
	if len(watches) != 5 {
		t.Errorf("expected 5 watches in result, got %d", len(watches))
	}

	// Test pagination
	filter.Page = 1
	filter.PageSize = 2
	watches, total, err = repo.List(ctx, filter)
	if err != nil {
		t.Errorf("failed to list watches with pagination: %v", err)
	}
	if len(watches) != 2 {
		t.Errorf("expected 2 watches with PageSize=2, got %d", len(watches))
	}
}

func TestSlotWatchRepository_ListActiveByCarrier(t *testing.T) {
	db := mockSlotWatchDB(t)
	repo := NewSlotWatchRepository(db)
	ctx := context.Background()

	// Create active watch for MAEU
	watch1 := &model.SlotWatch{
		TenantID:     "tenant-1",
		Reference:    uniqueWatchRef(t),
		POL:          "CNSHA",
		POD:          "USLAX",
		CarrierCodes: []string{"MAEU"},
		LockStrategy: model.LockStrategyAutoLock,
		Status:       model.WatchStatusActive,
	}
	_ = repo.Create(ctx, watch1)

	// Create active watch for both MAEU and MSCU
	watch2 := &model.SlotWatch{
		TenantID:     "tenant-1",
		Reference:    uniqueWatchRef(t),
		POL:          "CNSHA",
		POD:          "USLAX",
		CarrierCodes: []string{"MAEU", "MSCU"},
		LockStrategy: model.LockStrategyAutoLock,
		Status:       model.WatchStatusActive,
	}
	_ = repo.Create(ctx, watch2)

	// Create triggered watch (should not be returned)
	watch3 := &model.SlotWatch{
		TenantID:     "tenant-1",
		Reference:    uniqueWatchRef(t),
		POL:          "CNSHA",
		POD:          "USLAX",
		CarrierCodes: []string{"MAEU"},
		LockStrategy: model.LockStrategyAutoLock,
		Status:       model.WatchStatusTriggered,
	}
	_ = repo.Create(ctx, watch3)

	// List active for MAEU
	watches, err := repo.ListActiveByCarrier(ctx, "MAEU")
	if err != nil {
		t.Errorf("failed to list active watches: %v", err)
	}
	if len(watches) != 2 {
		t.Errorf("expected 2 active watches for MAEU, got %d", len(watches))
	}
}

func TestSlotWatchRepository_ListActive(t *testing.T) {
	db := mockSlotWatchDB(t)
	repo := NewSlotWatchRepository(db)
	ctx := context.Background()

	// Create active watches
	for i := 0; i < 3; i++ {
		watch := &model.SlotWatch{
			TenantID:     "tenant-1",
			Reference:    uniqueWatchRef(t),
			POL:          "CNSHA",
			POD:          "USLAX",
			CarrierCodes: []string{"MAEU"},
			LockStrategy: model.LockStrategyAutoLock,
			Status:       model.WatchStatusActive,
		}
		_ = repo.Create(ctx, watch)
	}

	// Create non-active watch
	watch := &model.SlotWatch{
		TenantID:     "tenant-1",
		Reference:    uniqueWatchRef(t),
		POL:          "CNSHA",
		POD:          "USLAX",
		CarrierCodes: []string{"MAEU"},
		LockStrategy: model.LockStrategyAutoLock,
		Status:       model.WatchStatusCancelled,
	}
	_ = repo.Create(ctx, watch)

	watches, err := repo.ListActive(ctx)
	if err != nil {
		t.Errorf("failed to list active watches: %v", err)
	}
	if len(watches) != 3 {
		t.Errorf("expected 3 active watches, got %d", len(watches))
	}
}

func TestSlotWatchRepository_MarkTriggered(t *testing.T) {
	db := mockSlotWatchDB(t)
	repo := NewSlotWatchRepository(db)
	ctx := context.Background()

	// Create watch
	watch := &model.SlotWatch{
		TenantID:     "tenant-1",
		Reference:    uniqueWatchRef(t),
		POL:          "CNSHA",
		POD:          "USLAX",
		CarrierCodes: []string{"MAEU"},
		LockStrategy: model.LockStrategyAutoLock,
		Status:       model.WatchStatusActive,
	}
	_ = repo.Create(ctx, watch)

	// Mark triggered
	bookingRef := "BK123456"
	err := repo.MarkTriggered(ctx, watch.ID, "MAEU", bookingRef)
	if err != nil {
		t.Errorf("failed to mark triggered: %v", err)
	}

	// Verify
	retrieved, _ := repo.GetByID(ctx, watch.ID)
	if retrieved.Status != model.WatchStatusTriggered {
		t.Errorf("expected status TRIGGERED, got %s", retrieved.Status)
	}
	if retrieved.TriggeredByCarrier != "MAEU" {
		t.Errorf("expected triggered by MAEU, got %s", retrieved.TriggeredByCarrier)
	}
	if retrieved.BookingRef == nil || *retrieved.BookingRef != bookingRef {
		t.Errorf("expected booking ref %s", bookingRef)
	}
}

func TestSlotWatchRepository_IncrementRetry(t *testing.T) {
	db := mockSlotWatchDB(t)
	repo := NewSlotWatchRepository(db)
	ctx := context.Background()

	// Create watch
	watch := &model.SlotWatch{
		TenantID:     "tenant-1",
		Reference:    uniqueWatchRef(t),
		POL:          "CNSHA",
		POD:          "USLAX",
		CarrierCodes: []string{"MAEU"},
		LockStrategy: model.LockStrategyAutoLock,
		Status:       model.WatchStatusActive,
		RetryCount:   0,
	}
	_ = repo.Create(ctx, watch)

	// Increment retry
	err := repo.IncrementRetry(ctx, watch.ID)
	if err != nil {
		t.Errorf("failed to increment retry: %v", err)
	}

	// Verify
	retrieved, _ := repo.GetByID(ctx, watch.ID)
	if retrieved.RetryCount != 1 {
		t.Errorf("expected retry count 1, got %d", retrieved.RetryCount)
	}

	// Increment again
	_ = repo.IncrementRetry(ctx, watch.ID)
	retrieved, _ = repo.GetByID(ctx, watch.ID)
	if retrieved.RetryCount != 2 {
		t.Errorf("expected retry count 2, got %d", retrieved.RetryCount)
	}
}

func TestSlotWatchRepository_CleanupExpired(t *testing.T) {
	db := mockSlotWatchDB(t)
	repo := NewSlotWatchRepository(db)
	ctx := context.Background()

	past := time.Now().Add(-1 * time.Hour)
	future := time.Now().Add(1 * time.Hour)

	// Create expired watch
	watch1 := &model.SlotWatch{
		TenantID:     "tenant-1",
		Reference:    uniqueWatchRef(t),
		POL:          "CNSHA",
		POD:          "USLAX",
		CarrierCodes: []string{"MAEU"},
		LockStrategy: model.LockStrategyAutoLock,
		Status:       model.WatchStatusActive,
		ExpiresAt:    &past,
	}
	_ = repo.Create(ctx, watch1)

	// Create non-expired watch
	watch2 := &model.SlotWatch{
		TenantID:     "tenant-1",
		Reference:    uniqueWatchRef(t),
		POL:          "CNSHA",
		POD:          "USLAX",
		CarrierCodes: []string{"MAEU"},
		LockStrategy: model.LockStrategyAutoLock,
		Status:       model.WatchStatusActive,
		ExpiresAt:    &future,
	}
	_ = repo.Create(ctx, watch2)

	// Cleanup
	count, err := repo.CleanupExpired(ctx)
	if err != nil {
		t.Errorf("failed to cleanup expired: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 expired watch cleaned, got %d", count)
	}

	// Verify expired watch is marked
	retrieved, _ := repo.GetByID(ctx, watch1.ID)
	if retrieved.Status != model.WatchStatusExpired {
		t.Errorf("expected status EXPIRED, got %s", retrieved.Status)
	}

	// Verify non-expired watch is still active
	retrieved, _ = repo.GetByID(ctx, watch2.ID)
	if retrieved.Status != model.WatchStatusActive {
		t.Errorf("expected status ACTIVE, got %s", retrieved.Status)
	}
}
