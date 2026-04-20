// Import appropriate package.
package main // Package main is the entry point of the Assessment Core Engine application.

// Import necessary libraries.
import (
	"github.com/gofiber/fiber/v2"
    "github.com/gofiber/swagger"
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
// @host localhost:8080
// @BasePath /
func main() {
    // Database connection string from .env
    dsn := os.Getenv("DB_DSN")
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal("Failed to connect to database")
    }
    // Auto migration.
    db.AutoMigrate(&domain.Tenant{})
    app := fiber.New()
    // Swagger setup.
    app.Get("/swagger/*", swagger.HandlerDefault)
    // Dependency injection.
    tenantRepo := repository.NewMysqlTenantRepository(db)
    tenantUsecase := usecases.NewTenantUsecase(tenantRepo)
    http.NewTenantHandler(app, tenantUsecase)
    log.Fatal(app.Listen(":8080"))
}