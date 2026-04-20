// Import appropriate package.
package http

// Import necessary libraries.
import (
	"github.com/gofiber/fiber/v2"
	"letuan.com/code_demo_backend/domain"
	"strconv"
)

// TenantHandler manages incoming API requests and HTTP responses for tenants.
type TenantHandler struct {
	Usecase domain.TenantUsecase
}

// NewTenantHandler registers routes for tenant management.
func NewTenantHandler(app *fiber.App, us domain.TenantUsecase) {
	handler := &TenantHandler{Usecase: us}
	api := app.Group("/api/v1/tenants")

	api.Get("/", handler.GetAll)
	api.Get("/:id", handler.GetByID)
	api.Post("/", handler.Create)
	api.Put("/:id", handler.Update)
	api.Delete("/:id", handler.Delete)
}

// GetAll retrieves a list of all tenants.
// @Summary Get all tenants
// @Description Retrieve a list of all organizations
// @Tags Tenants
// @Produce json
// @Success 200 {array} domain.Tenant "List of tenants"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/tenants [get]
func (h *TenantHandler) GetAll(c *fiber.Ctx) error {
	tenants, err := h.Usecase.GetAll(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(tenants)
}

// GetByID fetches a specific tenant by its ID.
// @Summary Get tenant by ID
// @Description Retrieve a single organization by its unique identifier
// @Tags Tenants
// @Produce json
// @Param id path int true "Tenant ID"
// @Success 200 {object} domain.Tenant "Successfully retrieved tenant"
// @Failure 404 {object} map[string]interface{} "Tenant not found"
// @Router /api/v1/tenants/{id} [get]
func (h *TenantHandler) GetByID(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	tenant, err := h.Usecase.GetByID(c.Context(), uint(id))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Tenant not found"})
	}
	return c.JSON(tenant)
}

// Create registers a new tenant in the system.
// @Summary Create a new tenant
// @Description Add a new organization to the system
// @Tags Tenants
// @Accept json
// @Produce json
// @Param tenant body domain.Tenant true "Tenant info"
// @Success 201 {object} domain.Tenant "Successfully created tenant"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 409 {object} map[string]interface{} "Tenant already exists or database error"
// @Router /api/v1/tenants [post]
func (h *TenantHandler) Create(c *fiber.Ctx) error {
	tenant := new(domain.Tenant)
	if err := c.BodyParser(tenant); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}
	if err := h.Usecase.Create(c.Context(), tenant); err != nil {
		return c.Status(409).JSON(fiber.Map{"error": "Tenant already exists or database error"})
	}
	return c.Status(201).JSON(tenant)
}

// Update modifies an existing tenant's details.
// @Summary Update entire tenant
// @Description Update a tenant by ID with the provided JSON payload and return the updated record.
// @Tags Tenants
// @Accept json
// @Produce json
// @Param id path int true "Tenant ID"
// @Param tenant body domain.Tenant true "Updated tenant information"
// @Success 200 {object} domain.Tenant "Successfully updated tenant"
// @Failure 400 {object} map[string]interface{} "Bad request - Invalid payload or update failed"
// @Router /api/v1/tenants/{id} [put]
func (h *TenantHandler) Update(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	tenant := new(domain.Tenant)
	if err := c.BodyParser(tenant); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if err := h.Usecase.Update(c.Context(), uint(id), tenant); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	updatedTenant, _ := h.Usecase.GetByID(c.Context(), uint(id))
	return c.JSON(updatedTenant)
}

// Delete removes a tenant from the system.
// @Summary Soft delete a tenant
// @Description Delete a tenant by ID
// @Tags Tenants
// @Param id path int true "Tenant ID"
// @Success 204 "No content - Successfully deleted"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/tenants/{id} [delete]
func (h *TenantHandler) Delete(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	if err := h.Usecase.Delete(c.Context(), uint(id)); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(204)
}