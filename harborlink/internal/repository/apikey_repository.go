package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/yourname/harborlink/internal/model"
)

var (
	ErrAPIKeyNotFound = errors.New("api key not found")
	ErrAPIKeyExpired  = errors.New("api key expired")
	ErrAPIKeyInactive = errors.New("api key inactive")
)

// APIKeyRepository defines the interface for API key data access
type APIKeyRepository interface {
	Create(ctx context.Context, apiKey *model.APIKeyRecord) error
	GetByKey(ctx context.Context, key string) (*model.APIKeyRecord, error)
	GetByID(ctx context.Context, id uint) (*model.APIKeyRecord, error)
	Update(ctx context.Context, apiKey *model.APIKeyRecord) error
	UpdateLastUsed(ctx context.Context, key string) error
	Delete(ctx context.Context, id uint) error
	Deactivate(ctx context.Context, key string) error
	List(ctx context.Context, tenantID string) ([]model.APIKeyRecord, error)
}

// apiKeyRepository implements APIKeyRepository
type apiKeyRepository struct {
	db *Database
}

// NewAPIKeyRepository creates a new API key repository
func NewAPIKeyRepository(db *Database) APIKeyRepository {
	return &apiKeyRepository{db: db}
}

// Create creates a new API key
func (r *apiKeyRepository) Create(ctx context.Context, apiKey *model.APIKeyRecord) error {
	return r.db.WithContext(ctx).Create(apiKey).Error
}

// GetByKey gets an API key by its key value
func (r *apiKeyRepository) GetByKey(ctx context.Context, key string) (*model.APIKeyRecord, error) {
	var apiKey model.APIKeyRecord
	result := r.db.WithContext(ctx).Where("key = ?", key).First(&apiKey)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrAPIKeyNotFound
		}
		return nil, result.Error
	}

	// Check if key is active
	if !apiKey.Active {
		return nil, ErrAPIKeyInactive
	}

	// Check if key is expired
	if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
		return nil, ErrAPIKeyExpired
	}

	return &apiKey, nil
}

// GetByID gets an API key by its ID
func (r *apiKeyRepository) GetByID(ctx context.Context, id uint) (*model.APIKeyRecord, error) {
	var apiKey model.APIKeyRecord
	result := r.db.WithContext(ctx).First(&apiKey, id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrAPIKeyNotFound
		}
		return nil, result.Error
	}

	return &apiKey, nil
}

// Update updates an API key
func (r *apiKeyRepository) Update(ctx context.Context, apiKey *model.APIKeyRecord) error {
	result := r.db.WithContext(ctx).Save(apiKey)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrAPIKeyNotFound
	}
	return nil
}

// UpdateLastUsed updates the last used timestamp for an API key
func (r *apiKeyRepository) UpdateLastUsed(ctx context.Context, key string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&model.APIKeyRecord{}).
		Where("key = ?", key).
		Update("last_used_at", &now)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrAPIKeyNotFound
	}
	return nil
}

// Delete deletes an API key
func (r *apiKeyRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&model.APIKeyRecord{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrAPIKeyNotFound
	}
	return nil
}

// Deactivate deactivates an API key
func (r *apiKeyRepository) Deactivate(ctx context.Context, key string) error {
	result := r.db.WithContext(ctx).
		Model(&model.APIKeyRecord{}).
		Where("key = ?", key).
		Update("active", false)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrAPIKeyNotFound
	}
	return nil
}

// List lists all API keys for a tenant
func (r *apiKeyRepository) List(ctx context.Context, tenantID string) ([]model.APIKeyRecord, error) {
	var keys []model.APIKeyRecord
	query := r.db.WithContext(ctx).Model(&model.APIKeyRecord{})

	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}

	result := query.Order("created_at DESC").Find(&keys)
	if result.Error != nil {
		return nil, result.Error
	}

	return keys, nil
}
