package user

import (
	"net/http"

	"github.com/baolamabcd13/datahiding-text-app/internal/utils"
	"github.com/gin-gonic/gin"
)

// Handler - Xử lý HTTP requests cho user
type Handler struct {
	service Service
}

// NewHandler - Tạo handler mới
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// ProfileResponse - Response cho profile
type ProfileResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	CCCD     string `json:"cccd"`
	Avatar   string `json:"avatar"`
}

// GetProfile - Lấy thông tin profile của người dùng hiện tại
func (h *Handler) GetProfile(c *gin.Context) {
	// Lấy user_id từ context (đã được set bởi middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Lấy thông tin người dùng
	user, err := h.service.GetUserByID(userID.(uint))
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	if user == nil {
		utils.RespondWithError(c, http.StatusNotFound, "user not found")
		return
	}

	// Trả về response
	utils.RespondWithSuccess(c, http.StatusOK, "Profile retrieved successfully", ProfileResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Name:     user.Name,
		Phone:    user.Phone,
		CCCD:     user.CCCD,
		Avatar:   user.Avatar,
	})
}

// UpdateProfileRequest - Request body cho cập nhật profile
type UpdateProfileRequest struct {
	Username string `json:"username" binding:"omitempty,username"`
	Name     string `json:"name" binding:"omitempty,min=2,max=100,validname"`
	Phone    string `json:"phone" binding:"omitempty,phone"`
	Avatar   string `json:"avatar" binding:"omitempty,url"`
}

// UpdateProfile - Cập nhật thông tin profile
func (h *Handler) UpdateProfile(c *gin.Context) {
	// Lấy user_id từ context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Parse request body
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithValidationError(c, err)
		return
	}

	// Lấy thông tin người dùng hiện tại
	user, err := h.service.GetUserByID(userID.(uint))
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	if user == nil {
		utils.RespondWithError(c, http.StatusNotFound, "user not found")
		return
	}

	// Cập nhật thông tin nếu được cung cấp
	if req.Username != "" {
		// Kiểm tra xem username đã tồn tại chưa nếu khác username hiện tại
		if req.Username != user.Username {
			existingUser, err := h.service.GetUserByUsername(req.Username)
			if err != nil {
				utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
				return
			}
			if existingUser != nil {
				utils.RespondWithError(c, http.StatusBadRequest, "username already exists")
				return
			}
		}
		user.Username = req.Username
	}
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}

	// Lưu thông tin
	if err := h.service.UpdateUser(user); err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Trả về response
	utils.RespondWithSuccess(c, http.StatusOK, "Profile updated successfully", ProfileResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Name:     user.Name,
		Phone:    user.Phone,
		CCCD:     user.CCCD,
		Avatar:   user.Avatar,
	})
}

// SetupRoutes - Thiết lập routes cho user
func (h *Handler) SetupRoutes(router *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	users := router.Group("/users")
	{
		// Routes cần xác thực
		users.Use(authMiddleware)
		users.GET("/me", h.GetProfile)
		users.PUT("/me", h.UpdateProfile)
	}
}