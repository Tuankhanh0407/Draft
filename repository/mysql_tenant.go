// Import appropriate package.
package repository

// Import necessary libraries.
import (
	"gorm.io/gorm"
	"letuan.com/code_demo_backend/domain"
)

// mysqlTenantRepository represents a MySQL-backed repository for managing tenant data using GORM.
type mysqlTenantRepository struct {
	db *gorm.DB
}

// NewMySQLTenantRepository creates a new instance of TenantRepository.
func NewMySQLTenantRepository(db *gorm.DB) domain.TenantRepository {
	return &mysqlTenantRepository{db: db}
}

// Create inserts a new tenant into the database.
func (m *mysqlTenantRepository) Create(tenant *domain.Tenant) error {
	return m.db.Create(tenant).Error
}

// GetAll fetches all active tenants from the database.
func (m *mysqlTenantRepository) GetAll() ([]domain.Tenant, error) {
	var tenants []domain.Tenant
	err := m.db.Find(&tenants).Error
	return tenants, err
}

// GetByID fetches a specific tenant by its primary key.
func (m *mysqlTenantRepository) GetByID(id uint) (*domain.Tenant, error) {
	var tenant domain.Tenant
	err := m.db.Where("id = ?", id).First(&tenant).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

// Update completely modifies an existing tenant's record.
func (m *mysqlTenantRepository) Update(tenant *domain.Tenant) error {
	return m.db.Save(tenant).Error
}

// Patch updates only the specific fields of a tenant using a map.
func (m *mysqlTenantRepository) Patch(id uint, updates map[string]interface{}) error {
	// GORM's updates method with a map will only update the specified fields.
	return m.db.Model(&domain.Tenant{}).Where("id = ?", id).Updates(updates).Error
}

// Delete removes a tenant by its ID (performs a soft delete).
func (m *mysqlTenantRepository) Delete(id uint) error {
	return m.db.Where("id = ?", id).Delete(&domain.Tenant{}).Error
}