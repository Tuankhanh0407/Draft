// Import appropriate package.
package http

// Import necessary libraries.
import (
	"github.com/gofiber/fiber/v2"
	"letuan.com/code_demo_backend/delivery/http/middleware"
	"letuan.com/code_demo_backend/domain"
	"strconv"
)

// QuestionHandler manages incoming API requests and HTTP responses for questions.
type QuestionHandler struct {
	Usecase domain.QuestionUsecase
}

// NewQuestionHandler registers protected routes for question management.
func NewQuestionHandler(app *fiber.App, us domain.QuestionUsecase) {
	handler := &QuestionHandler{Usecase: us}
	// Apply authentication and role-based access control (RBAC) middleware to this group.
	api := app.Group("/api/v1/questions", middleware.AuthProtected(), middleware.RoleProtected("teacher", "admin"))
	
	api.Get("/", handler.GetAll)
	api.Get("/:id", handler.GetByID)
	api.Post("/", handler.Create)
	api.Post("/bulk", handler.CreateBulk)
	api.Put("/:id", handler.Update)
	api.Delete("/:id", handler.Delete)
}

// GetAll retrieves a paginated list of questions.
// @Summary List questions
// @Description Fetch questions belonging to the tenant, exclude correct answers depending on logic
// @Tags Questions
// @Security BearerAuth
// @Produce json
// @Param type query string false "Filter by Type"
// @Param tag query string false "Filter by Tag"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {array} domain.QuestionResponse "Successfully retrieved list"
// @Failure 401 {object} map[string]string "Unauthorized: Missing or invalid token"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/questions [get]
func (h *QuestionHandler) GetAll(c *fiber.Ctx) error {
	tenantID := uint(c.Locals("tenant_id").(float64)) // JWT claims are parsed as float64.
	qType := c.Query("type")
	tag := c.Query("tag")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	res, err := h.Usecase.GetAll(c.Context(), tenantID, qType, tag, page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(res)
}

// Create adds a new question to the system.
// @Summary Create a question
// @Description Add a new abstract syntax tree (AST) question (teacher/admin only)
// @Tags Questions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body domain.QuestionRequest true "Question payload"
// @Success 201 {object} domain.QuestionResponse "Successfully created"
// @Failure 400 {object} map[string]string "Bad request: Invalid payload"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden: Insufficient role permissions"
// @Router /api/v1/questions [post]
func (h *QuestionHandler) Create(c *fiber.Ctx) error {
	tenantID := uint(c.Locals("tenant_id").(float64))
	var req domain.QuestionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}
	res, err := h.Usecase.Create(c.Context(), tenantID, &req)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(res)
}

// CreateBulk inserts multiple questions at once.
// @Summary Bulk insert questions
// @Description Upload an array of questions simultaneously
// @Tags Questions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body []domain.QuestionRequest true "Array of questions"
// @Success 201 {string} string "Created"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Router /api/v1/questions/bulk [post]
func (h *QuestionHandler) CreateBulk(c *fiber.Ctx) error {
	tenantID := uint(c.Locals("tenant_id").(float64))
	var reqs []domain.QuestionRequest
	if err := c.BodyParser(&reqs); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if err := h.Usecase.CreateBulk(c.Context(), tenantID, reqs); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(201)
}

// GetByID fetches a specific question by its ID.
// @Summary Get question by ID
// @Description Retrieve a single question's details
// @Tags Questions
// @Security BearerAuth
// @Produce json
// @Param id path int true "Question ID"
// @Success 200 {object} domain.QuestionResponse "Successfully retrieved"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Not found: Question does not exist"
// @Router /api/v1/questions/{id} [get]
func (h *QuestionHandler) GetByID(c *fiber.Ctx) error {
	tenantID := uint(c.Locals("tenant_id").(float64))
	id, _ := strconv.Atoi(c.Params("id"))
	res, err := h.Usecase.GetByID(c.Context(), tenantID, uint(id))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(res)
}

// Update modifies an existing question.
// @Summary Update question
// @Description Modify an existing question's details by ID
// @Tags Questions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Question ID"
// @Param request body domain.QuestionRequest true "Question payload"
// @Success 200 {object} domain.QuestionResponse "Successfully updated"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Router /api/v1/questions/{id} [put]
func (h *QuestionHandler) Update(c *fiber.Ctx) error {
	tenantID := uint(c.Locals("tenant_id").(float64))
	id, _ := strconv.Atoi(c.Params("id"))
	var req domain.QuestionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}
	res, err := h.Usecase.Update(c.Context(), tenantID, uint(id), &req)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(res)
}

// Delete removes a question from the system.
// @Summary Delete question
// @Description Remove a question by ID
// @Tags Questions
// @Security BearerAuth
// @Produce json
// @Param id path int true "Question ID"
// @Success 204 "No content: Successfully deleted"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/questions/{id} [delete]
func (h *QuestionHandler) Delete(c *fiber.Ctx) error {
	tenantID := uint(c.Locals("tenant_id").(float64))
	id, _ := strconv.Atoi(c.Params("id"))
	if err := h.Usecase.Delete(c.Context(), tenantID, uint(id)); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(204)
}