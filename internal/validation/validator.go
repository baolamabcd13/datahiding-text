package validation

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

// Global variables for validator and translator
var (
	uni      *ut.UniversalTranslator
	trans    ut.Translator
	validate *validator.Validate
)

// Initialize sets up the validator with custom validations and translations
func Initialize() {
	// Get validator instance from Gin
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		validate = v

		// Initialize translator
		enLocale := en.New()
		uni = ut.New(enLocale, enLocale)
		trans, _ = uni.GetTranslator("en")

		// Register default English translations
		en_translations.RegisterDefaultTranslations(validate, trans)

		// Register custom validations
		registerCustomValidations()

		// Register custom translations
		registerCustomTranslations()

		// Register tag name function to use json tag names in error messages
		validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return fld.Name
			}
			return name
		})
	}
}

// registerCustomValidations registers custom validation functions
func registerCustomValidations() {
	// Validate Vietnamese phone number
	_ = validate.RegisterValidation("phone", func(fl validator.FieldLevel) bool {
		phone := fl.Field().String()
		// Vietnamese phone number format: +84xxxxxxxxx or 0xxxxxxxxx (9-11 digits)
		re := regexp.MustCompile(`^(\+84|0)[3|5|7|8|9][0-9]{8,9}$`)
		return re.MatchString(phone)
	})

	// Validate Vietnamese CCCD (Citizen Identity Card)
	_ = validate.RegisterValidation("cccd", func(fl validator.FieldLevel) bool {
		cccd := fl.Field().String()
		// CCCD format: 12 digits
		re := regexp.MustCompile(`^[0-9]{12}$`)
		return re.MatchString(cccd)
	})

	// Validate username format
	_ = validate.RegisterValidation("username", func(fl validator.FieldLevel) bool {
		username := fl.Field().String()
		// Username format: 3-20 characters, alphanumeric and underscore
		re := regexp.MustCompile(`^[a-zA-Z0-9_]{3,20}$`)
		return re.MatchString(username)
	})

	// Validate strong password
	_ = validate.RegisterValidation("strongpassword", func(fl validator.FieldLevel) bool {
		password := fl.Field().String()
		// Password must be at least 8 characters long
		if len(password) < 8 {
			return false
		}

		var (
			hasUpper   bool
			hasLower   bool
			hasNumber  bool
			hasSpecial bool
		)

		for _, char := range password {
			switch {
			case unicode.IsUpper(char):
				hasUpper = true
			case unicode.IsLower(char):
				hasLower = true
			case unicode.IsNumber(char):
				hasNumber = true
			case unicode.IsPunct(char) || unicode.IsSymbol(char):
				hasSpecial = true
			}
		}

		// Password must contain at least 3 of the 4 character types
		return (hasUpper && hasLower && hasNumber) || 
		       (hasUpper && hasLower && hasSpecial) || 
		       (hasUpper && hasNumber && hasSpecial) || 
		       (hasLower && hasNumber && hasSpecial)
	})

	// Validate name format (no numbers or special characters)
	_ = validate.RegisterValidation("validname", func(fl validator.FieldLevel) bool {
		name := fl.Field().String()
		// Name should only contain letters, spaces, and some special characters like hyphen or apostrophe
		re := regexp.MustCompile(`^[a-zA-ZÀ-ỹ\s\-'\.]+$`)
		return re.MatchString(name)
	})

	// Validate URL format
	_ = validate.RegisterValidation("url", func(fl validator.FieldLevel) bool {
		url := fl.Field().String()
		// Simple URL validation
		re := regexp.MustCompile(`^(http|https)://[a-zA-Z0-9\-\.]+\.[a-zA-Z]{2,}(\/\S*)?$`)
		return url == "" || re.MatchString(url) // Empty URL is also valid
	})
}

// registerCustomTranslations registers custom error messages for validations
func registerCustomTranslations() {
	// Phone validation message
	registerTranslation("phone", "{0} must be a valid Vietnamese phone number")

	// CCCD validation message
	registerTranslation("cccd", "{0} must be a valid 12-digit Citizen Identity Card number")

	// Username validation message
	registerTranslation("username", "{0} must be 3-20 characters long and can only contain letters, numbers, and underscores")

	// Strong password validation message
	registerTranslation("strongpassword", "{0} must be at least 8 characters long and contain at least one uppercase letter, one lowercase letter, and one number")

	// Valid name validation message
	registerTranslation("validname", "{0} must contain only letters, spaces, and characters like hyphen or apostrophe")

	// URL validation message
	registerTranslation("url", "{0} must be a valid URL or empty")

	// Required field message
	registerTranslation("required", "{0} is required")

	// Email validation message
	registerTranslation("email", "{0} must be a valid email address")

	// Min length message
	registerTranslation("min", "{0} must be at least {1} characters long")

	// Max length message
	registerTranslation("max", "{0} must not exceed {1} characters")

	// EqField message (for password confirmation)
	registerTranslation("eqfield", "{0} must match {1}")
}

// registerTranslation is a helper function to register custom error messages
func registerTranslation(tag string, message string) {
	_ = validate.RegisterTranslation(tag, trans, func(ut ut.Translator) error {
		return ut.Add(tag, message, true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T(tag, fe.Field(), fe.Param())
		return t
	})
}

// ValidateStruct validates a struct and returns formatted error messages
func ValidateStruct(s interface{}) (bool, map[string]string) {
	err := validate.Struct(s)
	if err == nil {
		return true, nil
	}

	// Convert validation errors to a map of field:error message
	errors := make(map[string]string)
	
	for _, err := range err.(validator.ValidationErrors) {
		fieldName := err.Field()
		// Convert the first letter to lowercase for JSON field names
		if len(fieldName) > 0 {
			fieldName = strings.ToLower(fieldName[:1]) + fieldName[1:]
		}
		errors[fieldName] = err.Translate(trans)
	}

	return false, errors
}

// FormatError formats a single validation error
func FormatError(field string, tag string, param string) string {
	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", field, param)
	case "max":
		return fmt.Sprintf("%s must not exceed %s characters", field, param)
	case "eqfield":
		return fmt.Sprintf("%s must match %s", field, param)
	case "phone":
		return fmt.Sprintf("%s must be a valid Vietnamese phone number", field)
	case "cccd":
		return fmt.Sprintf("%s must be a valid 12-digit Citizen Identity Card number", field)
	case "username":
		return fmt.Sprintf("%s must be 3-20 characters long and can only contain letters, numbers, and underscores", field)
	case "strongpassword":
		return fmt.Sprintf("%s must be at least 8 characters long and contain at least one uppercase letter, one lowercase letter, and one number", field)
	case "validname":
		return fmt.Sprintf("%s must contain only letters, spaces, and characters like hyphen or apostrophe", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL or empty", field)
	default:
		return fmt.Sprintf("%s failed validation for tag %s with param %s", field, tag, param)
	}
}

// GetTranslator returns the translator instance
func GetTranslator() ut.Translator {
	return trans
} 