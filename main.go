// Import appropriate package.
package main // Package main is the entry point of the Assessment Core Engine application.

// Import necessary libraries.
import (
	"github.com/gofiber/fiber/v2"
    "github.com/gofiber/swagger"
    "github.com/redis/go-redis/v9"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "letuan.com/code_demo_backend/delivery/http"
    _ "letuan.com/code_demo_backend/docs" // Import swagger docs.
    "letuan.com/code_demo_backend/domain"
    "letuan.com/code_demo_backend/repository"
    "letuan.com/code_demo_backend/usecases"
    "log"
    "os"
)

// @title Assessment Core Engine API
// @version 1.0
// @description API for managing multi-tenant assessments
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @host localhost:8080
// @BasePath /
func main() {
    // 1. MySQL connection.
    dsn := os.Getenv("DB_DSN")
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal("Failed to connect to MySQL")
    }
    // 2. Redis connection.
    redisAddr := os.Getenv("REDIS_ADDR")
    rdb := redis.NewClient(&redis.Options{
        Addr: redisAddr,
    })
    // 3. Auto migrate schema.
    db.AutoMigrate(
        &domain.Tenant{},
        &domain.User{},
        &domain.Question{},
        &domain.Exam{},
        &domain.ExamQuestion{},
        &domain.Submission{},
    )
    app := fiber.New()
    // 4. Swagger route.
    app.Get("/swagger/*", swagger.HandlerDefault)
    // 5. Dependency injection.
    // 5.1. Tenants.
    tenantRepo := repository.NewMysqlTenantRepository(db)
    tenantUsecase := usecases.NewTenantUsecase(tenantRepo)
    http.NewTenantHandler(app, tenantUsecase)
    // 5.2. Auth.
    userRepo := repository.NewMysqlUserRepository(db)
    userUsecase := usecases.NewUserUsecase(userRepo)
    http.NewUserHandler(app, userUsecase)
    // 5.3. Questions (protected).
    questionRepo := repository.NewMysqlQuestionRepository(db)
    questionUsecase := usecases.NewQuestionUsecase(questionRepo)
    http.NewQuestionHandler(app, questionUsecase)
    // 5.4. Exams (protected).
    examRepo := repository.NewMysqlExamRepository(db)
    examUsecase := usecases.NewExamUsecase(examRepo, rdb)
    http.NewExamHandler(app, examUsecase)
    // 6. Start server.
    log.Fatal(app.Listen(":8080"))
}