// Import appropriate package.
package http

// Import necessary libraries.
import (
	"strconv"
	"github.com/casbin/casbin/v2"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"letuan.com/code_demo_backend/delivery/http/middleware"
	"letuan.com/code_demo_backend/domain"
)

// TenantHandler manages HTTP requests related to tenants.
type TenantHandler struct {
	TenantUC	domain.TenantUseCase
	Validate	*validator.Validate
}

// NewTenantHandler initializes routing for tenant operations.
// It is protected by JWT and Casbin RBAC.
func NewTenantHandler(app *fiber.App, uc domain.TenantUseCase, val *validator.Validate, enforcer *casbin.Enforcer) {
	handler := &TenantHandler{
		TenantUC:	uc,
		Validate:	val,
	}
	// Only SuperAdmin can manage tenants (handled by Casbin).
	api := app.Group("/api/v1/tenants", middleware.Protected(), middleware.RoleBasedAuth(enforcer))
	api.Post("/", handler.CreateTenant)
	api.Get("/", handler.GetAllTenants)
	api.Get("/:id", handler.GetTenant)
	api.Put("/:id", handler.UpdateTenant)
	api.Patch("/:id", handler.PatchTenant)
	api.Delete("/:id", handler.DeleteTenant)
}

// CreateTenant creates a new organizational tenant.
// @Summary Create a tenant
// @Description Register a new organizational/school in the system (SuperAdmin only).
// @Tags Tenants
// @Accept json
// @Product json
// @Security BearerAuth
// @Param request body domain.Tenant true "Tenant payload"
// @Success 201 {object} map[string]interface{} "Tenant created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 409 {object} map[string]interface{} "Conflict in tenant creation"
// @Router /api/v1/tenants [post]
func (h *TenantHandler) CreateTenant(c *fiber.Ctx) error {
	var tenant domain.Tenant
	if err := c.BodyParser(&tenant); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if err := h.Validate.Struct(tenant); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if err := h.TenantUC.CreateTenant(&tenant); err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":	"Tenant created successfully",
		"data":		tenant,
	})
}

// GetAllTenants retrieves a list of all active tenants.
// @Summary Get all tenants
// @Description Fetch a list of all organizations/
// @Tags Tenants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "List of tenants"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/tenants [get]
func (h *TenantHandler) GetAllTenants(c *fiber.Ctx) error {
	tenants, err := h.TenantUC.GetAllTenants()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": tenants})
}

// GetTenant fetches a specific tenant by its ID.
// @Summary Get tenant details
// @Description Fetch information of a specific tenant.
// @Tags Tenants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Tenant ID"
// @Success 200 {object} map[string]interface{} "Tenant details"
// @Failure 400 {object} map[string]interface{} "Invalid tenant ID format"
// @Failure 404 {object} map[string]interface{} "Tenant not found"
// @Router /api/v1/tenants/{id} [get]
func (h *TenantHandler) GetTenant(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid tenant ID format"})
	}
	tenant, err := h.TenantUC.GetTenantByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Tenant not found"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": tenant})
}

// UpdateTenant completely replaces an existing tenant's data.
// @Summary Update a tenant
// @Description Completely replace tenant's data.
// @Tags Tenants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Tenant ID"
// @Param request body domain.Tenant true "Tenant payload"
// @Success 200 {object} map[string]interface{} "Tenant updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 404 {object} map[string]interface{} "Tenant not found"
// @Router /api/v1/tenants/{id} [put]
func (h *TenantHandler) UpdateTenant(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid tenant ID format"})
	}
	var tenant domain.Tenant
	if err := c.BodyParser(&tenant); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if err := h.Validate.Struct(tenant); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}
	if err := h.TenantUC.UpdateTenant(uint(id), &tenant); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":	"Tenant updated successfully",
		"data":		tenant,
	})
}

// PatchTenant partially updates a tenant based on provided fields.
// @Summary Patch a tenant
// @Description Partially update specific fields of a tenant.
// @Tags Tenants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Tenant ID"
// @Param request body map[string]interface{} true "Fields to update"
// @Success 200 {object} map[string]interface{} "Tenant patched successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 404 {object} map[string]interface{} "Tenant not found"
// @Router /api/v1/tenants/{id} [patch]
func (h *TenantHandler) PatchTenant(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid tenant ID format"})
	}
	var updates map[string]interface{}
	if err := c.BodyParser(&updates); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if err := h.TenantUC.PatchTenant(uint(id), updates); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}
	updatedTenant, _ := h.TenantUC.GetTenantByID(uint(id))
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":	"Tenant patched successfully",
		"data":		updatedTenant,
	})
}

// DeleteTenant performs a soft delete on a tenant.
// @Summary Delete a tenant
// @Description Perform a soft delete on an organization.
// @Tags Tenants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Tenant ID"
// @Success 200 {object} map[string]interface{} "Tenant deleted successfully"
// @Failure 400 {object} map[string]interface{} "Invalid tenant ID format"
// @Failure 404 {object} map[string]interface{} "Tenant not found"
// @Router /api/v1/tenants/{id} [delete]
func (h *TenantHandler) DeleteTenant(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid tenant ID format"})
	}
	if err := h.TenantUC.DeleteTenant(uint(id)); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Tenant deleted successfully"})
}