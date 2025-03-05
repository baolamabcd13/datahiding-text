package middleware

import (
	"strings"

	"github.com/baolamabcd13/datahiding-text-app/internal/utils"
	"github.com/baolamabcd13/datahiding-text-app/internal/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ValidationMiddleware handles validation errors in a consistent way
func ValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any validation errors
		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				// Check if it's a validator error
				if validationErrors, ok := e.Err.(validator.ValidationErrors); ok {
					errorMessages := make(map[string]string)
					for _, err := range validationErrors {
						fieldName := err.Field()
						// Convert the first letter to lowercase for JSON field names
						if len(fieldName) > 0 {
							fieldName = strings.ToLower(fieldName[:1]) + fieldName[1:]
						}
						errorMessages[fieldName] = err.Translate(validation.GetTranslator())
					}
					
					c.AbortWithStatusJSON(400, utils.ValidationErrorResponse{
						Status:  "error",
						Message: "Validation failed",
						Errors:  errorMessages,
					})
					return
				}
			}
		}
	}
} 