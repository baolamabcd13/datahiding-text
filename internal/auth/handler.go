package auth

import (
	"log"
	"net/http"
	"strings"

	"github.com/baolamabcd13/datahiding-text-app/internal/models"
	"github.com/baolamabcd13/datahiding-text-app/internal/utils"
	"github.com/gin-gonic/gin"
)

// Handler - Xử lý HTTP requests cho auth
type Handler struct {
	service Service
}

// NewHandler - Tạo handler mới
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// RegisterRequest - Request body cho đăng ký
type RegisterRequest struct {
	Username        string `json:"username" binding:"required,username"`
	Email           string `json:"email" binding:"required,email"`
	Password        string `json:"password" binding:"required,strongpassword"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=Password"`
	Name            string `json:"name" binding:"required,min=2,max=100,validname"`
	Phone           string `json:"phone" binding:"required,phone"`
	CCCD            string `json:"cccd" binding:"required,cccd"`
}

// RegisterResponse - Response cho đăng ký thành công
type RegisterResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Token    string `json:"token"`
}

// Register - Xử lý đăng ký tài khoản
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithValidationError(c, err)
		return
	}

	// Tạo user mới
	user := &models.User{
		Username: req.Username,
		Password: req.Password,
		Email:    req.Email,
		Name:     req.Name,
		Phone:    req.Phone,
		CCCD:     req.CCCD,
	}

	// Gọi service để đăng ký
	err := h.service.Register(user)
	if err != nil {
		// Thêm log để debug
		log.Printf("Register error: %v", err)
		
		// Kiểm tra lỗi cụ thể
		if strings.Contains(err.Error(), "username already exists") {
			utils.RespondWithError(c, http.StatusConflict, "Username already exists")
			return
		}
		if strings.Contains(err.Error(), "email already exists") {
			utils.RespondWithError(c, http.StatusConflict, "Email already exists")
			return
		}
		if strings.Contains(err.Error(), "cccd already exists") {
			utils.RespondWithError(c, http.StatusConflict, "CCCD already exists")
			return
		}
		
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithSuccess(c, http.StatusCreated, "User registered successfully. Please check your email to verify your account.", nil)
}

// LoginRequest - Request body cho đăng nhập
type LoginRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginResponse - Response cho đăng nhập thành công
type LoginResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Token    string `json:"token"`
}

// Login - Xử lý đăng nhập
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithValidationError(c, err)
		return
	}

	// In ra log để debug
	log.Printf("Login attempt: username=%s", req.Username)

	// Gọi service để đăng nhập
	user, token, err := h.service.Login(req.Username, req.Password)
	if err != nil {
		// In ra log để debug
		log.Printf("Login failed: %v", err)
		
		utils.RespondWithError(c, http.StatusUnauthorized, err.Error())
		return
	}

	// In ra log để debug
	log.Printf("Login successful: username=%s, user_id=%d", user.Username, user.ID)

	// Trả về thông tin user và token
	utils.RespondWithSuccess(c, http.StatusOK, "Login successful", gin.H{
		"token": token,
		"user":  user,
	})
}

// VerifyEmail - Xác thực email
func (h *Handler) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"message": "Token không hợp lệ",
		})
		return
	}

	// Thêm log để debug
	log.Printf("Received verification token: %s", token)

	err := h.service.VerifyEmail(token)
	if err != nil {
		// Thêm log để debug
		log.Printf("Error verifying email: %v", err)
		
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"message": err.Error(),
		})
		return
	}

	c.HTML(http.StatusOK, "verify_success.html", nil)
}

// Logout - Xử lý đăng xuất
func (h *Handler) Logout(c *gin.Context) {
	// Lấy token từ header Authorization
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		utils.RespondWithError(c, http.StatusBadRequest, "authorization header is required")
		return
	}

	// Kiểm tra định dạng Bearer token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		utils.RespondWithError(c, http.StatusBadRequest, "authorization header format must be Bearer {token}")
		return
	}

	// Lấy token
	tokenString := parts[1]
	
	// Đăng xuất
	if err := h.service.Logout(tokenString); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}
	
	// Trả về response
	utils.RespondWithSuccess(c, http.StatusOK, "Logout successful", nil)
}

// ForgotPasswordRequest - Request body cho quên mật khẩu
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest - Request body cho đặt lại mật khẩu
type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8,max=100,password"`
}

// ForgotPassword - Xử lý yêu cầu quên mật khẩu
func (h *Handler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithValidationError(c, err)
		return
	}

	err := h.service.ForgotPassword(req.Email)
	if err != nil {
		// Không tiết lộ thông tin về việc email có tồn tại hay không
		utils.RespondWithSuccess(c, http.StatusOK, "If your email is registered, you will receive a password reset link", nil)
		return
	}

	utils.RespondWithSuccess(c, http.StatusOK, "Password reset link has been sent to your email", nil)
}

// ResetPassword - Xử lý đặt lại mật khẩu
func (h *Handler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithValidationError(c, err)
		return
	}

	err := h.service.ResetPassword(req.Token, req.NewPassword)
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithSuccess(c, http.StatusOK, "Password has been reset successfully", nil)
}

// SetupRoutes - Thiết lập routes cho auth
func (h *Handler) SetupRoutes(router *gin.RouterGroup) {
	auth := router.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.POST("/logout", h.Logout)
		auth.GET("/verify-email", h.VerifyEmail)
		auth.POST("/forgot-password", h.ForgotPassword)
		auth.POST("/reset-password", h.ResetPassword)
	}
}
