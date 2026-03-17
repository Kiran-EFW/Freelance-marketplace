package validator

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

// validate is the singleton validator instance.
var validate *validator.Validate

func init() {
	validate = validator.New()
}

// ValidationError represents a single field validation error.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// Error implements the error interface.
func (v ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", v.Field, v.Message)
}

// Validate validates a struct using struct tags (e.g., `validate:"required"`).
func Validate(s interface{}) error {
	return validate.Struct(s)
}

// phone regex: E.164 format (+country_code followed by digits, 7-15 digits total).
var phoneRegex = regexp.MustCompile(`^\+[1-9]\d{6,14}$`)

// ValidatePhone validates that a phone number is in E.164 format.
func ValidatePhone(phone string) error {
	if phone == "" {
		return ValidationError{
			Field:   "phone",
			Message: "phone number is required",
			Code:    "required",
		}
	}

	if !phoneRegex.MatchString(phone) {
		return ValidationError{
			Field:   "phone",
			Message: "phone number must be in E.164 format (e.g., +919876543210)",
			Code:    "invalid_format",
		}
	}

	return nil
}

// postcodeRegex validates an Indian pincode (6 digits, first digit non-zero).
var postcodeRegex = regexp.MustCompile(`^[1-9]\d{5}$`)

// ValidatePostcode validates a postcode for the given jurisdiction. Currently
// supports Indian postcodes (6-digit PINs).
func ValidatePostcode(postcode string, jurisdictionID string) error {
	if postcode == "" {
		return ValidationError{
			Field:   "postcode",
			Message: "postcode is required",
			Code:    "required",
		}
	}

	postcode = strings.TrimSpace(postcode)

	// Default to Indian postcode validation.
	_ = jurisdictionID // reserved for future multi-jurisdiction support

	if !postcodeRegex.MatchString(postcode) {
		return ValidationError{
			Field:   "postcode",
			Message: "postcode must be a valid 6-digit Indian PIN code",
			Code:    "invalid_format",
		}
	}

	return nil
}

// emailRegex is a simplified email validation regex.
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// ValidateEmail validates an email address format.
func ValidateEmail(email string) error {
	if email == "" {
		return ValidationError{
			Field:   "email",
			Message: "email is required",
			Code:    "required",
		}
	}

	email = strings.TrimSpace(strings.ToLower(email))

	if !emailRegex.MatchString(email) {
		return ValidationError{
			Field:   "email",
			Message: "email address is not valid",
			Code:    "invalid_format",
		}
	}

	return nil
}

// ValidateRating validates that a rating is between 1 and 5 (inclusive).
func ValidateRating(rating int) error {
	if rating < 1 || rating > 5 {
		return ValidationError{
			Field:   "rating",
			Message: "rating must be between 1 and 5",
			Code:    "out_of_range",
		}
	}

	return nil
}

// FormatValidationErrors converts a validator.ValidationErrors into a slice
// of ValidationError for consistent API responses.
func FormatValidationErrors(err error) []ValidationError {
	if err == nil {
		return nil
	}

	var result []ValidationError

	// Handle go-playground/validator errors.
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fe := range validationErrors {
			ve := ValidationError{
				Field: toSnakeCase(fe.Field()),
				Code:  fe.Tag(),
			}

			switch fe.Tag() {
			case "required":
				ve.Message = fmt.Sprintf("%s is required", toSnakeCase(fe.Field()))
			case "email":
				ve.Message = "must be a valid email address"
			case "min":
				ve.Message = fmt.Sprintf("must be at least %s characters", fe.Param())
			case "max":
				ve.Message = fmt.Sprintf("must be at most %s characters", fe.Param())
			case "oneof":
				ve.Message = fmt.Sprintf("must be one of: %s", fe.Param())
			case "uuid":
				ve.Message = "must be a valid UUID"
			case "url":
				ve.Message = "must be a valid URL"
			default:
				ve.Message = fmt.Sprintf("failed validation: %s", fe.Tag())
			}

			result = append(result, ve)
		}
		return result
	}

	// Handle our own ValidationError type.
	if ve, ok := err.(ValidationError); ok {
		return []ValidationError{ve}
	}

	// Fallback for unknown error types.
	result = append(result, ValidationError{
		Field:   "unknown",
		Message: err.Error(),
		Code:    "validation_failed",
	})

	return result
}

// toSnakeCase converts a PascalCase or camelCase field name to snake_case.
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		if r >= 'A' && r <= 'Z' {
			result.WriteRune(r + 32) // lowercase
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}
