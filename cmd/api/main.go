package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/baolamabcd13/datahiding-text-app/internal/auth"
	"github.com/baolamabcd13/datahiding-text-app/internal/config"
	"github.com/baolamabcd13/datahiding-text-app/internal/email"
	"github.com/baolamabcd13/datahiding-text-app/internal/middleware"
	"github.com/baolamabcd13/datahiding-text-app/internal/models"
	"github.com/baolamabcd13/datahiding-text-app/internal/tasks"
	"github.com/baolamabcd13/datahiding-text-app/internal/user"
	"github.com/baolamabcd13/datahiding-text-app/internal/validation"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq" // PostgreSQL driver
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Tải cấu hình
	cfg := config.LoadConfig()

	// Kiểm tra và tạo database nếu chưa tồn tại
	ensureDBExists(cfg)

	// Kết nối database
	db, err := gorm.Open(postgres.Open(cfg.GetDSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto migrate
	log.Println("Running database migrations...")
	err = db.AutoMigrate(&models.User{}, &models.VerificationToken{}, &models.BlacklistedToken{}, &models.PasswordResetToken{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	log.Println("Database migrations completed successfully")

	// Khởi tạo validation
	validation.Initialize()

	// Khởi tạo email service
	emailConfig := email.Config{
		Host:     os.Getenv("SMTP_HOST"),
		Port:     cfg.EmailPort,
		Username: os.Getenv("SMTP_USERNAME"),
		Password: os.Getenv("SMTP_PASSWORD"),
		From:     os.Getenv("SMTP_FROM"),
		AppURL:   "http://localhost:8080", // Thay đổi thành domain thực tế của ứng dụng
	}
	emailService := email.NewEmailService(emailConfig)

	// Khởi tạo repositories
	authRepo := auth.NewPostgresRepository(db)
	userRepo := user.NewPostgresRepository(db)
	tokenRepo := auth.NewPostgresTokenRepository(db)

	// Khởi tạo auth config
	authConfig := auth.Config{
		JWTSecret:               cfg.JWTSecret,
		JWTExpirationHours:      cfg.JWTExpirationHours,
		EmailVerificationRequired: cfg.EmailVerificationRequired,
	}

	// Khởi tạo services
	authService := auth.NewAuthService(authRepo, cfg.JWTSecret, emailService, authConfig, tokenRepo)
	userService := user.NewUserService(userRepo)

	// Khởi tạo handlers
	authHandler := auth.NewHandler(authService)
	userHandler := user.NewHandler(userService)

	// Khởi tạo middleware
	authMiddleware := middleware.AuthMiddleware(cfg.JWTSecret, tokenRepo)

	// Khởi tạo router
	router := gin.Default()
	
	// Áp dụng middleware
	router.Use(middleware.ValidationMiddleware())

	// Tải templates
	router.LoadHTMLGlob(filepath.Join("templates", "*.html"))

	// Thiết lập routes
	api := router.Group("/api")
	authHandler.SetupRoutes(api)
	userHandler.SetupRoutes(api, authMiddleware)

	// Lên lịch xóa token hết hạn (chạy mỗi 24 giờ)
	tasks.ScheduleTokenCleanup(db, 24*time.Hour)

	// Thêm route cho trang reset password
	router.GET("/reset-password", func(c *gin.Context) {
		token := c.Query("token")
		if token == "" {
			c.HTML(http.StatusBadRequest, "error.html", gin.H{
				"message": "Token không hợp lệ",
			})
			return
		}
		c.HTML(http.StatusOK, "reset_password_form.html", gin.H{
			"Token": token,
		})
	})

	// Khởi chạy server
	log.Printf("Server running on port %s", cfg.ServerPort)
	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// ensureDBExists kiểm tra và tạo database nếu chưa tồn tại
func ensureDBExists(cfg *config.Config) {
	// Kết nối tới PostgreSQL server (sử dụng database mặc định postgres)
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword)
	
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL server: %v", err)
	}
	defer db.Close()

	// Kiểm tra kết nối
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping PostgreSQL server: %v", err)
	}

	// Kiểm tra xem database đã tồn tại chưa
	var exists bool
	query := fmt.Sprintf("SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = '%s');", cfg.DBName)
	err = db.QueryRow(query).Scan(&exists)
	if err != nil {
		log.Fatalf("Failed to check if database exists: %v", err)
	}

	// Nếu database chưa tồn tại, tạo mới
	if !exists {
		log.Printf("Database '%s' does not exist. Creating...", cfg.DBName)
		_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s;", cfg.DBName))
		if err != nil {
			log.Fatalf("Failed to create database: %v", err)
		}
		log.Printf("Database '%s' created successfully", cfg.DBName)
	} else {
		log.Printf("Database '%s' already exists", cfg.DBName)
	}
} 