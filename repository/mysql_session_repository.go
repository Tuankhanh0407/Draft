// Import appropriate package.
package repository

// Import necessary libraries.
import (
	"context"
	"gorm.io/gorm"
	"letuan.com/code_demo_backend/domain"
)

// mysqlSessionRepository provides MySQL data access for Session entities using GORM.
type mysqlSessionRepository struct {
	db *gorm.DB
}

// NewMysqlSessionRepository creates a new instance of SessionRepository.
func NewMysqlSessionRepository(db *gorm.DB) domain.SessionRepository {
	return &mysqlSessionRepository{db}
}

// CountAttempts counts how many times a user has tried an exam.
// SQL: SELECT COUNT(id) FROM exam_sessions WHERE tenant_id = ? AND exam_id = ? AND user_id = ?;
func (r *mysqlSessionRepository) CountAttempts(ctx context.Context, tenantID, examID, userID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.ExamSession{}).
		Where("tenant_id = ? AND exam_id = ? AND user_id = ?", tenantID, examID, userID).Count(&count).Error
	return count, err
}

// CreateSession initializes a new attempt.
// SQL: INSERT INTO exam_sessions (...) VALUES (...);
func (r *mysqlSessionRepository) CreateSession(ctx context.Context, session *domain.ExamSession) error {
	return r.db.WithContext(ctx).Create(session).Error
}

// GetSession fetches a session by ID.
// SQL: SELECT * FROM exam_sessions WHERE id = ? AND tenant_id = ? LIMIT 1;
func (r *mysqlSessionRepository) GetSession(ctx context.Context, tenantID, sessionID uint) (domain.ExamSession, error) {
	var session domain.ExamSession
	err := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", sessionID, tenantID).First(&session).Error
	return session, err
}

// UpdateSessionSubmit finalizes the exam using optimistic locking via ID.
// SQL: UPDATE exam_sessions SET status = ?, submitted_at = ?, provided_answers = ? WHERE id = ?;
func (r *mysqlSessionRepository) UpdateSessionSubmit(ctx context.Context, session *domain.ExamSession) error {
	return r.db.WithContext(ctx).Save(session).Error
}

// GetMyAttempts returns history for a user.
// SQL: SELECT * FROM exam_sessions WHERE tenant_id = ? AND user_id = ? AND exam_id = ? ORDER BY started_at DESC;
func (r *mysqlSessionRepository) GetMyAttempts(ctx context.Context, tenantID, userID, examID uint) ([]domain.ExamSession, error) {
	var sessions []domain.ExamSession
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND user_id = ? AND exam_id = ?", tenantID, userID, examID).
		Order("started_at DESC").Find(&sessions).Error
	return sessions, err
}