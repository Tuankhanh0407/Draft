// Import appropriate package.
package domain

// Import necessary libraries.
import (
	"context"
	"gorm.io/gorm"
	"time"
)

// Exam represents a test containing multiple questions.
type Exam struct {
	ID					uint				`json:"id" gorm:"primaryKey;autoIncrement"`
	TenantID			uint				`json:"-" gorm:"index;not null"`
	Title				string				`json:"title" gorm:"size:255;not null"`
	Duration			int					`json:"duration" gorm:"not null"` // Duration in minutes.
	CreatedAt			time.Time			`json:"created_at"`
	UpdatedAt			time.Time			`json:"updated_at"`
	DeletedAt			gorm.DeletedAt		`json:"-" gorm:"index"`
}

// ExamQuestion is the mapping table.
// UniqueIndex on QuestionID ensures no question reuse.
type ExamQuestion struct {
	ExamID				uint				`gorm:"primaryKey"`
	QuestionID			uint				`gorm:"primaryKey;uniqueIndex:idx_unique_question"`
}

// ExamRequest represents the payload for creating/updating an exam.
type ExamRequest struct {
	Title				string				`json:"title" validate:"required"`
	Duration			int					`json:"duration" validate:"required,min=1"`
	QuestionIDs			[]uint				`json:"question_ids" validate:"required,min=10"`
}

// ExamResponse represents the exam payload returned to clients.
type ExamResponse struct {
	ID					uint				`json:"id"`
	Title				string				`json:"title"`
	Duration			int					`json:"duration"`
	Questions			[]QuestionResponse	`json:"questions,omitempty"`
	CreatedAt			time.Time			`json:"created_at"`
}

// AnalyticsResponse represents basic statistics for an exam.
type AnalyticsResponse struct {
	ExamID				uint				`json:"exam_id"`
	TotalSubmissions	int64				`json:"total_submissions"`
	AverageScore		float64				`json:"average_score"`
}

// ExamRepository defines data access methods for exams.
type ExamRepository interface {
	Fetch(ctx context.Context, tenantID uint, limit, offset int) ([]Exam, error)
	GetByID(ctx context.Context, tenantID, id uint) (Exam, []Question, error)
	Create(ctx context.Context, exam *Exam, questionIDs []uint) error
	Update(ctx context.Context, exam *Exam, questionIDs []uint) error
	Delete(ctx context.Context, tenantID, id uint) error
	GetAnalytics(ctx context.Context, tenantID, id uint) (AnalyticsResponse, error)
}

// ExamUsecase defines business logic for exam management.
type ExamUsecase interface {
	GetAll(ctx context.Context, tenantID uint, page, limit int) ([]ExamResponse, error)
	GetByID(ctx context.Context, tenantID, id uint, role string) (ExamResponse, error)
	Create(ctx context.Context, tenantID uint, req *ExamRequest) (ExamResponse, error)
	Update(ctx context.Context, tenantID, id uint, req *ExamRequest) (ExamResponse, error)
	Delete(ctx context.Context, tenantID, id uint) error
	GetAnalytics(ctx context.Context, tenantID, id uint) (AnalyticsResponse, error)
}