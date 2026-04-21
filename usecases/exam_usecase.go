// // Import appropriate package.
package usecases

// // Import necessary libraries.
// import (
//     "context"
//     "encoding/json"
//     "errors"
//     "fmt"
//     "github.com/redis/go-redis/v9"
//     "letuan.com/code_demo_backend/domain"
//     "sort"
//     "strings"
//     "time"
// )

// // examUseCase implements the domain.ExamUseCase interface.
// type examUseCase struct {
//     examRepo        domain.ExamRepository
//     questionRepo    domain.QuestionRepository
//     submissionRepo  domain.SubmissionRepository
//     cheatLogRepo    domain.CheatLogRepository
//     redisClient     *redis.Client
// }

// // NewExamUseCase constructs the exam business logic layer.
// func NewExamUseCase(eRepo domain.ExamRepository, qRepo domain.QuestionRepository, subRepo domain.SubmissionRepository, cRepo domain.CheatLogRepository, rdb *redis.Client) domain.ExamUseCase {
//     return &examUseCase{
//         examRepo:       eRepo,
//         questionRepo:   qRepo,
//         submissionRepo: subRepo,
//         cheatLogRepo:   cRepo,
//         redisClient:    rdb,
//     }
// }

// // CreateExam validates input and prevents duplicate questions across exams.
// func (u *examUseCase) CreateExam(exam *domain.Exam) error {
//     inUse, err := u.examRepo.CheckQuestionsInUse(exam.TenantID, exam.QuestionIDs)
//     if err != nil {
//         return errors.New("Failed to validate question uniqueness")
//     }
//     if inUse {
//         return errors.New("Invalid request: One or more questions are already used in another exam")
//     }
//     return u.examRepo.Create(exam)
// }

// // GetExamForClient constructs the exam for a student, tracking start time and stripping secrets.
// func (u *examUseCase) GetExamForClient(id uint, tenantID uint, userID uint) (*domain.ExamDetailResponse, error) {
//     ctx := context.Background()
//     cacheKey := fmt.Sprintf("exam:%d:%d", tenantID, id)
//     var responseData *domain.ExamDetailResponse
//     // Try fetching the pre-assembled exam structure from cache.
//     cachedData, err := u.redisClient.Get(ctx, cacheKey).Result()
//     if err == nil {
//         var cachedResponse domain.ExamDetailResponse
//         json.Unmarshal([]byte(cachedData), &cachedResponse)
//         responseData = &cachedResponse
//     } else {
//         // Fetch from database.
//         exam, err := u.examRepo.GetByIDAndTenant(id, tenantID)
//         if err != nil {
//             return nil, errors.New("Exam not found")
//         }
//         var questions []domain.Question
//         for _, qID := range exam.QuestionIDs {
//             q, err := u.questionRepo.GetByIDAndTenant(qID, tenantID)
//             if err == nil {
//                 // Strip sensitive data before sending to frontend.
//                 q.CorrectData = nil
//                 q.Explanation = "" // Ensure explanations are strictly hidden here.
//                 questions = append(questions, *q)
//             }
//         }
//         responseData := &domain.ExamDetailResponse{
//             ExamID:             exam.ID,
//             Title:              exam.Title,
//             DurationMinutes:    exam.DurationMinutes,
//             Questions:          questions,
//         }
//         responseBytes, _ := json.Marshal(responseData)
//         u.redisClient.Set(ctx, cacheKey, responseBytes, time.Hour)
//     }
//     // Initialize start time in Redis to monitor the 'duration_minutes' limits.
//     startKey := fmt.Sprintf("exam_start:%d:%d:%d", tenantID, id, userID)
//     if u.redisClient.Exists(ctx, startKey).Val() == 0 {
//         u.redisClient.Set(ctx, startKey, time.Now().Format(time.RFC3339), 24 * time.Hour)
//     }
//     return responseData, nil
// }

// // GetExamAnalytics provides detailed performance metrics to admins.
// func (u *examUseCase) GetExamAnalytics(examID uint, tenantID uint) (*domain.ExamAnalyticsResponse, error) {
//     exam, err := u.examRepo.GetByIDAndTenant(examID, tenantID)
//     if err != nil {
//         return nil, errors.New("Exam not found")
//     }
//     passingScore := 50.0
//     stats, err := u.submissionRepo.GetStatsByExam(examID, tenantID, passingScore)
//     if err != nil {
//         return nil, err
//     }
//     passRate := 0.0
//     if stats.TotalSubmissions > 0 {
//         passRate = (float64(stats.PassedCount) / float64(stats.TotalSubmissions)) * 100.0
//     }
//     response := &domain.ExamAnalyticsResponse{
//         ExamID:             exam.ID,
//         TotalSubmissions:   stats.TotalSubmissions,
//         AverageScore:       stats.AverageScore,
//         PassRate:           passRate,
//         TopMissedGaps:      []domain.MissedGapStat{},
//     }
//     if stats.TotalSubmissions == 0 {
//         return response, nil
//     }
//     // Aggregate commonly missed gaps in memory.
//     correctDataMap := make(map[uint]map[string][]string)
//     for _, qID := range exam.QuestionIDs {
//         q, err := u.questionRepo.GetByIDAndTenant(qID, tenantID)
//         if err == nil && q.CorrectData != nil {
//             correctDataMap[qID] = q.CorrectData
//         }
//     }
//     submissions, _ := u.submissionRepo.GetByExamAndTenant(examID, tenantID)
//     missCountMap := make(map[string]int)
//     for _, sub := range submissions {
//         for strQID, userAnswers := range sub.Answers {
//             var qID uint
//             fmt.Sscanf(strQID, "%d", &qID)
//             correctAnswers, exists := correctDataMap[qID]
//             if !exists {
//                 continue
//             }
//             for gapID, validOptions := range correctAnswers {
//                 userAns := strings.ToLower(strings.TrimSpace(userAnswers[gapID]))
//                 isCorrect := false
//                 for _, validOpt := range validOptions {
//                     if userAns == strings.ToLower(strings.TrimSpace(validOpt)) {
//                         isCorrect = true
//                         break
//                     }
//                 }
//                 if !isCorrect {
//                     key := fmt.Sprintf("%d|%s", qID, gapID)
//                     missCountMap[key]++;
//                 }
//             }
//         }
//     }
//     // Sort and extract the top 3 missed gaps.
//     var missedGaps []domain.MissedGapStat
//     for key, count := range missCountMap {
//         parts := strings.Split(key, "|")
//         var qID uint
//         fmt.Sscanf(parts[0], "%d", &qID)
//         missedGaps = append(missedGaps, domain.MissedGapStat{
//             QuestionID: qID,
//             GapID:      parts[1],
//             MissCount:  count,
//         })
//     }
//     sort.Slice(missedGaps, func(i, j int) bool {
//         return missedGaps[i].MissCount > missedGaps[j].MissCount
//     })
//     if len(missedGaps) > 3 {
//         missedGaps = missedGaps[:3]
//     }
//     response.TopMissedGaps = missedGaps
//     return response, nil
// }

// // LogCheatEvent processes and stores cheating behaviors reported by the client.
// func (u *examUseCase) LogCheatEvent(req *domain.CheatLogRequest) error {
//     // 1. Verify the exam actually exists to prevent junk data.
//     _, err := u.examRepo.GetByIDAndTenant(req.ExamID, req.TenantID)
//     if err != nil {
//         return errors.New("Exam not found")
//     }
//     // 2. Map request to domain entity.
//     cheatLog := &domain.CheatLog{
//         TenantID:   req.TenantID,
//         ExamID:     req.ExamID,
//         UserID:     req.UserID,
//         EventType:  req.EventType,
//     }
//     // 3. Save the event to MySQL.
//     return u.cheatLogRepo.Create(cheatLog)
// }