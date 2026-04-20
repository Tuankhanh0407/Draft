// Import appropriate package.
package repository

// Import necessary libraries.
import (
	"context"
	"gorm.io/gorm"
	"letuan.com/code_demo_backend/domain"
)

// mysqlTenantRepository provides MySQL data access for Tenant entities using GORM.
type mysqlTenantRepository struct {
	db *gorm.DB
}

// NewMysqlTenantRepository creates a new instance of TenantRepository.
func NewMysqlTenantRepository(db *gorm.DB) domain.TenantRepository {
	return &mysqlTenantRepository{db}
}

// Fetch retrieves all non-deleted tenants.
// SQL: SELECT * FROM tenants WHERE deleted_at IS NULL;
func (r *mysqlTenantRepository) Fetch(ctx context.Context) ([]domain.Tenant, error) {
	var tenants []domain.Tenant
	err := r.db.WithContext(ctx).Find(&tenants).Error
	return tenants, err
}

// GetByID retrieves a specific tenant by its ID.
// SQL: SELECT * FROM tenants WHERE id = ? AND deleted_at IS NULL LIMIT 1;
func (r *mysqlTenantRepository) GetByID(ctx context.Context, id uint) (domain.Tenant, error) {
	var tenant domain.Tenant
	err := r.db.WithContext(ctx).First(&tenant, id).Error
	return tenant, err
}

// Create inserts a new tenant record.
// SQL: INSERT INTO tenants (name, code, created_at, updated_at) VALUES (?, ?, ?, ?);
func (r *mysqlTenantRepository) Create(ctx context.Context, tenant *domain.Tenant) error {
	return r.db.WithContext(ctx).Create(tenant).Error
}

// Update saves all fields of the tenant record.
// SQL: UPDATE tenants SET name = ?, code = ?, updated_at = ? WHERE id = ? AND deleted_at IS NULL;
func (r *mysqlTenantRepository) Update(ctx context.Context, tenant *domain.Tenant) error {
	return r.db.WithContext(ctx).Save(tenant).Error
}

// Delete performs a soft delete by updating the deleted_at column.
// SQL: UPDATE tenants SET deleted_at = NOW() WHERE id = ?;
func (r *mysqlTenantRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Tenant{}, id).Error
}