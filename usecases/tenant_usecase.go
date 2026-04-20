// Import appropriate package.
package usecases

// Import necessary libraries.
import (
	"context"
	"letuan.com/code_demo_backend/domain"
)

// tenantUsecase implements the business logic for tenant operations.
type tenantUsecase struct {
	repo domain.TenantRepository
}

// NewTenantUsecase initializes the business logic layer for tenants.
func NewTenantUsecase(repo domain.TenantRepository) domain.TenantUsecase {
	return &tenantUsecase{repo}
}

// GetAll retrieves a list of all available tenants.
func (u *tenantUsecase) GetAll(ctx context.Context) ([]domain.Tenant, error) {
	return u.repo.Fetch(ctx)
}

// GetByID fetches a specific tenant by its unique identifier.
func (u *tenantUsecase) GetByID(ctx context.Context, id uint) (domain.Tenant, error) {
	return u.repo.GetByID(ctx, id)
}

// Create validates and registers a new tenant in the system.
func (u *tenantUsecase) Create(ctx context.Context, tenant *domain.Tenant) error {
	// Business logic: Check for potential conflicts or validation here.
	return u.repo.Create(ctx, tenant)
}

// Update modifies an existing tenant's details while preserving protected fields like CreatedAt.
func (u *tenantUsecase) Update(ctx context.Context, id uint, tenant *domain.Tenant) error {
	// 1. Fetch the existing record to ensure it exists and to retain its current state.
	existing, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	// 2. Apply safe updates: Only modify fields that are allowed to be changed.
	existing.Name = tenant.Name
	existing.Code = tenant.Code
	// 3. Persist the safely merged record back to the database.
	return u.repo.Update(ctx, &existing)
}

// Delete removes a tenant from the system based on its ID.
func (u *tenantUsecase) Delete(ctx context.Context, id uint) error {
	return u.repo.Delete(ctx, id)
}