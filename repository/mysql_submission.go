// Import appropriate package.
package repository

// Import necessary libraries.
import (
	"gorm.io/gorm"
	"letuan.com/code_demo_backend/domain"
)

// mysqlSubmissionRepository represents a MySQL-backed repository for managing submission data using GORM.
type mysqlSubmissionRepository struct {
	db *gorm.DB
}

// NewMySQLSubmissionRepository creates a new instance of SubmissionRepository.
func NewMySQLSubmissionRepository(db *gorm.DB) domain.SubmissionRepository {
	return &mysqlSubmissionRepository{db: db}
}

// Create inserts a new submission record into the database.
func (m *mysqlSubmissionRepository) Create(submission *domain.Submission) error {
	return m.db.Create(submission).Error
}

// GetStatsByExam uses raw SQL aggregation functions to calculate statistics efficiently.
func (m *mysqlSubmissionRepository) GetStatsByExam(examID uint, tenantID uint, passingScore float64) (*domain.SubmissionStats, error) {
	var stats domain.SubmissionStats
	// Use COALESCE to return 0 instead of NULL if there are no submissions.
	// Use SUM(CASE WHEN...) to count how many students passed.
	err := m.db.Model(&domain.Submission{}).
		Select("COUNT(id) as total_submissions, COALESCE(AVG(score), 0) as average_score, COALESCE(SUM(CASE WHEN score >= ? THEN 1 ELSE 0 END), 0) as passed_count", passingScore).
		Where("exam_id = ? AND tenant_id = ?", examID, tenantID).
		Scan(&stats).Error
	return &stats, err
}

// GetByExamAndTenant fetches all submissions made by a specific student.
func (m *mysqlSubmissionRepository) GetByExamAndTenant(examID uint, tenantID uint) ([]domain.Submission, error) {
	var submissions []domain.Submission
	err := m.db.Where("exam_id = ? AND tenant_id = ?", examID, tenantID).Find(&submissions).Error
	return submissions, err
}

// Update modifies an existing submission record (in example, after regrading).
func (m *mysqlSubmissionRepository) Update(submission *domain.Submission) error {
	return m.db.Save(submission).Error
}

// GetHistoryByUserID fetches all submissions made by a specific student.
func (m *mysqlSubmissionRepository) GetHistoryByUserID(userID uint, tenantID uint) ([]domain.Submission, error) {
	var submissions []domain.Submission
	err := m.db.Where("user_id = ? AND tenant_id = ?", userID, tenantID).
		Order("created_at desc").
		Find(&submissions).Error
	return submissions, err
}

// GetByIDAndTenant fetches a single submission by its primary key securely.
func (m *mysqlSubmissionRepository) GetByIDAndTenant(submissionID uint, tenantID uint) (*domain.Submission, error) {
	var sub domain.Submission
	err := m.db.Where("id = ? AND tenant_id = ?", submissionID, tenantID).First(&sub).Error
	return &sub, err
}

// CountAttempts calculates how many times a specific user has submitted a specific exam.
func (m *mysqlSubmissionRepository) CountAttempts(examID uint, userID uint, tenantID uint) (int64, error) {
	var count int64
	err := m.db.Model(&domain.Submission{}).
		Where("exam_id = ? AND user_id = ? AND tenant_id = ?", examID, userID, tenantID).
		Count(&count).Error
	return count, err
}

// GetLeaderboard fetches the top submissions joined with users table to retrieve usernames.
func (m *mysqlSubmissionRepository) GetLeaderboard(examID uint, tenantID uint, limit int) ([]domain.LeaderboardEntry, error) {
	var entries []domain.LeaderboardEntry
	// Using GORM to JOIN 'submissions' with 'users' to fetch the Username.
	// Ordered by Score (highest first) and CreatedAt (earliest first to break ties).
	err := m.db.Table("submissions").
		Select("users.username, submissions.score, submissions.created_at as submitted_at").
		Joins("INNER JOIN users ON submissions.user_id = users.id").
		Where("submissions.exam_id = ? AND submissions.tenant_id = ?", examID, tenantID).
		Order("submissions.score DESC, submissions.created_at ASC").
		Limit(limit).
		Scan(&entries).Error
	if err != nil {
		return nil, err
	}
	// Application logic: Assign ranks iteratively to the results.
	for i := range entries {
		entries[i].Rank = i + 1
	}
	return entries, nil
}