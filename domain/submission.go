// Import appropriate package.
package domain

// Import necessary library.
import (
	"time"
)

// Submission represents a student's completed exam (Stub for analytics).
type Submission struct {
	ID				uint			`json:"id" gorm:"primaryKey;autoIncrement"`
	TenantID		uint			`json:"tenant_id" gorm:"index;not null"`
	ExamID			uint			`json:"exam_id" gorm:"index;not null"`
	UserID			uint			`json:"user_id" gorm:"index;not null"`
	Score			float64			`json:"score"`
	SubmittedAt		time.Time		`json:"submitted_at"`
}