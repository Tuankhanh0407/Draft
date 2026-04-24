// Import appropriate package.
package http

// Import necessary libraries.
import (
	"github.com/gofiber/fiber/v2"
	"letuan.com/code_demo_backend/delivery/http/middleware"
	"letuan.com/code_demo_backend/domain"
	"strconv"
	"strings"
)

// ExamHandler manages incoming API requests and HTTP responses for exams.
type ExamHandler struct {
	Usecase domain.ExamUsecase
}

// NewExamHandler registers routes for exam management.
func NewExamHandler(app *fiber.App, us domain.ExamUsecase) {
	handler := &ExamHandler{Usecase: us}

	// Group requires authentication for all.
	api := app.Group("/api/v1/exams", middleware.AuthProtected())
	// Read routes (allowed for all authenticated roles).
	api.Get("/", handler.GetAll)
	api.Get("/:id", handler.GetByID)

	// Write and analytics routes (strictly teacher/admin).
	teacherAPI := api.Group("/", middleware.RoleProtected("teacher", "admin"))
	teacherAPI.Get("/:id/analytics", handler.GetAnalytics)
	teacherAPI.Post("/", handler.Create)
	teacherAPI.Put("/:id", handler.Update)
	teacherAPI.Delete("/:id", handler.Delete)
}

// handleUsecaseError maps domain errors to appropriate HTTP status codes based on their prefixes.
func handleUsecaseError(c *fiber.Ctx, err error) error {
	errMsg := err.Error()
	if strings.HasPrefix(errMsg, "422") {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": errMsg})
	}
	if strings.HasPrefix(errMsg, "409") {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": errMsg})
	}
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": errMsg})
}

// GetAll retrieves a paginated list of exams for the current tenant.
// @Summary List exams
// @Description Retrieve a list of exams with pagination support
// @Tags Exams
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {array} domain.ExamResponse "List of exams"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/exams [get]
func (h *ExamHandler) GetAll(c *fiber.Ctx) error {
	tenantID := uint(c.Locals("tenant_id").(float64))
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	res, err := h.Usecase.GetAll(c.Context(), tenantID, page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(res)
}

// GetByID fetches detailed information about a specific exam.
// @Summary Get exam details
// @Description Teachers get correct answers, students get cached sanitized versions
// @Tags Exams
// @Security BearerAuth
// @Produce json
// @Param id path int true "Exam ID"
// @Success 200 {object} domain.ExamResponse "Exam details"
// @Failure 404 {object} map[string]string "Exam not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/exams/{id} [get]
func (h *ExamHandler) GetByID(c *fiber.Ctx) error {
	tenantID := uint(c.Locals("tenant_id").(float64))
	role := c.Locals("role").(string)
	id, _ := strconv.Atoi(c.Params("id"))
	res, err := h.Usecase.GetByID(c.Context(), tenantID, uint(id), role)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(res)
}

// Create validates and stores a new exam in the system.
// @Summary Create an exam
// @Description Add a new exam to the system under the current tenant
// @Tags Exams
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body domain.ExamRequest true "Exam payload"
// @Success 201 {object} domain.ExamResponse "Successfully created exam"
// @Failure 400 {object} map[string]string "Bad request: Invalid payload"
// @Failure 409 {object} map[string]string "Conflict: Exam data conflicts"
// @Failure 422 {object} map[string]string "Unprocessable entity: Validation failed"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/exams [post]
func (h *ExamHandler) Create(c *fiber.Ctx) error {
	tenantID := uint(c.Locals("tenant_id").(float64))
	var req domain.ExamRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}
	res, err := h.Usecase.Create(c.Context(), tenantID, &req)
	if err != nil {
		return handleUsecaseError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(res)
}

// Update modifies an existing exam's information.
// @Summary Update an exam
// @Description Update an exam's details by its ID
// @Tags Exams
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Exam ID"
// @Param request body domain.ExamRequest true "Exam payload"
// @Success 200 {object} domain.ExamResponse "Successfully updated exam"
// @Failure 400 {object} map[string]string "Bad request: Invalid payload"
// @Failure 409 {object} map[string]string "Conflict"
// @Failure 422 {object} map[string]string "Unprocessable entity"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/exams/{id} [put]
func (h *ExamHandler) Update(c *fiber.Ctx) error {
	tenantID := uint(c.Locals("tenant_id").(float64))
	id, _ := strconv.Atoi(c.Params("id"))
	var req domain.ExamRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}
	res, err := h.Usecase.Update(c.Context(), tenantID, uint(id), &req)
	if err != nil {
		return handleUsecaseError(c, err)
	}
	return c.JSON(res)
}

// Delete removes an exam from the system based on its ID.
// @Summary Delete an exam
// @Description Delete an exam by ID
// @Tags Exams
// @Security BearerAuth
// @Produce json
// @Param id path int true "Exam ID"
// @Success 204 "No content: Successfully deleted"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/exams/{id} [delete]
func (h *ExamHandler) Delete(c *fiber.Ctx) error {
	tenantID := uint(c.Locals("tenant_id").(float64))
	id, _ := strconv.Atoi(c.Params("id"))
	if err := h.Usecase.Delete(c.Context(), tenantID, uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// GetAnalytics retrieves performance metrics and statistics for a specific exam.
// @Summary Get exam analytics
// @Description Retrieve analytics data for an exam (in example: average score, completion rate).
// @Tags Exams
// @Security BearerAuth
// @Produce json
// @Param id path int true "Exam ID"
// @Success 200 {object} domain.AnalyticsResponse "Analytics data"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/exams/{id}/analytics [get]
func (h *ExamHandler) GetAnalytics(c *fiber.Ctx) error {
	tenantID := uint(c.Locals("tenant_id").(float64))
	id, _ := strconv.Atoi(c.Params("id"))
	res, err := h.Usecase.GetAnalytics(c.Context(), tenantID, uint(id))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(res)
}