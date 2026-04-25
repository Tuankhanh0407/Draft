// Import appropriate package.
package domain

// Import necessary libraries.
import (
    "context"
    "gorm.io/datatypes"
    "time"
)

// Constants for session status.
const (
    StatusInProgress = "IN_PROGRESS"
    StatusSubmitted = "SUBMITTED"
    StatusLate = "LATE_REJECTED"
)

// ExamSession tracks a user's attempt at an exam.
type ExamSession struct {
    ID                  uint                    `json:"id" gorm:"primaryKey;autoIncrement"`
    TenantID            uint                    `json:"-" gorm:"index;not null"`
    ExamID              uint                    `json:"exam_id" gorm:"index;not null"`
    UserID              uint                    `json:"user_id" gorm:"index;not null"`
    Status              string                  `json:"status" gorm:"size:50;not null"`
    StartedAt           time.Time               `json:"started_at"`
    EndedAt             time.Time               `json:"ended_at"` // Dynamic deadline calculated at start.
    SubmittedAt         *time.Time              `json:"submitted_at"`
    ProvidedAnswers     datatypes.JSON          `json:"provided_answers" gorm:"type:json" swaggertype:"object"`
}

// SubmitRequest represents the payload from the frontend.
type SubmitRequest struct {
    Answers             map[string]interface{}  `json:"answers" validate:"required"`
}

// SessionResponse is what the frontend needs to render the timer.
type SessionResponse struct {
    SessionID           uint                    `json:"session_id"`
    ExamID              uint                    `json:"exam_id"`
    Status              string                  `json:"status"`
    StartedAt           time.Time               `json:"started_at"`
    EndedAt             time.Time               `json:"ended_at"`
    TimeLeftMs          int64                   `json:"time_left_ms"`
}

// SessionRepository defines database operations for sessions.
type SessionRepository interface {
    CountAttempts(ctx context.Context, tenantID, examID, userID uint) (int64, error)
    CreateSession(ctx context.Context, session *ExamSession) error
    GetSession(ctx context.Context, tenantID, sessionID uint) (ExamSession, error)
    GetMyAttempts(ctx context.Context, tenantID, userID, examID uint) ([]ExamSession, error)
    UpdateSessionSubmit(ctx context.Context, session *ExamSession) error
}

// SessionUsecase handles the core time and state logic.
type SessionUsecase interface {
    StartSession(ctx context.Context, tenantID, userID, examID uint) (SessionResponse, error)
    ResumeSession(ctx context.Context, tenantID, userID, sessionID uint) (SessionResponse, error)
    SubmitSession(ctx context.Context, tenantID, userID, sessionID uint, req *SubmitRequest) error
    GetMyAttempts(ctx context.Context, tenantID, userID, examID uint) ([]ExamSession, error)
}