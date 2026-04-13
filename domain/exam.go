// Import appropriate package.
package domain

// Import necessary libraries.
import (
	"gorm.io/gorm"
	"time"
)

// Exam represents a test containing multiple compiled questions.
type Exam struct {
	ID					uint				`json:"exam_id" gorm:"primaryKey;autoIncrement"`
	TenantID			uint				`json:"tenant_id" gorm:"not null;index" validate:"required"`
	Title				string				`json:"title" gorm:"type:varchar(200);not null" validate:"required"`
	QuestionIDs			[]uint				`json:"question_ids" gorm:"type:json;serializer:json" validate:"required"` // Array of referenced question IDs.
	StartTime			*time.Time			`json:"start_time" gorm:"type:datetime"` // Nullable: No start limit if null.
	EndTime				*time.Time			`json:"end_time" gorm:"type:datetime"` // Nullable: No end limit if null.
	DurationMinutes		int					`json:"duration_minutes" gorm:"type:int;default:0"` // 0 means unlimited duration.
	MaxAttempts			int					`json:"max_attempts" gorm:"type:int;default:0"` // Limit number of times a user can take this exam (0 = unlimited).
	CreatedAt			time.Time			`json:"created_at"`
	UpdatedAt			time.Time			`json:"updated_at"`
	DeletedAt			gorm.DeletedAt		`json:"-" gorm:"index"` // Soft delete.
}

// ExamDetailResponse represents the formatted payload sent to the frontend when starting an exam.
type ExamDetailResponse struct {
	ExamID				uint				`json:"exam_id"`
	Title				string				`json:"title"`
	DurationMinutes		int					`json:"duration_minutes"` // Used by frontend to render a countdown timer.
	Questions			[]Question			`json:"questions"` // Contain AST nodes but NO correct_data and NO explanation.
}

// MissedGapStat holds statistics for a specific incorrectly answered gap.
type MissedGapStat struct {
	QuestionID			uint				`json:"question_id"`
	GapID				string				`json:"gap_id"`
	MissCount			int					`json:"miss_count"`
}

// ExamAnalyticsResponse provides an overview of performance metrics for an exam.
type ExamAnalyticsResponse struct {
	ExamID				uint				`json:"exam_id"`
	TotalSubmissions	int64				`json:"total_submissions"`
	AverageScore		float64				`json:"average_score"`
	PassRate			float64				`json:"pass_rate"` // Percentage (0 - 100).
	TopMissedGaps		[]MissedGapStat		`json:"top_missed_gaps"`
}

// ExamRepository defines database operations for exams.
type ExamRepository interface {
	// Create inserts a new exam record into the database.
	Create(exam *Exam) error
	// GetByIDAndTenant retrieves an exam, strictly bound to its tenant.
	GetByIDAndTenant(id uint, tenantID uint) (*Exam, error)
	// CheckQuestionsInUse verifies if any question IDs are already part of other exams to prevent duplicates.
	CheckQuestionsInUse(tenantID uint, questionIDs []uint) (bool, error)
}

// ExamUseCase defines business logic for exams.
type ExamUseCase interface {
	// CreateExam validates exam parameters and ensures question uniqueness before creation.
	CreateExam(exam *Exam) error
	// GetExamForClient aggregates questions and tracks the start time in Redis for duration calculation.
	GetExamForClient(id uint, tenantID uint, userID uint) (*ExamDetailResponse, error)
	// GetExamAnalytics aggregates submission data to highlight common student mistakes.
	GetExamAnalytics(examID uint, tenantID uint) (*ExamAnalyticsResponse, error)
	// LogCheatEvent records a proctoring alert during an active exam.
	LogCheatEvent(req *CheatLogRequest) error
}