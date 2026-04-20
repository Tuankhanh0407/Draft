// Import appropriate package.
package domain

// Import necessary libraries.
import (
	"context"
	"gorm.io/gorm"
	"time"
)

// Tenant represents an organization or school using the system.
// It includes soft delete support via gorm.DeletedAt.
type Tenant struct {
	ID			uint			`json:"id" gorm:"primaryKey;autoIncrement"`
	Name		string			`json:"name" gorm:"unique;not null;size:255"`
	Code		string			`json:"code" gorm:"unique;not null;size:50"`
	CreatedAt	time.Time		`json:"created_at"`
	UpdatedAt	time.Time		`json:"updated_at"`
	DeletedAt	gorm.DeletedAt	`json:"-" gorm:"index"`
}

// TenantRepository defines the contract for tenant data operations.
type TenantRepository interface {
	Fetch(ctx context.Context) ([]Tenant, error)
	GetByID(ctx context.Context, id uint) (Tenant, error)
	Create(ctx context.Context, tenant *Tenant) error
	Update(ctx context.Context, tenant *Tenant) error
	Delete(ctx context.Context, id uint) error
}

// TenantUsecase defines the business logic for tenant management.
type TenantUsecase interface {
	GetAll(ctx context.Context) ([]Tenant, error)
	GetByID(ctx context.Context, id uint) (Tenant, error)
	Create(ctx context.Context, tenant *Tenant) error
	Update(ctx context.Context, id uint, tenant *Tenant) error
	Delete(ctx context.Context, id uint) error
}