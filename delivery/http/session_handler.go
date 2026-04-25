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

// SessionHandler manages incoming API requests and HTTP responses for sessions.
type SessionHandler struct {
	Usecase domain.SessionUsecase
}

// NewSessionHandler registers routes for session management.
func NewSessionHandler(app *fiber.App, us domain.SessionUsecase) {
	handler := &SessionHandler{Usecase: us}

	// Only students should actively manage sessions.
	api := app.Group("/api/v1/exams", middleware.AuthProtected(), middleware.RoleProtected("student"))

	api.Post("/:exam_id/sessions", handler.StartSession)
	api.Get("/sessions/:session_id", handler.ResumeSession)
	api.Post("/sessions/:session_id/submit", handler.SubmitSession)
	api.Get("/:exam_id/my-attempts", handler.GetMyAttempts)
}

// handleSessionError maps session-related domain errors to appropriate HTTP status codes.
func handleSessionError(c *fiber.Ctx, err error) error {
	errMsg := err.Error()
	if strings.HasPrefix(errMsg, "403") {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": errMsg})
	}
	if strings.HasPrefix(errMsg, "409") {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": errMsg})
	}
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": errMsg})
}

// StartSession initializes a new exam session for the current user.
// @Summary Start exam session
// @Description Create a new attempt for a specific exam
// @Tags Sessions
// @Security BearerAuth
// @Produce json
// @Param exam_id path int true "Exam ID"
// @Success 201 {object} domain.SessionResponse "Session started successfully"
// @Failure 403 {object} map[string]string "Forbidden: User cannot start session (in example, max attempts reached)"
// @Failure 409 {object} map[string]string "Conflict: An active session already exists"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/exams/{exam_id}/sessions [post]
func (h *SessionHandler) StartSession(c *fiber.Ctx) error {
	tenantID := uint(c.Locals("tenant_id").(float64))
	userID := uint(c.Locals("user_id").(float64))
	examID, _ := strconv.Atoi(c.Params("exam_id"))
	res, err := h.Usecase.StartSession(c.Context(), tenantID, userID, uint(examID))
	if err != nil {
		return handleSessionError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(res)
}

// ResumeSession retrieves the status of an ongoing exam session.
// @Summary Get active session status
// @Description Retrieve the current state of a specific session.
// @Tags Sessions
// @Security BearerAuth
// @Produce json
// @Param session_id path int true "Session ID"
// @Success 200 {object} domain.SessionResponse "Session details"
// @Failure 404 {object} map[string]string "Not found: Session does not exist"
// @Router /api/v1/exams/sessions/{session_id} [get]
func (h *SessionHandler) ResumeSession(c *fiber.Ctx) error {
	tenantID := uint(c.Locals("tenant_id").(float64))
	userID := uint(c.Locals("user_id").(float64))
	sessionID, _ := strconv.Atoi(c.Params("session_id"))
	res, err := h.Usecase.ResumeSession(c.Context(), tenantID, userID, uint(sessionID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(res)
}

// SubmitSession finalizes an exam attempt and records the user's answers.
// @Summary Submit answers
// @Description Submit user answers and finalize the session, protected by Redis lock to prevent double-submissions
// @Tags Sessions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param session_id path int true "Session ID"
// @Param request body domain.SubmitRequest true "Answers payload"
// @Success 200 "OK: Successfully submitted"
// @Failure 400 {object} map[string]string "Bad request: Invalid payload"
// @Failure 403 {object} map[string]string "Forbidden: Unauthorized to submit this session"
// @Failure 409 {object} map[string]string "Conflict: Already submitted or being processed"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/exams/sessions/{session_id}/submit [post]
func (h *SessionHandler) SubmitSession(c *fiber.Ctx) error {
	tenantID := uint(c.Locals("tenant_id").(float64))
	userID := uint(c.Locals("user_id").(float64))
	sessionID, _ := strconv.Atoi(c.Params("session_id"))
	var req domain.SubmitRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid payload"})
	}
	if err := h.Usecase.SubmitSession(c.Context(), tenantID, userID, uint(sessionID), &req); err != nil {
		return handleSessionError(c, err)
	}
	return c.SendStatus(fiber.StatusOK)
}

// GetMyAttempts retrieves all past and present attempts of the user for a specific exam.
// @Summary List my attempts for an exam
// @Description Fetch a list of sessions/attempts the current user has made for a given exam
// @Tags Sessions
// @Security BearerAuth
// @Produce json
// @Param exam_id path int true "Exam ID"
// @Success 200 {array} domain.ExamSession "List of attempts"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/exams/{exam_id}/my-attempts [get]
func (h *SessionHandler) GetMyAttempts(c *fiber.Ctx) error {
	tenantID := uint(c.Locals("tenant_id").(float64))
	userID := uint(c.Locals("user_id").(float64))
	examID, _ := strconv.Atoi(c.Params("exam_id"))
	sessions, err := h.Usecase.GetMyAttempts(c.Context(), tenantID, userID, uint(examID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(sessions)
}