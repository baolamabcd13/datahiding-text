package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/baolamabcd13/datahiding-text-app/internal/auth"
	"github.com/baolamabcd13/datahiding-text-app/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// AuthMiddleware - Middleware xác thực JWT
func AuthMiddleware(jwtSecret string, tokenRepo auth.TokenRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Lấy token từ header Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.RespondWithError(c, http.StatusUnauthorized, "authorization header is required")
			c.Abort()
			return
		}

		// Kiểm tra định dạng Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.RespondWithError(c, http.StatusUnauthorized, "authorization header format must be Bearer {token}")
			c.Abort()
			return
		}

		// Parse token
		tokenString := parts[1]
		
		// Kiểm tra xem token có trong blacklist không
		isBlacklisted, err := tokenRepo.IsBlacklisted(tokenString)
		if err != nil {
			utils.RespondWithError(c, http.StatusInternalServerError, "failed to validate token")
			c.Abort()
			return
		}
		
		if isBlacklisted {
			utils.RespondWithError(c, http.StatusUnauthorized, "token has been revoked")
			c.Abort()
			return
		}
		
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Kiểm tra thuật toán ký
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(jwtSecret), nil
		})

		if err != nil {
			utils.RespondWithError(c, http.StatusUnauthorized, "invalid token")
			c.Abort()
			return
		}

		// Kiểm tra token hợp lệ
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Lưu user_id vào context
			userID := uint(claims["user_id"].(float64))
			c.Set("user_id", userID)
			c.Next()
		} else {
			utils.RespondWithError(c, http.StatusUnauthorized, "invalid token")
			c.Abort()
			return
		}
	}
}