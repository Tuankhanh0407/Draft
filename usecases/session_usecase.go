// Import appropriate package.
package usecases

// Import necessary libraries.
import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"gorm.io/datatypes"
	"letuan.com/code_demo_backend/domain"
	"time"
)

// sessionUsecase implements the business logic for managing sessions.
type sessionUsecase struct {
	sessionRepo		domain.SessionRepository
	examRepo		domain.ExamRepository
	redisClient		*redis.Client
}

// NewSessionUsecase initializes the session usecase with its required repository.
func NewSessionUsecase(sr domain.SessionRepository, er domain.ExamRepository, rdb *redis.Client) domain.SessionUsecase {
	return &sessionUsecase{sr, er, rdb}
}

// getVNTime ensures all server operations strictly follow (UTC+7).
func getVNTime() time.Time {
	vnZone := time.FixedZone("ICT", 7 * 3600)
	return time.Now().In(vnZone)
}

// StartSession initiates a new exam attempt for a user.
func (u *sessionUsecase) StartSession(ctx context.Context, tenantID, userID, examID uint) (domain.SessionResponse, error) {
	now := getVNTime()
	// 1. Fetch exam & validate bounds.
	exam, _, err := u.examRepo.GetByID(ctx, tenantID, examID)
	if err != nil {
		return domain.SessionResponse{}, errors.New("Exam not found")
	}
	if now.Before(exam.ValidFrom) {
		return domain.SessionResponse{}, errors.New("403: Exam has not started yet")
	}
	if now.After(exam.ValidTo) {
		return domain.SessionResponse{}, errors.New("403: Exam has ended")
	}
	// 2. Enforce max attempts.
	attempts, _ := u.sessionRepo.CountAttempts(ctx, tenantID, examID, userID)
	if attempts >= int64(exam.MaxAttempts) {
		return domain.SessionResponse{}, errors.New("409: You have reached the maximum number of attempts")
	}
	// 3. Dynamic duration logic (min of duration or time left until valid_to).
	plannedEndTime := now.Add(time.Duration(exam.Duration) * time.Minute)
	actualEndTime := plannedEndTime
	if plannedEndTime.After(exam.ValidTo) {
		actualEndTime = exam.ValidTo // Force cut-off if they start late.
	}
	session := domain.ExamSession{
		TenantID:	tenantID,
		ExamID:		examID,
		UserID:		userID,
		Status:		domain.StatusInProgress,
		StartedAt:	now,
		EndedAt:	actualEndTime,
	}
	if err := u.sessionRepo.CreateSession(ctx, &session); err != nil {
		return domain.SessionResponse{}, err
	}
	return domain.SessionResponse{
		SessionID:	session.ID,
		ExamID:		session.ExamID,
		Status:		session.Status,
		StartedAt:	session.StartedAt,
		EndedAt:	session.EndedAt,
		TimeLeftMs:	actualEndTime.Sub(now).Milliseconds(),
	}, nil
}

// ResumeSession retrieves an ongoing exam session for a user and calculates the exact remaining time.
func (u *sessionUsecase) ResumeSession(ctx context.Context, tenantID, userID, sessionID uint) (domain.SessionResponse, error) {
	session, err := u.sessionRepo.GetSession(ctx, tenantID, sessionID)
	if err != nil || session.UserID != userID {
		return domain.SessionResponse{}, errors.New("403: Session not found or unauthorized")
	}
	now := getVNTime()
	timeLeftMs := session.EndedAt.Sub(now).Milliseconds()
	if timeLeftMs < 0 {
		timeLeftMs = 0
	}
	return domain.SessionResponse{
		SessionID:	session.ID,
		ExamID:		session.ExamID,
		Status:		session.Status,
		StartedAt:	session.StartedAt,
		EndedAt:	session.EndedAt,
		TimeLeftMs: timeLeftMs,
	}, nil
}

// SubmitSession finalizes an exam attempt and records the user's answers.
func (u *sessionUsecase) SubmitSession(ctx context.Context, tenantID, userID, sessionID uint, req *domain.SubmitRequest) error {
	now := getVNTime()
	// 1. Redis distributed lock to prevent double submit (spam clicks).
	lockKey := fmt.Sprintf("lock:submit:%d:%d", tenantID, sessionID)
	err := u.redisClient.SetArgs(ctx, lockKey, "locked", redis.SetArgs{
		Mode:		"NX",
		TTL:		10 * time.Second,
	}).Err()
	if err == redis.Nil {
		return errors.New("409: Submission is already being processed")
	} else if err != nil {
		return err
	}
	defer u.redisClient.Del(ctx, lockKey) // Ensure lock is released.
	// 2. Fetch session.
	session, err := u.sessionRepo.GetSession(ctx, tenantID, sessionID)
	if err != nil || session.UserID != userID {
		return errors.New("403: Session not found or unauthorized")
	}
	if session.Status != domain.StatusInProgress {
		return errors.New("409: Session already submitted or closed")
	}
	// 3. Grace period validation (10 seconds for network latency).
	gracePeriodDeadline := session.EndedAt.Add(10 * time.Second)
	if now.After(gracePeriodDeadline) {
		session.Status = domain.StatusLate
	} else {
		session.Status = domain.StatusSubmitted
	}
	answersBytes, _ := json.Marshal(req.Answers)
	session.ProvidedAnswers = datatypes.JSON(answersBytes)
	session.SubmittedAt = &now
	return u.sessionRepo.UpdateSessionSubmit(ctx, &session)
}

// GetMyAttempts retrieves a history of all sessions taken by a specific user for a given exam.
func (u *sessionUsecase) GetMyAttempts(ctx context.Context, tenantID, userID, examID uint) ([]domain.ExamSession, error) {
	return u.sessionRepo.GetMyAttempts(ctx, tenantID, userID, examID)
}