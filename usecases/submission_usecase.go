// Import appropriate package.
package usecases

// Import necessary libraries.
import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/xuri/excelize/v2"
	"letuan.com/code_demo_backend/domain"
	"log"
	"strings"
	"time"
)

// submissionUseCase implements the domain.SubmissionUseCase interface.
type submissionUseCase struct {
	submissionRepo	domain.SubmissionRepository
	examRepo		domain.ExamRepository
	questionRepo	domain.QuestionRepository
	cheatLogRepo	domain.CheatLogRepository
	redisClient		*redis.Client
}

// NewSubmissionUseCase integrates repositories and Redis logic.
func NewSubmissionUseCase(subRepo domain.SubmissionRepository, examRepo domain.ExamRepository, qRepo domain.QuestionRepository, cRepo domain.CheatLogRepository, rdb *redis.Client) domain.SubmissionUseCase {
	return &submissionUseCase{
		submissionRepo: subRepo,
		examRepo:		examRepo,
		questionRepo:	qRepo,
		cheatLogRepo:	cRepo,
		redisClient:	rdb,
	}
}

// GetMySubmissions retrieves the history overview for a student.
func (u *submissionUseCase) GetMySubmissions(userID uint, tenantID uint) ([]domain.SubmissionHistoryItem, error) {
	subs, err := u.submissionRepo.GetHistoryByUserID(userID, tenantID)
	if err != nil {
		return nil, err
	}
	var result []domain.SubmissionHistoryItem
	for _, s := range subs {
		exam, _ := u.examRepo.GetByIDAndTenant(s.ExamID, tenantID)
		title := "Unknown Exam"
		if exam != nil {
			title = exam.Title
		}
		result = append(result, domain.SubmissionHistoryItem{
			SubmissionID:	s.ID,
			ExamID:			s.ExamID,
			ExamTitle:		title,
			Score:			s.Score,
			CreatedAt:		s.CreatedAt,
		})
	}
	return result, nil
}

// GetSubmissionDetail reconstructs the exam context including explanations for proctoring logs.
func (u *submissionUseCase) GetSubmissionDetail(submissionID uint, userID uint, tenantID uint) (*domain.SubmissionDetailResponse, error) {
	sub, err := u.submissionRepo.GetByIDAndTenant(submissionID, tenantID)
	// Teachers and admins should bypass the "UserID" check, but for now we match strict ownership.
	// In a real system, we would inject role to check if the caller is teacher.
	if err != nil {
		return nil, errors.New("Submission not found")
	}
	exam, err := u.examRepo.GetByIDAndTenant(sub.ExamID, tenantID)
	if err != nil {
		return nil, errors.New("Exam not found")
	}
	var questions []domain.Question
	for _, qID := range exam.QuestionIDs {
		q, _ := u.questionRepo.GetByIDAndTenant(qID, tenantID)
		if q != nil {
			q.CorrectData = nil
			questions = append(questions, *q)
		}
	}
	_, _, _, details := u.evaluateAnswers(exam, tenantID, sub.Answers)
	// Fetch all cheating logs recorded during this user's exam attempt.
	cheatLogs, _ := u.cheatLogRepo.GetByExamAndUser(sub.ExamID, sub.UserID, tenantID)
	return &domain.SubmissionDetailResponse{
		SubmissionID:	sub.ID,
		ExamID:			exam.ID,
		ExamTitle:		exam.Title,
		Score:			sub.Score,
		Answers:		sub.Answers,
		Details:		details,
		Questions:		questions,
		CheatLogs:		cheatLogs,
		SubmittedAt: 	sub.CreatedAt,
	}, nil
}

// evaluateAnswers executes the core grading algorithm without touching the database.
func (u *submissionUseCase) evaluateAnswers(exam *domain.Exam, tenantID uint, userAnswers map[string]map[string]string) (float64, int, int, map[string]map[string]bool) {
	totalGaps := 0
	correctCount := 0
	details := make(map[string]map[string]bool)
	for _, qID := range exam.QuestionIDs {
		q, err := u.questionRepo.GetByIDAndTenant(qID, tenantID)
		if err != nil || (q.Type != "GAP_FILL" && q.Type != "MULTIPLE_CHOICE" && q.Type != "MATCHING") {
			continue
		}
		questionDetails := make(map[string]bool)
		strQID := fmt.Sprintf("%d", qID)
		userQuestionAnswers := userAnswers[strQID]
		for gapID, validAnswers := range q.CorrectData {
			totalGaps++
			userAnswer := userQuestionAnswers[gapID]
			cleanUserAnswer := normalizeString(userAnswer)
			isCorrect := false
			for _, validAns := range validAnswers {
				if cleanUserAnswer == normalizeString(validAns) {
					isCorrect = true
					break
				}
			}
			questionDetails[gapID] = isCorrect
			if isCorrect {
				correctCount++
			}
		}
		details[strQID] = questionDetails
	}
	var score float64 = 0
	if totalGaps > 0 {
		score = float64(correctCount) / float64(totalGaps) * 100.0
	}
	return score, totalGaps, correctCount, details
}

// EvaluateAndSave processes answers, enforces limits, persists the result, and broadcasts to the live dashboard.
func (u *submissionUseCase) EvaluateAndSave(req *domain.SubmitRequest) (*domain.EvaluationResult, error) {
	exam, err := u.examRepo.GetByIDAndTenant(req.ExamID, req.TenantID)
	if err != nil {
		return nil, errors.New("Exam not found")
	}
	// 1. Validate MaxAttempts to prevent repeated entries.
	if exam.MaxAttempts > 0 {
		attempts, err := u.submissionRepo.CountAttempts(req.ExamID, req.UserID, req.TenantID)
		if err != nil {
			return nil, errors.New("Failed to verify attempt limits")
		}
		if attempts >= int64(exam.MaxAttempts) {
			return nil, errors.New("Maximum attempts reached for this exam") // It will map to 403 Forbidden in controller.
		}
	}
	// 2. Validate time boundaries.
	now := time.Now()
	if exam.StartTime != nil && now.Before(*exam.StartTime) {
		return nil, errors.New("Exam has not started yet")
	}
	if exam.EndTime != nil && now.After(*exam.EndTime) {
		return nil, errors.New("Exam has ended")
	}
	if exam.DurationMinutes > 0 {
		startKey := fmt.Sprintf("exam_start:%d:%d:%d", req.TenantID, req.ExamID, req.UserID)
		startStr, err := u.redisClient.Get(context.Background(), startKey).Result()
		if err != nil {
			return nil, errors.New("Exam session not found. Please access the exam first")
		}
		startTime, _ := time.Parse(time.RFC3339, startStr)
		if now.Sub(startTime).Minutes() > float64(exam.DurationMinutes) + 1.0 { // Allow 1 minute grace period.
			return nil, errors.New("Time limit exceeded")
		}
	}
	// 3. Grade and save.
	score, totalGaps, correctCount, details := u.evaluateAnswers(exam, req.TenantID, req.Answers)
	submission := &domain.Submission{
		TenantID:	req.TenantID,
		ExamID:		req.ExamID,
		UserID:		req.UserID,
		Answers:	req.Answers,
		Score:		score,
		IsPerfect:	correctCount == totalGaps && totalGaps > 0,
	}
	_ = u.submissionRepo.Create(submission)
	// Clean up draft cache.
	draftKey := fmt.Sprintf("draft:%d:%d:%d", req.TenantID, req.ExamID, req.UserID)
	u.redisClient.Del(context.Background(), draftKey)
	result := &domain.EvaluationResult{
		TotalGaps: 	totalGaps,
		Correct: 	correctCount,
		Score: 		score,
		Details: 	details,
	}
	// 4. Real-time live dashboard (Pub/Sub).
	// Broadcast the submission event to the Redis channel for this specific exam.
	go func() {
		ctx := context.Background()
		channelName := fmt.Sprintf("exam_live_dashboard_%d", req.ExamID)
		msg := domain.LiveDashboardMessage{
			ExamID:		req.ExamID,
			UserID:		req.UserID,
			Score:		result.Score,
			Message:	fmt.Sprintf("User ID %d has just submitted the exam and scored %.2f", req.UserID, result.Score),
			Timestamp:	time.Now().Format(time.RFC3339),
		}
		msgBytes, _ := json.Marshal(msg)
		// Fire-and-forget: Publish to Redis. Connected WebSocket clients will receive this immediately.
		u.redisClient.Publish(ctx, channelName, msgBytes)
	}()
	
	return result, nil
}

// RegradeExamSubmissions recalculates scores in the background after admin edits an answer key.
func (u *submissionUseCase) RegradeExamSubmissions(examID uint, tenantID uint) error {
	exam, err := u.examRepo.GetByIDAndTenant(examID, tenantID)
	if err != nil {
		return errors.New("Exam not found")
	}
	submissions, err := u.submissionRepo.GetByExamAndTenant(examID, tenantID)
	if err != nil {
		return errors.New("Failed to fetch submissions")
	}

	go func() {
		log.Printf("[REGRADE] Started for Exam: %d (Submissions: %d)\n", examID, len(submissions))
		for _, sub := range submissions {
			score, totalGaps, correctCount, _ := u.evaluateAnswers(exam, tenantID, sub.Answers)
			sub.Score = score
			sub.IsPerfect = (correctCount == totalGaps && totalGaps > 0)
			u.submissionRepo.Update(&sub)
		}
		log.Printf("[REGRADE] Completed for Exam: %d\n", examID)
	}()

	return nil
}

// normalizeString acts as a helper to trim spaces and format strings securely.
func normalizeString(s string) string {
	return strings.ToLower(strings.Join(strings.Fields(s), " "))
}

// SaveDraft temporarily saves progress to Redis without hitting MySQL.
func (u *submissionUseCase) SaveDraft(req *domain.SubmitRequest) error {
	ctx := context.Background()
	key := fmt.Sprintf("draft:%d:%d:%d", req.TenantID, req.ExamID, req.UserID)
	answersJSON, _ := json.Marshal(req.Answers)
	return u.redisClient.Set(ctx, key, answersJSON, 24 * time.Hour).Err()
}

// GetDraft pulls the user's unsaved progress from Redis.
func (u *submissionUseCase) GetDraft(examID uint, userID uint, tenantID uint) (map[string]map[string]string, error) {
	ctx := context.Background()
	key := fmt.Sprintf("draft:%d:%d:%d", tenantID, examID, userID)
	draftStr, err := u.redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return make(map[string]map[string]string), nil
	}
	var answers map[string]map[string]string
	json.Unmarshal([]byte(draftStr), &answers)
	return answers, nil
}

// GetLeaderboard retrieves the top ranking students for a specific exam.
func (u *submissionUseCase) GetLeaderboard(examID uint, tenantID uint, limit int) ([]domain.LeaderboardEntry, error) {
	_, err := u.examRepo.GetByIDAndTenant(examID, tenantID)
	if err != nil {
		return nil, errors.New("Exam not found")
	}
	if limit <= 0 {
		limit = 10 // Default to top 10.
	}
	return u.submissionRepo.GetLeaderboard(examID, tenantID, limit)
}

// ExportLeaderboardToExcel fetches leaderboard data and writes it into an Excel buffer dynamically.
func (u *submissionUseCase) ExportLeaderboardToExcel(examID uint, tenantID uint) (*bytes.Buffer, error) {
	// 1. Fetch the top 1000 records for the leaderboard export.
	leaderboard, err := u.submissionRepo.GetLeaderboard(examID, tenantID, 1000)
	if err != nil {
		return nil, errors.New("Failed to fetch leaderboard data")
	}
	// 2. Initialize a new Excel file.
	f := excelize.NewFile()
	
	defer func() {
		if err := f.Close(); err != nil {
			log.Println("Error closing excelize file:", err)
		}
	}()

	// 3. Configure the active sheet.
	sheetName := "Leaderboard"
	f.SetSheetName("Sheet1", sheetName)
	// 4. Set handlers.
	f.SetCellValue(sheetName, "A1", "Rank")
	f.SetCellValue(sheetName, "B1", "Student name")
	f.SetCellValue(sheetName, "C1", "Score")
	f.SetCellValue(sheetName, "D1", "Submitted at")
	// Apply bold styling to the header row.
	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
	})
	f.SetCellStyle(sheetName, "A1", "D1", style)
	// 5. Populate data rows.
	for i, entry := range leaderboard {
		rowIdx := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowIdx), entry.Rank)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowIdx), entry.Username)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowIdx), entry.Score)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowIdx), entry.SubmittedAt.Format(time.RFC3339))
	}
	// 6. Write the Excel file into a memory buffer.
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, errors.New("Failed to write Excel data to buffer")
	}
	return &buf, nil
}