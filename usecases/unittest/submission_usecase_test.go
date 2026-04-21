// // Import appropriate package.
package unittest

// // Import necessary libraries.
// import (
// 	"bytes"
// 	"errors"
// 	"github.com/redis/go-redis/v9"
// 	"letuan.com/code_demo_backend/domain"
// 	"letuan.com/code_demo_backend/usecases"
// 	"math"
// 	"testing"
// 	"time"
// )

// // =======================================================
// // 1. Shared mock data (package-level scope for all tests)
// // =======================================================

// // dummyExam represents a standard exam configuration used across multiple test functions.
// var dummyExam = &domain.Exam{
// 	ID:					1,
// 	TenantID:			1,
// 	Title:				"Grammar test: Verb forms",
// 	QuestionIDs:		[]uint{1, 2, 7, 9, 10, 15, 16, 17, 18, 20},
// 	MaxAttempts:		1, // Only 1 attempt allowed.
// 	DurationMinutes:	0, // Bypass Redis time checks for standard tests.
// }

// // dummyQuestionsDB simulates the database containing the exam questions and correct answers.
// var dummyQuestionsDB = map[uint]*domain.Question{
// 	1:	{ID: 1, TenantID: 1, Type: "GAP_FILL", CorrectData: map[string][]string{"g1": {"are standing"}}},
// 	2:	{ID: 2, TenantID: 1, Type: "GAP_FILL", CorrectData: map[string][]string{"g1": {"went"}, "g2": {"were playing"}}},
// 	7:	{ID: 7, TenantID: 1, Type: "GAP_FILL", CorrectData: map[string][]string{"g1": {"left"}, "g2": {"went"}}},
// 	9:	{ID: 9, TenantID: 1, Type: "GAP_FILL", CorrectData: map[string][]string{"g1": {"Do you believe"}, "g2": {"don't believe", "do not believe"}}},
// 	10: {ID: 10, TenantID: 1, Type: "GAP_FILL", CorrectData: map[string][]string{"g1": {"have gone"}}},
// 	15: {ID: 15, TenantID: 1, Type: "GAP_FILL", CorrectData: map[string][]string{"g1": {"were going"}}},
// 	16: {ID: 16, TenantID: 1, Type: "GAP_FILL", CorrectData: map[string][]string{"g1": {"were walking"}, "g2": {"were playing"}}},
// 	17: {ID: 17, TenantID: 1, Type: "GAP_FILL", CorrectData: map[string][]string{"g1": {"Have you ever played"}, "g2": {"tried"}, "g3": {"was"}}},
// 	18: {ID: 18, TenantID: 1, Type: "GAP_FILL", CorrectData: map[string][]string{"g1": {"have not won", "haven't won"}, "g2": {"managed"}}},
// 	20: {ID: 20, TenantID: 1, Type: "GAP_FILL", CorrectData: map[string][]string{"g1": {"will travel"}}},
// 	99: {ID: 99, TenantID: 1, Type: "ESSAY", CorrectData: nil}, // For testing unsupported types.
// }

// // ==================================================================
// // 2. Mock repositories (fake database implementations for isolation)
// // ==================================================================

// // mockQuestionRepo provides a fake implementation of domain.QuestionRepository.
// type mockQuestionRepo struct {
// 	mockQuestions	map[uint]*domain.Question
// 	mockError		error
// }

// func (m *mockQuestionRepo) Create(question *domain.Question) error {
// 	return nil
// }

// func (m *mockQuestionRepo) CreateInBatches(questions []domain.Question, batchSize int) error {
// 	return nil
// }

// func (m *mockQuestionRepo) List(tenantID uint, limit int, offset int, tag string) ([]domain.Question, int64, error) {
// 	return nil, 0, nil
// }

// func (m *mockQuestionRepo) GetByIDAndTenant(id uint, tenantID uint) (*domain.Question, error) {
// 	if m.mockError != nil {
// 		return nil, m.mockError
// 	}
// 	if q, exists := m.mockQuestions[id]; exists {
// 		return q, nil
// 	}
// 	return nil, errors.New("Question not found")
// }

// // mockSubmissionRepo provides a fake implementation of domain.SubmissionRepository.
// type mockSubmissionRepo struct {
// 	savedSubmission		*domain.Submission
// 	mockAttempts		int64 // Used to simulate MaxAttempts logic.
// }

// func (m *mockSubmissionRepo) Create(submission *domain.Submission) error {
// 	m.savedSubmission = submission
// 	return nil
// }

// func (m *mockSubmissionRepo) Update(submission *domain.Submission) error {
// 	return nil
// }

// func (m *mockSubmissionRepo) GetStatsByExam(examID uint, tenantID uint, passingScore float64) (*domain.SubmissionStats, error) {
// 	return &domain.SubmissionStats{}, nil
// }

// func (m *mockSubmissionRepo) GetByExamAndTenant(examID uint, tenantID uint) ([]domain.Submission, error) {
// 	return nil, nil
// }

// func (m *mockSubmissionRepo) GetHistoryByUserID(userID uint, tenantID uint) ([]domain.Submission, error) {
// 	return nil, nil
// }

// func (m *mockSubmissionRepo) GetByIDAndTenant(submissionID uint, tenantID uint) (*domain.Submission, error) {
// 	return nil, nil
// }

// func (m *mockSubmissionRepo) CountAttempts(examID uint, userID uint, tenantID uint) (int64, error) {
// 	return m.mockAttempts, nil
// }

// func (m *mockSubmissionRepo) GetLeaderboard(examID uint, tenantID uint, limit int) ([]domain.LeaderboardEntry, error) {
// 	return nil, nil
// }

// // mockExamRepo provides a fake implementation of domain.ExamRepository.
// type mockExamRepo struct {
// 	mockExam	*domain.Exam
// 	mockError	error
// }

// func (m *mockExamRepo) Create(exam *domain.Exam) error {
// 	return nil
// }

// func (m *mockExamRepo) CheckQuestionsInUse(tenantID uint, questionIDs []uint) (bool, error) {
// 	return false, nil
// }

// func (m *mockExamRepo) GetByIDAndTenant(id uint, tenantID uint) (*domain.Exam, error) {
// 	if m.mockError != nil {
// 		return nil, m.mockError
// 	}
// 	return m.mockExam, nil
// }

// // mockCheatLogRepo provides a fake implementation of domain.CheatLogRepository.
// type mockCheatLogRepo struct{

// }

// func (m *mockCheatLogRepo) Create(log *domain.CheatLog) error {
// 	return nil
// }

// func (m *mockCheatLogRepo) GetByExamAndUser(examID uint, userID uint, tenantID uint) ([]domain.CheatLog, error) {
// 	return nil, nil
// }

// // ============================
// // 3. Unit tests implementation
// // ============================

// // TestEvaluateAndSave validates the core grading engine with 12 comprehensive cases.
// func TestEvaluateAndSave(t *testing.T) {
// 	// 3.1. Initialize a dummy Redis client to prevent nil pointer dereference.
// 	dummyRedis := redis.NewClient(&redis.Options{
// 		Addr:				"127.0.0.1:9999", // Fake address.
// 		DialTimeout:		1 * time.Millisecond,
// 		ReadTimeout:		1 * time.Millisecond,
// 		WriteTimeout:		1 * time.Millisecond,
// 	})
// 	defer dummyRedis.Close()
// 	// 3.2. Define time boundaries for testing restrictions.
// 	futureTime := time.Now().Add(24 * time.Hour)
// 	pastTime := time.Now().Add(-24 * time.Hour)
// 	// 3.3. Test table definition.
// 	tests := []struct {
// 		name				string
// 		req					*domain.SubmitRequest
// 		mockExamOverride	*domain.Exam // Used to test exams with specific time or questions.
// 		mockEError			error
// 		mockAttempts		int64
// 		expectedScore		float64
// 		expectedErr			bool
// 		expectedErrMsg		string // Strictly match the returned error.
// 	}{
// 		{
// 			name: "Case 1: 100% score (all 17 gaps correct)",
// 			req: &domain.SubmitRequest{
// 				TenantID: 1, ExamID: 1, UserID: 1,
// 				Answers: map[string]map[string]string{
// 					"1": {"g1": "are standing"},
// 					"2": {"g1": "went", "g2": "were playing"},
// 					"7": {"g1": "left", "g2": "went"},
// 					"9": {"g1": "Do you believe", "g2": "do not believe"},
// 					"10": {"g1": "have gone"},
// 					"15": {"g1": "were going"},
// 					"16": {"g1": "were walking", "g2": "were playing"},
// 					"17": {"g1": "Have you ever played", "g2": "tried", "g3": "was"},
// 					"18": {"g1": "haven't won", "g2": "managed"},
// 					"20": {"g1": "will travel"},
// 				},
// 			},
// 			expectedScore:	100.0,
// 			expectedErr:	false,
// 		},
// 		{
// 			name: "Case 2: Sanitization (spaces, uppercase and alternative correct options)",
// 			req: &domain.SubmitRequest{
// 				TenantID: 1, ExamID: 1, UserID: 2,
// 				Answers: map[string]map[string]string{
// 					"1": {"g1": " ARE standing   "},
// 					"2": {"g1": "WENT", "g2": "  were PLAYING"},
// 					"7": {"g1": "left", "g2": "went"},
// 					"9": {"g1": "DO YOU believe", "g2": "don't believe"},
// 					"10": {"g1": "have gone"},
// 					"15": {"g1": "were  going"},
// 					"16": {"g1": "were walking", "g2": "were playing"},
// 					"17": {"g1": "Have you EVER played", "g2": "tried", "g3": "was"},
// 					"18": {"g1": "have not won", "g2": "managed"},
// 					"20": {"g1": "WILL TRAVEL"},
// 				},
// 			},
// 			expectedScore:	100.0,
// 			expectedErr:	false,
// 		},
// 		{
// 			name: "Case 3: Partial correct (10 out of 17 gaps correct)",
// 			req: &domain.SubmitRequest{
// 				TenantID: 1, ExamID: 1, UserID: 3,
// 				Answers: map[string]map[string]string{
// 					"1": {"g1": "are standing"}, // 1/1 correct
// 					"2": {"g1": "went", "g2": "were playing"}, // 2/2 correct.
// 					"7": {"g1": "left", "g2": "went"}, // 2/2 correct.
// 					"9": {"g1": "Do you believe", "g2": "wrong answer"}, // 1/2 correct.
// 					"10": {"g1": "have gone"}, // 1/1 correct.
// 					"15": {"g1": "were going"}, // 1/1 correct.
// 					"16": {"g1": "were walking", "g2": "wrong"}, // 1/2 correct.
// 					"17": {"g1": "wrong", "g2": "wrong", "g3": "was"}, // 1/3 correct.
// 					"18": {"g1": "wrong", "g2": "wrong"}, // 0/2 correct.
// 					"20": {"g1": "wrong"}, // 0/1 correct.
// 				},
// 			},
// 			expectedScore:	(float64(10) / float64(17)) * 100.0,
// 			expectedErr:	false,
// 		},
// 		{
// 			name: "Case 4: Exceeded max attempts",
// 			req: &domain.SubmitRequest{
// 				TenantID: 1, ExamID: 1, UserID: 4,
// 				Answers: map[string]map[string]string{"1": {"g1": "are standing"}},
// 			},
// 			mockAttempts:	1, // Already took the exam once.
// 			expectedErr:	true,
// 			expectedErrMsg:	"Maximum attempts reached for this exam",
// 		},
// 		{
// 			name: "Case 5: Exam not found",
// 			req: &domain.SubmitRequest{
// 				TenantID: 1, ExamID: 999, UserID: 5,
// 				Answers: map[string]map[string]string{},
// 			},
// 			mockEError:		errors.New("Exam not found"),
// 			expectedErr:	true,
// 			expectedErrMsg: "Exam not found",
// 		},
// 		{
// 			name: "Case 6: 0% score (all gaps are explicitly answered incorrectly)",
// 			req: &domain.SubmitRequest{
// 				TenantID: 1, ExamID: 1, UserID: 6,
// 				Answers: map[string]map[string]string{
// 					"1": {"g1": "wrong answer"},
// 					"2": {"g1": "wrong", "g2": "incorrect"},
// 					"7": {"g1": "bad", "g2": "input"},
// 					"9": {"g1": "no", "g2": "idea"},
// 					"10": {"g1": "something else"},
// 					"15": {"g1": "wrong"},
// 					"16": {"g1": "wrong", "g2": "wrong"},
// 					"17": {"g1": "wrong", "g2": "wrong", "g3": "wrong"},
// 					"18": {"g1": "wrong", "g2": "wrong"},
// 					"20": {"g1": "wrong"},
// 				},
// 			},
// 			expectedScore:	0.0,
// 			expectedErr:	false,
// 		},
// 		{
// 			name: "Case 7: 0% score (empty submission or blank paper)",
// 			req: &domain.SubmitRequest{
// 				TenantID: 1, ExamID: 1, UserID: 7,
// 				Answers: map[string]map[string]string{}, // Blank maps.
// 			},
// 			expectedScore:	0.0,
// 			expectedErr:	false,
// 		},
// 		{
// 			name: "Case 8: Exam has not started yet",
// 			req: &domain.SubmitRequest{
// 				TenantID: 1, ExamID: 8, UserID: 8,
// 				Answers: map[string]map[string]string{"1": {"g1": "are standing"}},
// 			},
// 			mockExamOverride: &domain.Exam{
// 				ID: 8, TenantID: 1, Title: "Future exam", QuestionIDs: []uint{1},
// 				StartTime: &futureTime,
// 			},
// 			expectedErr:	true,
// 			expectedErrMsg: "Exam has not started yet",
// 		},
// 		{
// 			name: "Case 9: Exam has ended",
// 			req: &domain.SubmitRequest{
// 				TenantID: 1, ExamID: 9, UserID: 9,
// 				Answers: map[string]map[string]string{"1": {"g1": "are standing"}},
// 			},
// 			mockExamOverride: &domain.Exam{
// 				ID: 9, TenantID: 1, Title: "Past exam", QuestionIDs: []uint{1},
// 				EndTime: &pastTime,
// 			},
// 			expectedErr:	true,
// 			expectedErrMsg: "Exam has ended",
// 		},
// 		{
// 			name: "Case 10: Time limit exceeded",
// 			// Note: This case is dynamically skipped below to prevent Redis nil pointer panic.
// 			req: &domain.SubmitRequest{
// 				TenantID: 1, ExamID: 10, UserID: 10,
// 				Answers: map[string]map[string]string{},
// 			},
// 			mockExamOverride: &domain.Exam{
// 				ID: 10, TenantID: 1, Title: "Timed exam", QuestionIDs: []uint{1},
// 				DurationMinutes: 60,
// 			},
// 			expectedErr:	true,
// 			expectedErrMsg: "Time limit exceeded",
// 		},
// 		{
// 			name: "Case 11: Partial missing gaps (did first gap, forgot second gap in a question)",
// 			req: &domain.SubmitRequest{
// 				TenantID: 1, ExamID: 1, UserID: 11,
// 				Answers: map[string]map[string]string{
// 					"2": {"g1": "went"}, // Only 1 gap answered in the whole exam.
// 				},
// 			},
// 			expectedScore:	(float64(1) / float64(17)) * 100.0,
// 			expectedErr:	false,
// 		},
// 		{
// 			name: "Case 12: Unsupported question type (contains ESSAY question)",
// 			req: &domain.SubmitRequest{
// 				TenantID: 1, ExamID: 12, UserID: 12,
// 				Answers: map[string]map[string]string{
// 					"1": {"g1": "are standing"}, // Only GAP_FILL should be graded.
// 					"99": {"g1": "This is an essay answer that should be ignored by the engine."},
// 				},
// 			},
// 			mockExamOverride: &domain.Exam{
// 				ID: 12, TenantID: 1, Title: "Mixed exam", QuestionIDs: []uint{1, 99},
// 			},
// 			expectedScore:	100.0, // Question ID 99 is ignored, so 1/1 correct = 100%.
// 			expectedErr:	false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// Skip case 10 gracefully.
// 			// To fully test Redis constraints, a miniredis instance or an interface wrapper for redis.Client is required.
// 			if tt.name == "Case 10: Time limit exceeded" {
// 				t.Skip("Skipping case 10: Require redis.Client interface wrapper or miniredis to mock properly without panicking.")
// 			}
// 			// Decide whether to use the default exam or the overridden one for specific constraints.
// 			examToUse := dummyExam
// 			if tt.mockExamOverride != nil {
// 				examToUse = tt.mockExamOverride
// 			}
// 			mERepo := &mockExamRepo{
// 				mockExam:		examToUse,
// 				mockError:		tt.mockEError,
// 			}
// 			mQRepo := &mockQuestionRepo{
// 				mockQuestions:	dummyQuestionsDB,
// 				mockError:		nil,
// 			}
// 			mSRepo := &mockSubmissionRepo{
// 				mockAttempts:	tt.mockAttempts,
// 			}
// 			mCRepo := &mockCheatLogRepo{

// 			}
// 			// UseCase constructor.
// 			uc := usecases.NewSubmissionUseCase(mSRepo, mERepo, mQRepo, mCRepo, dummyRedis)
// 			result, err := uc.EvaluateAndSave(tt.req)
// 			if tt.expectedErr {
// 				if err == nil {
// 					t.Errorf("Expected an error but got nil")
// 				} else if tt.expectedErrMsg != "" && err.Error() != tt.expectedErrMsg {
// 					t.Errorf("Expected error '%s', but got '%v'", tt.expectedErrMsg, err)
// 				}
// 			} else {
// 				if err != nil {
// 					t.Errorf("Expected no error but got: %v", err)
// 				}
// 				// Use math.Abs to safely compare floating-point precision differences.
// 				if math.Abs(result.Score - tt.expectedScore) > 0.001 {
// 					t.Errorf("Expected score %f, but got %f", tt.expectedScore, result.Score)
// 				}
// 			}
// 		})
// 	}
// }

// // TestExportLeaderboardToExcel verifies that the in-memory Excel generation works correctly without panics.
// func TestExportLeaderboardToExcel(t *testing.T) {
// 	// Initialize a dummy Redis client for this specific test scope to prevent connection issues.
// 	dummyRedis := redis.NewClient(&redis.Options{
// 		Addr:				"127.0.0.1:9999",
// 		DialTimeout:		1 * time.Millisecond,
// 		ReadTimeout: 		1 * time.Millisecond,
// 		WriteTimeout:		1 * time.Millisecond,
// 	})
// 	defer dummyRedis.Close()
// 	// Setup dependencies (using the shared global mocks).
// 	mERepo := &mockExamRepo{mockExam: dummyExam}
// 	mQRepo := &mockQuestionRepo{mockQuestions: dummyQuestionsDB}
// 	mSRepo := &mockSubmissionRepo{mockAttempts: 0}
// 	mCRepo := &mockCheatLogRepo{}
// 	// Inject dependencies.
// 	uc := usecases.NewSubmissionUseCase(mSRepo, mERepo, mQRepo, mCRepo, dummyRedis)
// 	t.Run("Successfully generate Excel buffer", func(t *testing.T) {
// 		buf, err := uc.ExportLeaderboardToExcel(1, 1)
// 		if err != nil {
// 			t.Fatalf("Expected no error, got: %v", err)
// 		}
// 		if buf == nil {
// 			t.Fatal("Expected a valid bytes.Buffer, got nil")
// 		}
// 		if buf.Len() == 0 {
// 			t.Errorf("Expected buffer to have data, got length 0")
// 		}
// 		// Verify the magic bytes of an XLSX file.
// 		fileSignature := buf.Bytes()[:4]
// 		if !bytes.Equal(fileSignature, []byte{0x50, 0x4B, 0x03, 0x04}) {
// 			t.Errorf("Buffer does not contain a valid XLSX file signature")
// 		}
// 	})
// }