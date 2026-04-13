// Import appropriate package.
package domain

// Import necessary library.
import (
	"mime/multipart"
)

// UploadResponse is the payload returned to the client after a successful upload.
type UploadResponse struct {
	FileURL		string		`json:"file_url"`
}

// MediaUseCase defines the business logic for handling media files.
type MediaUseCase interface {
	// UploadFile validates, secures, and uploads a file to cloud storage (S3).
	// It returns the public URL of the uploaded file.
	UploadFile(file *multipart.FileHeader, tenantID uint) (string, error)
}