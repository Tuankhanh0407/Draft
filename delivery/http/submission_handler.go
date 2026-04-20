// Import appropriate package.
package http

// Import necessary libraries.
// import (
// 	"github.com/casbin/casbin/v2"
// 	"github.com/gofiber/fiber/v2"
// 	"github.com/go-playground/validator/v10"
// 	"letuan.com/code_demo_backend/delivery/http/middleware"
// 	"letuan.com/code_demo_backend/domain"
// 	"strconv"
// )

// // SubmissionHandler manages exam submissions, drafts, history, and leaderboards.
// type SubmissionHandler struct {
// 	SubmissionUC	domain.SubmissionUseCase
// 	Validate		*validator.Validate
// }

// // NewSubmissionHandler sets up submission-related endpoints protected by Casbin.
// func NewSubmissionHandler(app *fiber.App, uc domain.SubmissionUseCase, val *validator.Validate, enforcer *casbin.Enforcer) {
// 	handler := &SubmissionHandler{
// 		SubmissionUC:	uc,
// 		Validate:		val,
// 	}
// 	api := app.Group("/api/v1/submissions", middleware.Protected(), middleware.RoleBasedAuth(enforcer))
// 	// Draft endpoints for auto-saving progress.
// 	api.Post("/draft", handler.SaveDraft)
// 	api.Get("/draft/:exam_id", handler.GetDraft)
// 	// Core submission endpoints.
// 	api.Post("/", handler.SubmitExam)
// 	api.Post("/regrade/:exam_id", handler.RegradeSubmissions)
// 	// Leaderboard endpoint.
// 	api.Get("/leaderboard/:exam_id", handler.GetLeaderboard)
// 	// History endpoints for students.
// 	userApi := app.Group("/api/v1/users/me/submissions", middleware.Protected(), middleware.RoleBasedAuth(enforcer))
// 	userApi.Get("/", handler.GetMySubmissions)
// 	userApi.Get("/:id", handler.GetSubmissionDetail)
// }

// // SubmitExam grades the user's answers and records the submission.
// // @Summary Submit an exam
// // @Description Evaluate student's answers and save the submission record.
// // @Tags Submissions
// // @Accept json
// // @Produce json
// // @Security BearerAuth
// // @Param request body domain.SubmitRequest true "Submission payload"
// // @Success 200 {object} map[string]interface{} "Submission evaluated successfully"
// // @Failure 400 {object} map[string]interface{} "Invalid request body or exam timing"
// // @Failure 403 {object} map[string]interface{} "Maximum attempts reached"
// // @Failure 500 {object} map[string]interface{} "Internal server error"
// // @Router /api/v1/submissions [post]
// func (h *SubmissionHandler) SubmitExam(c *fiber.Ctx) error {
// 	var req domain.SubmitRequest
// 	if err := c.BodyParser(&req); err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
// 	}
// 	// Attach identity from JWT to prevent tampering.
// 	req.TenantID = c.Locals("tenant_id").(uint)
// 	req.UserID = c.Locals("user_id").(uint)
// 	if err := h.Validate.Struct(req); err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
// 	}
// 	result, err := h.SubmissionUC.EvaluateAndSave(&req)
// 	if err != nil {
// 		// Differentiate between logic errors (400/403) and server errors (500).
// 		if err.Error() == "Maximum attempts reached for this exam" {
// 			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
// 		}
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
// 	}
// 	return c.Status(fiber.StatusOK).JSON(fiber.Map{
// 		"message":	"Submission evaluated successfully",
// 		"data":		result,
// 	})
// }

// // RegradeSubmissions initiates a background task to recalculate all scores for an exam.
// // @Summary Regrade an exam
// // @Description Recalculate scores for all submissions of a specific exam.
// // @Tags Submissions
// // @Accept json
// // @Produce json
// // @Security BearerAuth
// // @Param exam_id path int true "Exam ID"
// // @Success 202 {object} map[string]interface{} "Regrading progress started"
// // @Failure 400 {object} map[string]interface{} "Invalid exam ID format"
// // @Router /api/v1/submissions/regrade/{exam_id} [post]
// func (h *SubmissionHandler) RegradeSubmissions(c *fiber.Ctx) error {
// 	examID, err := strconv.Atoi(c.Params("exam_id"))
// 	if err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid exam ID format"})
// 	}
// 	tenantID := c.Locals("tenant_id").(uint)
// 	if err := h.SubmissionUC.RegradeExamSubmissions(uint(examID), tenantID); err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
// 	}
// 	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
// 		"message": "Regrading process has been started in the background.",
// 	})
// }

// // GetMySubmissions fetches a high-level overview of the student's submission history.
// // @Summary Get user's submissions
// // @Description Retrieve the history of exams taken by the current user.
// // @Tags Submissions
// // @Accept json
// // @Produce json
// // @Security BearerAuth
// // @Success 200 {object} map[string]interface{} "User submission history"
// // @Failure 500 {object} map[string]interface{} "Internal server error"
// // @Router /api/v1/users/me/submissions [get]
// func (h *SubmissionHandler) GetMySubmissions(c *fiber.Ctx) error {
// 	userID := c.Locals("user_id").(uint)
// 	tenantID := c.Locals("tenant_id").(uint)
// 	result, err := h.SubmissionUC.GetMySubmissions(userID, tenantID)
// 	if err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
// 	}
// 	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": result})
// }

// // GetSubmissionDetail provides an in-depth review of a submission, including explanations.
// // @Summary Get submission details
// // @Description Review detailed answers, scores, and proctoring logs for a submission.
// // @Tags Submissions
// // @Accept json
// // @Produce json
// // @Security BearerAuth
// // @Param id path int true "Submission ID"
// // @Success 200 {object} map[string]interface{} "Submission details"
// // @Failure 400 {object} map[string]interface{} "Invalid submission ID"
// // @Failure 403 {object} map[string]interface{} "Forbidden access"
// // @Router /api/v1/users/me/submissions/{id} [get]
// func (h *SubmissionHandler) GetSubmissionDetail(c *fiber.Ctx) error {
// 	subID, err := strconv.Atoi(c.Params("id"))
// 	if err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid submission ID"})
// 	}
// 	userID := c.Locals("user_id").(uint)
// 	tenantID := c.Locals("tenant_id").(uint)
// 	result, err := h.SubmissionUC.GetSubmissionDetail(uint(subID), userID, tenantID)
// 	if err != nil {
// 		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
// 	}
// 	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": result})
// }

// // SaveDraft stores the student's current progress into a temporary Redis cache.
// // @Summary Save answer draft
// // @Description Temporarily save exam answers to Redis cache.
// // @Tags Submissions
// // @Accept json
// // @Produce json
// // @Security BearerAuth
// // @Param request body domain.SubmitRequest true "Draft payload"
// // @Success 200 {object} map[string]interface{} "Draft saved successfully"
// // @Failure 400 {object} map[string]interface{} "Invalid request body"
// // @Failure 500 {object} map[string]interface{} "Internal server error"
// // @Router /api/v1/submissions/draft [post]
// func (h *SubmissionHandler) SaveDraft(c *fiber.Ctx) error {
// 	var req domain.SubmitRequest
// 	if err := c.BodyParser(&req); err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
// 	}
// 	req.TenantID = c.Locals("tenant_id").(uint)
// 	req.UserID = c.Locals("users_id").(uint)
// 	if err := h.SubmissionUC.SaveDraft(&req); err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
// 	}
// 	return c.Status(fiber.StatusOK).JSON(fiber.Map{
// 		"message": "Draft saved successfully",
// 	})
// }

// // GetDraft pulls the most recent auto-saved answers for the student.
// // @Summary Get answer draft
// // @Description Retrieve auto-saved answers from Redis cache.
// // @Tags Submissions
// // @Accept json
// // @Produce json
// // @Security BearerAuth
// // @Param exam_id path int true "Exam ID"
// // @Success 200 {object} map[string]interface{} "Draft retrieved"
// // @Failure 400 {object} map[string]interface{} "Invalid exam ID format"
// // @Failure 500 {object} map[string]interface{} "Internal server error"
// // @Router /api/v1/submissions/draft/{exam_id} [get]
// func (h *SubmissionHandler) GetDraft(c *fiber.Ctx) error {
// 	examID, err := strconv.Atoi(c.Params("exam_id"))
// 	if err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid exam ID format"})
// 	}
// 	tenantID := c.Locals("tenant_id").(uint)
// 	userID := c.Locals("user_id").(uint)
// 	answers, err := h.SubmissionUC.GetDraft(uint(examID), userID, tenantID)
// 	if err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
// 	}
// 	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": answers})
// }

// // GetLeaderboard fetches the top-ranking students for a specific exam.
// // @Summary Get exam leaderboard
// // @Description Fetch top scores for a specific exam.
// // @Tags Submissions
// // @Accept json
// // @Produce json
// // @Security BearerAuth
// // @Param exam_id path int true "Exam ID"
// // @Param limit query int false "Number of top results" default(10)
// // @Success 200 {object} map[string]interface{} "Leaderboard data"
// // @Failure 400 {object} map[string]interface{} "Invalid exam ID format"
// // @Failure 404 {object} map[string]interface{} "Exam not found"
// // @Router /api/v1/submissions/leaderboard/{exam_id} [get]
// func (h *SubmissionHandler) GetLeaderboard(c *fiber.Ctx) error {
// 	examID, err := strconv.Atoi(c.Params("exam_id"))
// 	if err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid exam ID format"})
// 	}
// 	tenantID := c.Locals("tenant_id").(uint)
// 	limit := c.QueryInt("limit", 10) // Default to top 10.
// 	leaderboard, err := h.SubmissionUC.GetLeaderboard(uint(examID), tenantID, limit)
// 	if err != nil {
// 		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
// 	}
// 	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": leaderboard})
// }