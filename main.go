// Import appropriate package.
package main // Package main is the entry point of the Assessment Core Engine application.

// Import necessary libraries.
import (
	"context"
	"log"
	"os"
	"time"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/recover"
    "github.com/gofiber/swagger"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"letuan.com/code_demo_backend/delivery/http"
    _ "letuan.com/code_demo_backend/docs" // Required for Swagger to find the generated API docs.
	"letuan.com/code_demo_backend/domain"
	"letuan.com/code_demo_backend/repository"
	"letuan.com/code_demo_backend/usecases"
)

// @title Assessment Core Engine API
// @version 1.0
// @description This is a multi-tenant assessment and online examination core engine.
// @termOfService http://swagger.io/terms/
// @contact.name API support
// @contact.email hoangletuan031004@gmail.com
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer " followed by a space and JWT token.
func main() {
    // Load environment variables from .env file.
    if err := godotenv.Load(); err != nil {
        log.Println("Info: Running in Docker/Production mode. Using system environment variables natively.")
    }
    // 1. Initialize infrastructure (MySQL & Redis).
    db := setupDatabase()
    rdb := setupRedis()
    s3Client, bucketName, region := setupS3()
    enforcer := setupCasbin() // Initialize Casbin Enforcer for role-based access control (RBAC).
    // 2. Initialize input validator.
    validate := validator.New()
    // 3. Initialize repositories (data access layer).
    tenantRepo := repository.NewMySQLTenantRepository(db)
    userRepo := repository.NewMySQLUserRepository(db)
    questionRepo := repository.NewMySQLQuestionRepository(db)
    examRepo := repository.NewMySQLExamRepository(db)
    submissionRepo := repository.NewMySQLSubmissionRepository(db)
    cheatLogRepo := repository.NewMySQLCheatLogRepository(db)
    // 4. Initialize usecases (business logic layer).
    tenantUseCase := usecases.NewTenantUseCase(tenantRepo)
    authUseCase := usecases.NewAuthUseCase(userRepo, rdb)
    questionUseCase := usecases.NewQuestionUseCase(questionRepo, rdb)
    examUseCase := usecases.NewExamUseCase(examRepo, questionRepo, submissionRepo, cheatLogRepo, rdb)
    submissionUseCase := usecases.NewSubmissionUseCase(submissionRepo, examRepo, questionRepo, cheatLogRepo, rdb)
    mediaUseCase := usecases.NewMediaUseCase(s3Client, bucketName, region)
    // 5. Initialize Fiber web framework with default configurations.
    app := fiber.New(fiber.Config{
        AppName: "Assessment Core Engine API v1",
    })
    // 6. Setup global middlewares.
    setupMiddlewares(app)
    // 7. Setup infrastructure endpoints (health check & metrics dashboard).
    setupHealthAndMetrics(app, db, rdb)
    app.Get("/", func(c *fiber.Ctx) error {
        return c.SendString("Welcome to Assessment Core Engine API!")
    })
    // ========================================
    // Setup explicit Swagger UI configuration.
    // ========================================
    // Redirect base swagger route to index.html to prevent raw JSON dumps.
    app.Get("/swagger", func(c *fiber.Ctx) error {
        return c.Redirect("/swagger/index.html", fiber.StatusMovedPermanently)
    })
    // Use custom configuration to ensure the graphical user interface (GUI) loads the doc.json properly.
    app.Get("/swagger/*", swagger.New(swagger.Config{
        Title:          "Assessment Core Engine API Docs",
        URL:            "/swagger/doc.json", // Explicitly point to the generated JSON file.
        DeepLinking:    true,
        DocExpansion:   "list", // Expand the groups by default.
    }))
    // 8. Inject dependencies into handlers (including Casbin Enforcer).
    http.NewTenantHandler(app, tenantUseCase, validate, enforcer)
    http.NewAuthHandler(app, authUseCase, validate) // Auth does not need RBAC.
    http.NewQuestionHandler(app, questionUseCase, validate, enforcer)
    http.NewExamHandler(app, examUseCase, submissionUseCase, validate, enforcer)
    http.NewSubmissionHandler(app, submissionUseCase, validate, enforcer)
    http.NewMediaHandler(app, mediaUseCase, enforcer)
    http.NewWebSocketHandler(app, rdb) // Register WebSocket handler for real-time dashboard.
    // 9. Start the HTTP server.
    log.Println("Server is successfully running on port 8080...")
    if err := app.Listen(":8080"); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}

// setupS3 configures the AWS S3 client using credentials from the .env file.
func setupS3() (*s3.Client, string, string) {
    region := os.Getenv("AWS_REGION")
    bucketName := os.Getenv("AWS_BUCKET_NAME")
    accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
    secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
    if region == "" || bucketName == "" || accessKey == "" || secretKey == "" {
        log.Println("Warning: AWS S3 credentials are not fully configured in .env")
        return nil, "", ""
    }
    cfg, err := config.LoadDefaultConfig(
        context.TODO(),
        config.WithRegion(region),
        config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
    )
    if err != nil {
        log.Fatalf("Failed to load AWS configuration: %v", err)
    }
    client := s3.NewFromConfig(cfg)
    log.Println("AWS S3 client connected successfully!")
    return client, bucketName, region
}

// setupCasbin initializes the RBAC model and defines policies programmatically.
func setupCasbin() *casbin.Enforcer {
    // Define the access control model structure.
    text := `
    [request_definition]
    r = sub, obj, act

    [policy_definition]
    p = sub, obj, act

    [role_definition]
    g = _, _

    [policy_effect]
    e = some(where (p.eft == allow))

    [matchers]
    m = g(r.sub, p.sub) && keyMatch2(r.obj, p.obj) && regexMatch(r.act, p.act) || r.sub == "SuperAdmin"
    `
    m, _ := model.NewModelFromString(text)
    e, _ := casbin.NewEnforcer(m)
    // Define policies: (Role, Path, Method/Regex).
    // SuperAdmin: Has all access automatically via matcher (r.sub == "SuperAdmin").

    // TenantAdmin: Full access to their tenant's resources.
    e.AddPolicy("TenantAdmin", "/api/v1/exams", "(GET|POST)")
    e.AddPolicy("TenantAdmin", "/api/v1/exams/*", "(GET|PUT|PATCH|DELETE|POST)")
    e.AddPolicy("TenantAdmin", "/api/v1/exams/*/export", "GET")
    e.AddPolicy("TenantAdmin", "/api/v1/questions", "(GET|POST)")
    e.AddPolicy("TenantAdmin", "/api/v1/questions/*", "(GET|PUT|PATCH|DELETE)")
    e.AddPolicy("TenantAdmin", "/api/v1/submissions/*", "(GET|POST)")
    e.AddPolicy("TenantAdmin", "/api/v1/media/upload", "POST")

    // Teacher: Can create and update exams/questions, but cannot delete.
    e.AddPolicy("Teacher", "/api/v1/exams", "(GET|POST)")
    e.AddPolicy("Teacher", "/api/v1/exams/*", "(GET|PUT|PATCH)")
    e.AddPolicy("Teacher", "/api/v1/exams/*/export", "GET")
    e.AddPolicy("Teacher", "/api/v1/questions", "(GET|POST)")
    e.AddPolicy("Teacher", "/api/v1/questions/*", "(GET|PUT|PATCH)")
    e.AddPolicy("Teacher", "/api/v1/submissions/regrade/*", "POST")
    e.AddPolicy("Teacher", "/api/v1/submissions/leaderboard/*", "GET")
    e.AddPolicy("Teacher", "/api/v1/media/upload", "POST")

    // Student: Can only take exams and view own history.
    e.AddPolicy("Student", "/api/v1/exams/:id", "GET")
    e.AddPolicy("Student", "/api/v1/exams/*/cheat-log", "POST")
    e.AddPolicy("Student", "/api/v1/questions/:id", "GET")
    e.AddPolicy("Student", "/api/v1/submissions", "POST")
    e.AddPolicy("Student", "/api/v1/submissions/draft", "POST")
    e.AddPolicy("Student", "/api/v1/submissions/draft/*", "GET")
    e.AddPolicy("Student", "/api/v1/submissions/leaderboard/*", "GET")
    e.AddPolicy("Student", "/api/v1/users/me/submissions", "GET")
    e.AddPolicy("Student", "/api/v1/users/me/submissions/*", "GET")

    log.Println("Casbin RBAC engine loaded successfully!")
    return e
}

// setupDatabase establishes a connection to MySQL securely using environment variables.
func setupDatabase() *gorm.DB {
    // Dynamically load DSN from .env
    dsn := os.Getenv("DB_DSN")
    if dsn == "" {
        log.Fatal("Error: DB_DSN environment variable is not set")
    }
    var db * gorm.DB
    var err error
    for i := 1; i <= 10; i++ {
        db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
        if err == nil {
            break
        }
        log.Printf("Attempt %d: Failed to connect to MySQL. Retrying in 3 seconds...", i)
        time.Sleep(3 * time.Second)
    }
    if err != nil {
        log.Fatalf("Failed to connect to MySQL after 10 attempts: %v", err)
    }
    // Auto-migrate tables based on updated Domain definitions.
    err = db.AutoMigrate(
        &domain.Tenant{},
        &domain.User{},
        &domain.Question{},
        &domain.Exam{},
        &domain.Submission{},
        &domain.CheatLog{},
    )
    if err != nil {
        log.Fatalf("Database migration failed: %v", err)
    }
    log.Println("Connected to MySQL and migrated tables successfully!")
    return db
}

// setupRedis establishes a connection to the Redis cache server.
func setupRedis() *redis.Client {
    // Dynamically load default port of Redis container from .env
    redisAddr := os.Getenv("REDIS_ADDR")
    if redisAddr == "" {
        log.Fatal("Error: REDIS_ADDR environment variable is not set")
    }
    rdb := redis.NewClient(&redis.Options{
        Addr:       redisAddr,
        Password:   "", // Default password.
        DB:         0, // Default database namespace.
    })
    // Ping to test the connection.
    if err := rdb.Ping(context.Background()).Err(); err != nil {
        log.Fatalf("Failed to connect to Redis: %v", err)
    }
    log.Println("Connected to Redis successfully!")
    return rdb
}

// setupMiddlewares registers global middleware for security and stability.
func setupMiddlewares(app *fiber.App) {
    // Recover middleware prevents the app from crashing on panic.
    app.Use(recover.New())
    // Cross-origin resource sharing (CORS) middleware allows frontend applications to consume the API.
    app.Use(cors.New(cors.Config{
        AllowOrigins: "*",
        AllowHeaders: "Origin, Content-Type, Accept, Authorization",
    }))
    // Rate limiter protects the API against distributed denial of service (DDoS) attacks.
    app.Use(limiter.New(limiter.Config{
        Max:            100, // Max 100 requests per expiration interval.
        Expiration:     1 * time.Minute, // Reset the limit every 1 minute.
        KeyGenerator:   func(c *fiber.Ctx) string {
            return c.IP() // Identify unique users by IP address.
        },
        LimitReached:   func(c *fiber.Ctx) error {
            return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
                "error": "Too many requests. Please try again later",
            })
        },
    }))
}

// setupHealthAndMetrics configures endpoints for DevOps monitoring tools.
func setupHealthAndMetrics(app *fiber.App, db *gorm.DB, rdb *redis.Client) {
    // Health check endpoint.
    app.Get("/health", func(c *fiber.Ctx) error {
        status := fiber.Map{
            "status":       "UP",
            "mysql":        "UP",
            "redis":        "UP",
            "timestamp":    time.Now(),
        }
        // Ping MySQL.
        sqlDB, err := db.DB()
        if err != nil || sqlDB.Ping() != nil {
            status["mysql"] = "DOWN"
            status["status"] = "DOWN"
        }
        // Ping Redis.
        if err := rdb.Ping(context.Background()).Err(); err != nil {
            status["redis"] = "DOWN"
            status["status"] = "DOWN"
        }
        if status["status"] == "DOWN" {
            return c.Status(fiber.StatusServiceUnavailable).JSON(status)
        }
        return c.Status(fiber.StatusOK).JSON(status)
    })
    // Fiber monitor UI dashboard.
    app.Get("/metrics", monitor.New(monitor.Config{
        Title: "Assessment Core Engine metrics dashboard",
    }))
}