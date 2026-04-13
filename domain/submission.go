// Import appropriate package.
package domain

// Import necessary libraries.
import (
	"bytes"
	"gorm.io/gorm"
	"time"
)

// SubmitRequest represents the payload sent by the frontend when submitting answers.
type SubmitRequest struct {
	TenantID			uint							`json:"tenant_id" validate:"required"`
	ExamID				uint							`json:"exam_id" validate:"required"`
	UserID				uint							`json:"user_id"`
	Answers				map[string]map[string]string	`json:"answers" validate:"required"` // Format: map[questionID]map[gapID]answer
}

// EvaluationResult contains the detailed grading result to send back to the client immediately after submission.
type EvaluationResult struct {
	TotalGaps			int								`json:"total_gaps"`
	Correct				int								`json:"correct"`
	Score				float64							`json:"score"`
	Details				map[string]map[string]bool		`json:"details"` // True/False marks for each gap.
}

// Submission represents the permanent record of a user's attempt stored in the database.
type Submission struct {
	ID					uint							`json:"id" gorm:"primaryKey;autoIncrement"`
	TenantID			uint							`json:"tenant_id" gorm:"not null;index"`
	ExamID				uint							`json:"exam_id" gorm:"not null;index"`
	UserID				uint							`json:"user_id" gorm:"index"`
	Answers				map[string]map[string]string	`json:"answers" gorm:"type:json;serializer:json"`
	Score				float64							`json:"score"`
	IsPerfect			bool							`json:"is_perfect"`
	CreatedAt			time.Time						`json:"created_at"`
	DeleteAt			gorm.DeletedAt					`json:"-" gorm:"index"` // Soft delete.
}

// SubmissionStats holds optimized SQL aggregated data for analytics.
type SubmissionStats struct {
	TotalSubmissions	int64							`gorm:"column:total_submissions"`
	AverageScore		float64							`gorm:"column:average_score"`
	PassedCount			int64							`gorn:"column:passed_count"`
}

// SubmissionHistoryItem represents a simplified item for the history list view.
type SubmissionHistoryItem struct {
	SubmissionID		uint							`json:"submission_id"`
	ExamID				uint							`json:"exam_id"`
	ExamTitle			string							`json:"exam_title"`
	Score				float64							`json:"score"`
	CreatedAt			time.Time						`json:"submitted_at"`
}

// SubmissionDetailResponse represents the full breakdown when reviewing a past exam.
type SubmissionDetailResponse struct {
	SubmissionID		uint							`json:"submission_id"`
	ExamID				uint							`json:"exam_id"`
	ExamTitle			string							`json:"exam_title"`
	Score				float64							`json:"score"`
	Answers				map[string]map[string]string	`json:"answers"` // User's submitted answers.
	Details				map[string]map[string]bool		`json:"details"` // Re-calculated true/false marks.
	Questions			[]Question						`json:"questions"` // Exam structure WITH explanations but NO correct answers.
	CheatLogs			[]CheatLog						`json:"cheat_logs"` // Proctoring history for this attempt.
	SubmittedAt			time.Time						`json:"submitted_at"`
}

// LeaderboardEntry represents a single row in the leaderboard response.
type LeaderboardEntry struct {
	Rank				int								`json:"rank"`
	Username			string							`json:"username"`
	Score				float64							`json:"score"`
	SubmittedAt			time.Time						`json:"submitted_at"`
}

// LiveDashboardMessage represents the payload broadcasted via WebSockets to teachers.
type LiveDashboardMessage struct {
	ExamID				uint							`json:"exam_id"`
	UserID				uint							`json:"user_id"`
	Username			string							`json:"username,omitempty"`
	Score				float64							`json:"score"`
	Message				string							`json:"message"`
	Timestamp			string							`json:"timestamp"`
}

// SubmissionRepository defines database operations for submissions.
type SubmissionRepository interface {
	// Create saves a graded submission record to the database.
	Create(submission *Submission) error
	// GetStatsByExam fetches aggregated stats directly using raw SQL for high performance.
	GetStatsByExam(examID uint, tenantID uint, passingScore float64) (*SubmissionStats, error)
	// GetByExamAndTenant fetches all submissions for a specific exam to analyze detailed gaps.
	GetByExamAndTenant(examID uint, tenantID uint) ([]Submission, error)
	// Update modifies an existing submission (used during background auto-regrading).
	Update(submission *Submission) error
	// GetHistoryByUserID fetches the submission history belonging to a specific student.
	GetHistoryByUserID(userID uint, tenantID uint) ([]Submission, error)
	// GetByIDAndTenant retrieves a single submission record by its primary key.
	GetByIDAndTenant(submissionID uint, tenantID uint) (*Submission, error)
	// CountAttempts calculates how many times a user has submitted a specific exam.
	CountAttempts(userID uint, examID uint, tenantID uint) (int64, error)
	// GetLeaderboard fetches the top submissions joined with users table to get usernames.
	GetLeaderboard(examID uint, tenantID uint, limit int) ([]LeaderboardEntry, error)
}

// SubmissionUseCase defines business logic for grading and tracking progress.
type SubmissionUseCase interface {
	// EvaluateAndSave processes answers, enforces MaxAttempts and Time limits, and persists the result.
	EvaluateAndSave(req *SubmitRequest) (*EvaluationResult, error)
	// RegradeExamSubmissions runs a background job to recalculate scores if the answer key changes.
	RegradeExamSubmissions(examID uint, tenantID uint) error
	// GetMySubmissions retrieves a high-level list of past exams for the student.
	GetMySubmissions(userID uint, tenantID uint) ([]SubmissionHistoryItem, error)
	// GetSubmissionDetail reconstructs the exam context to show the student exactly where they failed, including explanations.
	GetSubmissionDetail(submissionID uint, userID uint, tenantID uint) (*SubmissionDetailResponse, error)
	// SaveDraft temporarily saves the student's progress to Redis to prevent data loss.
	SaveDraft(req *SubmitRequest) error
	// GetDraft retrieves the temporary auto-saved answers from Redis.
	GetDraft(examID uint, userID uint, tenantID uint) (map[string]map[string]string, error)
	// GetLeaderboard retrieves the top ranking students for a specific exam.
	GetLeaderboard(examID uint, tenantID uint, limit int) ([]LeaderboardEntry, error)
	// ExportLeaderboardToExcel generates an Excel file containing the exam leaderboard in memory.
	ExportLeaderboardToExcel(examID uint, tenantID uint) (*bytes.Buffer, error)
}