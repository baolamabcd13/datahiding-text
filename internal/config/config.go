package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config - Cấu hình ứng dụng
type Config struct {
	DBHost                  string
	DBPort                  string
	DBUser                  string
	DBPassword              string
	DBName                  string
	ServerPort              string
	JWTSecret               string
	JWTExpirationHours      int
	EmailVerificationRequired bool
	EmailPort               int
	AppURL                  string
	CORSAllowOrigins        []string
}

// LoadConfig - Tải cấu hình từ file .env
func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Đọc cấu hình database
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "datahiding_text_app")

	// Đọc cấu hình server
	serverPort := getEnv("SERVER_PORT", "8080")

	// Đọc cấu hình JWT
	jwtSecret := getEnv("JWT_SECRET", "your-secret-key")
	jwtExpirationHoursStr := getEnv("JWT_EXPIRATION_HOURS", "24")
	jwtExpirationHours, err := strconv.Atoi(jwtExpirationHoursStr)
	if err != nil {
		log.Printf("Warning: Invalid JWT_EXPIRATION_HOURS, using default value: %v", err)
		jwtExpirationHours = 24
	}

	// Đọc cấu hình email
	emailVerificationRequiredStr := getEnv("EMAIL_VERIFICATION_REQUIRED", "true")
	emailVerificationRequired, err := strconv.ParseBool(emailVerificationRequiredStr)
	if err != nil {
		log.Printf("Warning: Invalid EMAIL_VERIFICATION_REQUIRED, using default value: %v", err)
		emailVerificationRequired = true
	}

	emailPortStr := getEnv("SMTP_PORT", "1025")
	emailPort, err := strconv.Atoi(emailPortStr)
	if err != nil {
		log.Printf("Warning: Invalid SMTP_PORT, using default value: %v", err)
		emailPort = 1025
	}

	// Đọc cấu hình AppURL
	appURL := getEnv("APP_URL", "http://localhost:8080")

	// Đọc cấu hình CORS
	corsAllowOriginsStr := getEnv("CORS_ALLOW_ORIGINS", "http://localhost:3000")
	corsAllowOrigins := strings.Split(corsAllowOriginsStr, ",")

	return &Config{
		DBHost:                  dbHost,
		DBPort:                  dbPort,
		DBUser:                  dbUser,
		DBPassword:              dbPassword,
		DBName:                  dbName,
		ServerPort:              serverPort,
		JWTSecret:               jwtSecret,
		JWTExpirationHours:      jwtExpirationHours,
		EmailVerificationRequired: emailVerificationRequired,
		EmailPort:               emailPort,
		AppURL:                  appURL,
		CORSAllowOrigins:        corsAllowOrigins,
	}
}

// GetDSN - Lấy connection string cho database
func (c *Config) GetDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName)
}

// getEnv - Lấy giá trị từ biến môi trường hoặc giá trị mặc định
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
} 