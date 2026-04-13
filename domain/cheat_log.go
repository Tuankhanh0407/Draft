// Import appropriate package.
package domain

// Import necessary library.
import "time"

// CheatLog represents a recorded event of potential cheating during an exam.
type CheatLog struct {
	ID			uint			`json:"id" gorm:"primaryKey;autoIncrement"`
	TenantID	uint			`json:"tenant_id" gorm:"not null;index"`
	ExamID		uint			`json:"exam_id" gorm:"not null;index"`
	UserID		uint			`json:"user_id" gorm:"not null;index"`
	EventType	string			`json:"event_type" gorm:"type:varchar(50);not null"` // In example: "tab_switch", "exit_fullscreen", "disconnect".
	CreatedAt	time.Time		`json:"created_at"`
}

// CheatLogRequest is the payload received from the frontend when a cheat event occurs.
type CheatLogRequest struct {
	TenantID	uint			`json:"-"`
	ExamID		uint			`json:"exam_id" validate:"required"`
	UserID		uint			`json:"-"`
	EventType	string			`json:"event_type" validate:"required"`
}

// CheatLogRepository defines database operations for tracking cheating behaviors.
type CheatLogRepository interface {
	// Create inserts a new cheat event into the database.
	Create(log *CheatLog) error
	// GetByExamAndUser fetches the cheat history of a specific student in a specific exam.
	GetByExamAndUser(examID uint, userID uint, tenantID uint) ([]CheatLog, error)
}