package utils

import (
	"net/http"
	"strings"

	"github.com/baolamabcd13/datahiding-text-app/internal/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ErrorResponse - Cấu trúc response cho lỗi
type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// ValidationErrorResponse - Cấu trúc response cho lỗi validation
type ValidationErrorResponse struct {
	Status  string            `json:"status"`
	Message string            `json:"message"`
	Errors  map[string]string `json:"errors"`
}

// SuccessResponse - Cấu trúc response cho thành công
type SuccessResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// RespondWithError - Trả về response lỗi
func RespondWithError(c *gin.Context, code int, message string) {
	c.JSON(code, ErrorResponse{
		Status:  "error",
		Message: message,
	})
}

// RespondWithValidationError - Trả về response lỗi validation
func RespondWithValidationError(c *gin.Context, err error) {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		errorMessages := make(map[string]string)
		for _, err := range validationErrors {
			fieldName := err.Field()
			// Convert the first letter to lowercase for JSON field names
			if len(fieldName) > 0 {
				fieldName = strings.ToLower(fieldName[:1]) + fieldName[1:]
			}
			errorMessages[fieldName] = err.Translate(validation.GetTranslator())
		}
		
		c.JSON(http.StatusBadRequest, ValidationErrorResponse{
			Status:  "error",
			Message: "Validation failed",
			Errors:  errorMessages,
		})
		return
	}
	
	// If it's not a validation error, return a generic error
	RespondWithError(c, http.StatusBadRequest, err.Error())
}

// RespondWithSuccess - Trả về response thành công
func RespondWithSuccess(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(code, SuccessResponse{
		Status:  "success",
		Message: message,
		Data:    data,
	})
} 