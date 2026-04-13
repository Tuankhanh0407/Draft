// Import appropriate package.
package domain

// Import necessary libraries.
import (
	"gorm.io/gorm"
	"time"
)

// Tenant represents an organization or application using the assessment engine (in example: english_app).
type Tenant struct {
	ID			uint			`json:"tenant_id" gorm:"primaryKey;autoIncrement"`
	Name		string			`json:"name" gorm:"type:varchar(100);not null" validate:"required,min=3"`
	Status		string			`json:"status" gorm:"type:varchar(20);default:'active'"`
	CreatedAt	time.Time		`json:"create_at"`
	UpdatedAt	time.Time		`json:"updated_at"`
	DeleteAt	gorm.DeletedAt	`json:"-" gorm:"index"` // Soft delete column, "-" hides it from JSON responses.
}

// TenantRepository defines the database operations for a tenant.
type TenantRepository interface {
	// Create inserts a new tenant into the database.
	Create(tenant *Tenant) error
	// GetAll retrieves a list of all active tenants.
	GetAll() ([]Tenant, error)
	// GetByID fetches a specific tenant by its unique tenant ID.
	GetByID(id uint) (*Tenant, error)
	// Update completely replaces an existing tenant's data.
	Update(tenant *Tenant) error
	// Patch partially updates specific fields of a tenant.
	Patch(id uint, updates map[string]interface{}) error
	// Delete performs a soft delete on a tenant by its ID.
	Delete(id uint) error
}

// TenantUseCase defines the business logic operations for a tenant.
type TenantUseCase interface {
	// CreateTenant validates and creates a new tenant.
	CreateTenant(tenant *Tenant) error
	// GetAllTenants returns all available tenants in the system.
	GetAllTenants() ([]Tenant, error)
	// GetTenantByID retrieves tenant details by its ID.
	GetTenantByID(id uint) (*Tenant, error)
	// UpdateTenant fully updates a tenant's information.
	UpdateTenant(id uint, tenant *Tenant) error
	// PatchTenant applies partial modifications to a tenant.
	PatchTenant(id uint, updates map[string]interface{}) error
	// DeleteTenant safely removes a tenant from active usage.
	DeleteTenant(id uint) error
}