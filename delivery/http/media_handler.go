// Import appropriate package.
package http

// Import necessary libraries.
// import (
// 	"github.com/casbin/casbin/v2"
// 	"github.com/gofiber/fiber/v2"
// 	"letuan.com/code_demo_backend/delivery/http/middleware"
// 	"letuan.com/code_demo_backend/domain"
// )

// // MediaHandler manages HTTP endpoints for file uploads.
// type MediaHandler struct {
// 	MediaUC domain.MediaUseCase
// }

// // NewMediaHandler initializes the routing for media operations securely.
// func NewMediaHandler(app *fiber.App, uc domain.MediaUseCase, enforcer *casbin.Enforcer) {
// 	handler := &MediaHandler{
// 		MediaUC: uc,
// 	}
// 	// Protect media upload route with JWT and Casbin RBAC.
// 	api := app.Group("/api/v1/media", middleware.Protected(), middleware.RoleBasedAuth(enforcer))
// 	api.Post("/upload", handler.UploadMedia)
// }

// // UploadMedia handles multipart/form-data payload, enforcing size limits.
// // @Summary Upload a media file
// // @Description Upload images or audio files to cloud storage (max 10MB).
// // @Tags Media
// // @Accept mpfd
// // @Produce json
// // @Security BearerAuth
// // @Param file formData file true "File to upload"
// // @Success 201 {object} map[string]interface{} "File uploaded successfully"
// // @Failure 400 {object} map[string]interface{} "Invalid form data"
// // @Failure 401 {object} map[string]interface{} "Unauthorized access"
// // @Failure 413 {object} map[string]interface{} "File too large"
// // @Failure 500 {object} map[string]interface{} "Internal server error"
// // @Router /api/v1/media/upload [post]
// func (h *MediaHandler) UploadMedia(c *fiber.Ctx) error {
// 	// 1. Retrieve tenant context to organize files properly.
// 	tenantID, ok := c.Locals("tenant_id").(uint)
// 	if !ok || tenantID == 0 {
// 		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized access"})
// 	}
// 	// 2. Extract the file from the "file" form field.
// 	fileHeader, err := c.FormFile("file")
// 	if err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "File is required in 'file' form field"})
// 	}
// 	// 3. Enforce file size limit (in example, 10 megabytes (MB) max).
// 	const MaxFileSize = 10 * 1024 * 1024 // 10 MB.
// 	if fileHeader.Size > MaxFileSize {
// 		return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{"error": "File exceeds the 10 MB limit"})
// 	}
// 	// 4. Delegate to use case for uploading.
// 	fileURL, err := h.MediaUC.UploadFile(fileHeader, tenantID)
// 	if err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
// 	}
// 	// 5. Return the public URL.
// 	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
// 		"message":	"File uploaded successfully",
// 		"data":		domain.UploadResponse{
// 			FileURL: fileURL,
// 		},
// 	})
// }