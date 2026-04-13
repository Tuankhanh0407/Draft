// Import appropriate package.
package http

// Import necessary libraries.
import (
	"fmt"
	"strconv"
	"github.com/casbin/casbin/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/go-playground/validator/v10"
	"letuan.com/code_demo_backend/delivery/http/middleware"
	"letuan.com/code_demo_backend/domain"
)

// ExamHandler manages API routes related to exams.
type ExamHandler struct {
	ExamUC			domain.ExamUseCase
	SubmissionUC	domain.SubmissionUseCase
	Validate		*validator.Validate
}

// NewExamHandler connects exam endpoints to the router protected by Casbin.
func NewExamHandler(app *fiber.App, euc domain.ExamUseCase, suc domain.SubmissionUseCase, val *validator.Validate, enforcer *casbin.Enforcer) {
	handler := &ExamHandler{
		ExamUC:			euc,
		SubmissionUC:	suc,
		Validate:		val,
	}
	api := app.Group("/api/v1/exams", middleware.Protected(), middleware.RoleBasedAuth(enforcer))
	api.Post("/", handler.CreateExam)
	api.Get("/:id", handler.GetExam)
	api.Get("/:id/analytics", handler.GetExamAnalytics)
	api.Post("/:id/cheat-log", handler.LogCheatEvent)
	api.Get("/:id/export", handler.ExportExamLeaderboard)
}

// ExportExamLeaderboard generates and downloads an Excel file of the exam leaderboard.
// @Summary Export exam leaderboard to Excel
// @Description Download the leaderboard of a specific exam as an Excel file (.xlsx) for reporting.
// @Tags Exams
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Security BearerAuth
// @Param id path int true "Exam ID"
// @Success 200 {file} file "Excel file"
// @Failure 400 {object} map[string]interface{} "Invalid exam ID format"
// @Failure 401 {object} map[string]interface{} "Unauthorized access"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/exams/{id}/export [get]
func (h *ExamHandler) ExportExamLeaderboard(c *fiber.Ctx) error {
	examID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid exam ID format"})
	}
	tenantID, ok := c.Locals("tenant_id").(uint)
	if !ok || tenantID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized access"})
	}
	// 1. Generate the Excel buffer.
	buf, err := h.SubmissionUC.ExportLeaderboardToExcel(uint(examID), tenantID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	// 2. Set headers to force file download in the browser.
	filename := fmt.Sprintf("report_exam_id_%d_score.xlsx", examID)
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename = %s", filename))
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	// 3. Stream the buffer directly to the client.
	return c.SendStream(buf)
}

// CreateExam validates and creates an exam containing referenced questions.
// @Summary Create a new exam
// @Description Create a new exam with specific question IDs.
// @Tags Exams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.Exam true "Exam payload"
// @Success 201 {object} map[string]interface{} "Exam created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 409 {object} map[string]interface{} "Conflict in creating exam"
// @Router /api/v1/exams [post]
func (h *ExamHandler) CreateExam(c *fiber.Ctx) error {
	var exam domain.Exam
	if err := c.BodyParser(&exam); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}
	exam.TenantID = c.Locals("tenant_id").(uint)
	if err := h.Validate.Struct(exam); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if err := h.ExamUC.CreateExam(&exam); err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Exam created",
		"data": 	exam,
	})
}

// GetExam starts an exam session and returns the structure for the client.
// @Summary Get exam details
// @Description Fetch exam structure and details for a student to take.
// @Tags Exams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Exam ID"
// @Success 200 {object} map[string]interface{} "Exam details"
// @Failure 400 {object} map[string]interface{} "Invalid exam ID format"
// @Failure 404 {object} map[string]interface{} "Exam not found"
// @Router /api/v1/exams/{id} [get]
func (h *ExamHandler) GetExam(c *fiber.Ctx) error {
	examID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid exam ID format"})
	}
	tenantID := c.Locals("tenant_id").(uint)
	userID := c.Locals("user_id").(uint)
	exam, err := h.ExamUC.GetExamForClient(uint(examID), tenantID, userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": exam})
}

// GetExamAnalytics provides performance statistics for an exam.
// @Summary Get exam analytics
// @Description Fetch performance statistics for an exam.
// @Tags Exams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Exam ID"
// @Success 200 {object} map[string]interface{} "Exam analytics"
// @Failure 400 {object} map[string]interface{} "Invalid exam ID format"
// @Failure 404 {object} map[string]interface{} "Exam not found"
// @Router /api/v1/exams/{id}/analytics [get]
func (h *ExamHandler) GetExamAnalytics(c *fiber.Ctx) error {
	examID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid exam ID format"})
	}
	tenantID := c.Locals("tenant_id").(uint)
	analytics, err := h.ExamUC.GetExamAnalytics(uint(examID), tenantID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": analytics})
}

// LogCheatEvent handles incoming alerts when a student violates proctoring rules.
// @Summary Log a cheat event
// @Description Record a proctoring alert during an active exam.
// @Tags Exams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Exam ID"
// @Param request body domain.CheatLogRequest true "Cheat log payload"
// @Success 201 {object} map[string]interface{} "Cheat event recorded successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/exams/{id}/cheat-log [post]
func (h *ExamHandler) LogCheatEvent(c *fiber.Ctx) error {
	examID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid exam ID format"})
	}
	var req domain.CheatLogRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	// Securely bind identities from JWT.
	req.ExamID = uint(examID)
	req.TenantID = c.Locals("tenant_id").(uint)
	req.UserID = c.Locals("user_id").(uint)
	if err := h.Validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if err := h.ExamUC.LogCheatEvent(&req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Cheat event recorded successfully",
	})
}