// Import appropriate package.
package usecases

// Import necessary libraries.
import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"letuan.com/code_demo_backend/domain"
)

// mediaUseCase implements the domain.MediaUseCase interface using Amazon Web Services (AWS) S3.
type mediaUseCase struct {
	s3Client	*s3.Client
	bucketName	string
	region		string
}

// NewMediaUseCase creates a new instance of MediaUseCase.
func NewMediaUseCase(client *s3.Client, bucketName string, region string) domain.MediaUseCase {
	return &mediaUseCase{
		s3Client:	client,
		bucketName: bucketName,
		region:		region,
	}
}

// UploadFile processes the multipart file and streams it to AWS S3.
func (u *mediaUseCase) UploadFile(fileHeader *multipart.FileHeader, tenantID uint) (string, error) {
	// 1. Open the file stream.
	file, err := fileHeader.Open()
	if err != nil {
		return "", errors.New("Failed to open file stream")
	}
	defer file.Close()
	// 2. Validate file extension (security check).
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	allowedExts := map[string]bool{
		".jpg": true,
		".jpeg": true,
		".png": true,
		".mp3": true,
		".wav": true,
	}
	if !allowedExts[ext] {
		return "", errors.New("Unsupported file type. Allowed: jpg, png, mp3, wav")
	}
	// 3. Generate a unique, secure file key (Path in S3).
	// Format: tenants/{tenant_id}/media/{uuid}{ext}
	uniqueID := uuid.New().String()
	fileKey := fmt.Sprintf("tenants/%d/media/%s%s", tenantID, uniqueID, ext)
	// 4. Determine Content-Type dynamically.
	contentType := "application/octet-stream"
	switch ext {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".mp3":
		contentType = "audio/mpeg"
	case ".wav":
		contentType = "audio/wav"
	}
	// 5. Upload to S3.
	_, err = u.s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:			aws.String(u.bucketName),
		Key:			aws.String(fileKey),
		Body:			file,
		ContentType:	aws.String(contentType),
		// Access control list (ACL) is often controlled by bucket policies now, but setting it public-read is common for media.
	})
	if err != nil {
		return "", fmt.Errorf("Failed to upload file to storage: %v", err)
	}
	// 6. Construct and return the public URL.
	// Note: If using a custion CDN (CloudFront) or MinIO, this URL format might differ.
	publicURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", u.bucketName, u.region, fileKey)
	return publicURL, nil
}