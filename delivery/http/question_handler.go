// Import appropriate package.
package http

// Import necessary libraries.
import (
	"fmt"
	"github.com/casbin/casbin/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/go-playground/validator/v10"
	"letuan.com/code_demo_backend/delivery/http/middleware"
	"letuan.com/code_demo_backend/domain"
	"strconv"
)

// QuestionHandler manages API operations for questions.
type QuestionHandler struct {
	QuestionUC	domain.QuestionUseCase
	Validate	*validator.Validate
}

// NewQuestionHandler configures routing for question endpoints protected by Casbin.
func NewQuestionHandler(app *fiber.App, uc domain.QuestionUseCase, val *validator.Validate, enforcer *casbin.Enforcer) {
	handler := &QuestionHandler{
		QuestionUC: uc,
		Validate:	val,
	}
	api := app.Group("/api/v1/questions", middleware.Protected(), middleware.RoleBasedAuth(enforcer))
	api.Post("/bulk", handler.CreateQuestionsBulk)
	api.Post("/", handler.CreateQuestion)
	api.Get("/", handler.ListQuestions)
	api.Get("/:id", handler.GetQuestion)
}

// CreateQuestionsBulk processes mass-import of questions.
// @Summary Bulk create questions
// @Description Import multiple questions simultaneously.
// @Tags Questions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body []domain.Question true "Array of questions"
// @Success 201 {object} map[string]interface{} "Questions imported successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/questions/bulk [post]
func (h *QuestionHandler) CreateQuestionsBulk(c *fiber.Ctx) error {
	var questions []domain.Question
	if err := c.BodyParser(&questions); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body. Expected an array of questions"})
	}
	tenantID, ok := c.Locals("tenant_id").(uint)
	if !ok || tenantID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized access"})
	}
	// Validate each question and force tenant isolation.
	for i := range questions {
		questions[i].TenantID = tenantID
		if err := h.Validate.Struct(&questions[i]); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("Validation failed for question at index %d: %v", i, err),
			})
		}
	}
	if err := h.QuestionUC.CreateQuestionsBulk(questions); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": fmt.Sprintf("Successfully imported %d questions", len(questions)),
	})
}

// CreateQuestion adds a single question.
// @Summary Create a single question
// @Description Add a new question to the bank.
// @Tags Questions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.Question true "Question payload"
// @Success 201 {object} map[string]interface{} "Question created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 409 {object} map[string]interface{} "Conflict in creation"
// @Router /api/v1/questions [post]
func (h *QuestionHandler) CreateQuestion(c *fiber.Ctx) error {
	var question domain.Question
	if err := c.BodyParser(&question); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	question.TenantID = c.Locals("tenant_id").(uint)
	if err := h.Validate.Struct(question); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if err := h.QuestionUC.CreateQuestion(&question); err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":	"Question created successfully",
		"data":		question,
	})
}

// GetQuestion fetches a specific question by ID, stripped of sensitive answers.
// @Summary Get question details
// @Description Retrieve a specific question (correct answers are stripped for clients).
// @Tags Questions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Question ID"
// @Success 200 {object} map[string]interface{} "Question details"
// @Failure 400 {object} map[string]interface{} "Invalid question ID format"
// @Failure 404 {object} map[string]interface{} "Question not found"
// @Router /api/v1/questions/{id} [get]
func (h *QuestionHandler) GetQuestion(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid question ID format"})
	}
	tenantID := c.Locals("tenant_id").(uint)
	question, err := h.QuestionUC.GetQuestionForClient(uint(id), tenantID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": question})
}

// ListQuestions provides a paginated list of questions, filterable by tags.
// @Summary List all questions
// @Description Fetch a paginated list of questions, with optional tag filtering.
// @Tags Questions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param tags query string false "Filter by tags (comma separated)"
// @Success 200 {object} map[string]interface{} "List of questions"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/questions [get]
func (h *QuestionHandler) ListQuestions(c *fiber.Ctx) error {
	tenantID := c.Locals("tenant_id").(uint)
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	tag := c.Query("tags", "")
	result, err := h.QuestionUC.ListQuestions(tenantID, page, limit, tag)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch questions"})
	}
	return c.Status(fiber.StatusOK).JSON(result)
}