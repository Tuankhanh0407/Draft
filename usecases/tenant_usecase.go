// Import appropriate package.
package usecases

// Import necessary libraries.
import (
    "errors"
    "letuan.com/code_demo_backend/domain"
)

// tenantUseCase implements the domain.TenantUseCase interface.
type tenantUseCase struct {
    tenantRepo domain.TenantRepository
}

// NewTenantUseCase creates a new instance of TenantUseCase.
func NewTenantUseCase(repo domain.TenantRepository) domain.TenantUseCase {
    return &tenantUseCase{
        tenantRepo: repo,
    }
}

// CreateTenant validates and creates a new tenant.
func (u *tenantUseCase) CreateTenant(tenant *domain.Tenant) error {
    // Prevent duplicate IDs if manually provided.
    existingTenant, _ := u.tenantRepo.GetByID(tenant.ID)
    if existingTenant != nil {
        return errors.New("Tenant already exists")
    }
    return u.tenantRepo.Create(tenant)
}

// GetAllTenants retrieves a list of all active tenants.
func (u *tenantUseCase) GetAllTenants() ([]domain.Tenant, error) {
    return u.tenantRepo.GetAll()
}

// GetTenantByID retrieves specific tenant details using its ID.
func (u *tenantUseCase) GetTenantByID(id uint) (*domain.Tenant, error) {
    return u.tenantRepo.GetByID(id)
}

// UpdateTenant entirely replaces the data of an existing tenant.
func (u *tenantUseCase) UpdateTenant(id uint, tenant *domain.Tenant) error {
    existingTenant, err := u.tenantRepo.GetByID(id)
    if err != nil {
        return errors.New("Tenant not found")
    }
    // Secure immutable fields.
    tenant.ID = existingTenant.ID
    tenant.CreatedAt = existingTenant.CreatedAt
    return u.tenantRepo.Update(tenant)
}

// PatchTenant applies partial modifications to a tenant safely.
func (u *tenantUseCase) PatchTenant(id uint, updates map[string]interface{}) error {
    _, err := u.tenantRepo.GetByID(id)
    if err != nil {
        return errors.New("Tenant not found")
    }
    // Prevent patching critical fields.
    delete(updates, "tenant_id")
    delete(updates, "id")
    delete(updates, "create_at")
    return u.tenantRepo.Patch(id, updates)
}

// DeleteTenant performs a soft deletion on a tenant.
func (u *tenantUseCase) DeleteTenant(id uint) error {
    _, err := u.tenantRepo.GetByID(id)
    if err != nil {
        return errors.New("Tenant not found")
    }
    return u.tenantRepo.Delete(id)
}