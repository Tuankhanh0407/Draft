// Import appropriate package.
package http

// Import necessary libraries.
import (
	"github.com/gofiber/fiber/v2"
	"github.com/go-playground/validator/v10"
	"letuan.com/code_demo_backend/domain"
)

// AuthHandler manages authentication and password recovery requests.
type AuthHandler struct {
	AuthUC		domain.AuthUseCase
	Validate	*validator.Validate
}

// NewAuthHandler initializes endpoints for user access and recovery.
func NewAuthHandler(app *fiber.App, uc domain.AuthUseCase, val *validator.Validate) {
	handler := &AuthHandler{
		AuthUC: 	uc,
		Validate:	val,
	}
	api := app.Group("/api/v1/auth")
	// Standard access.
	api.Post("/register", handler.Register)
	api.Post("/login", handler.Login)
	// Password recovery.
	api.Post("/forgot-password", handler.ForgotPassword)
	api.Post("/reset-password", handler.ResetPassword)
}

// Register processes new user sign-ups.
// @Summary Register a new user
// @Description Create a new student or teacher account within a specific tenant.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body domain.AuthRequest true "Registration payload"
// @Success 201 {object} map[string]interface{} "User registered successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 409 {object} map[string]interface{} "Conflict - Username or email already exists"
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req domain.AuthRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if err := h.Validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	user, err := h.AuthUC.Register(&req)
	if err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":	"User registered successfully",
		"data":		user,
	})
}

// Login verifies credentials and returns a signed JWT token.
// @Summary Login to the system
// @Description Authenticate a user and retrieve a Bearer token.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body domain.AuthRequest true "Login credentials"
// @Success 200 {object} map[string]interface{} "Login successful with token"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 401 {object} map[string]interface{} "Unauthorized - Invalid credentials"
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req domain.AuthRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	token, err := h.AuthUC.Login(&req)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":	"Login successful",
		"token":	token,
	})
}

// ForgotPassword triggers the generation and emailing of an OTP.
// @Summary Request password reset OTP
// @Description Generate a 6-digit OTP and send it to the user's email.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body domain.ForgotPasswordRequest true "Email payload"
// @Success 200 {object} map[string]interface{} "OTP sent successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 404 {object} map[string]interface{} "Email not found"
// @Router /api/v1/auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *fiber.Ctx) error {
	var req domain.ForgotPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if err := h.Validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if err := h.AuthUC.RequestPasswordReset(&req); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "If the email exists, an OTP has been sent. It will expire in 5 minutes.",
	})
}

// ResetPassword verifies the OTP and sets the user's new password.
// @Summary Reset user password
// @Description Verify the OTP from email and apply the new password.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body domain.ResetPasswordRequest true "OTP and new password payload"
// @Success 200 {object} map[string]interface{} "Password reset successful"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 401 {object} map[string]interface{} "Unauthorized - Invalid or expired OTP"
// @Router /api/v1/auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *fiber.Ctx) error {
	var req domain.ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if err := h.Validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if err := h.AuthUC.ResetPassword(&req); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Password has been successfully reset. You can now login.",
	})
}