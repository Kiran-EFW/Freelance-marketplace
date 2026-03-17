package response

import (
	"github.com/gofiber/fiber/v2"

	"github.com/seva-platform/backend/pkg/validator"
)

// SuccessResponse is the standard envelope for successful API responses.
type SuccessResponse struct {
	Data interface{} `json:"data"`
}

// PaginatedResponse is the envelope for paginated API responses.
type PaginatedResponse struct {
	Data interface{}    `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

// PaginationMeta holds pagination metadata.
type PaginationMeta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// ErrorResponse is the standard envelope for API error responses.
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail holds error information.
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ValidationErrorResponse is the envelope for validation error responses.
type ValidationErrorResponse struct {
	Error  ErrorDetail                `json:"error"`
	Errors []validator.ValidationError `json:"errors"`
}

// Success sends a 200 OK response with the given data.
func Success(c *fiber.Ctx, data interface{}) error {
	return c.Status(fiber.StatusOK).JSON(SuccessResponse{
		Data: data,
	})
}

// Created sends a 201 Created response with the given data.
func Created(c *fiber.Ctx, data interface{}) error {
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Data: data,
	})
}

// Paginated sends a 200 OK response with data and pagination metadata.
func Paginated(c *fiber.Ctx, data interface{}, page, limit, total int) error {
	totalPages := 0
	if limit > 0 {
		totalPages = (total + limit - 1) / limit
	}

	return c.Status(fiber.StatusOK).JSON(PaginatedResponse{
		Data: data,
		Meta: PaginationMeta{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

// Error sends an error response with the given HTTP status, error code, and
// human-readable message.
func Error(c *fiber.Ctx, status int, code string, message string) error {
	return c.Status(status).JSON(ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}

// ValidationError sends a 422 Unprocessable Entity response with field-level
// validation errors.
func ValidationError(c *fiber.Ctx, errors []validator.ValidationError) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(ValidationErrorResponse{
		Error: ErrorDetail{
			Code:    "validation_error",
			Message: "one or more fields failed validation",
		},
		Errors: errors,
	})
}

// NotFound sends a 404 Not Found response for the named resource.
func NotFound(c *fiber.Ctx, resource string) error {
	return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
		Error: ErrorDetail{
			Code:    "not_found",
			Message: resource + " not found",
		},
	})
}

// Unauthorized sends a 401 Unauthorized response.
func Unauthorized(c *fiber.Ctx) error {
	return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
		Error: ErrorDetail{
			Code:    "unauthorized",
			Message: "authentication required",
		},
	})
}

// Forbidden sends a 403 Forbidden response.
func Forbidden(c *fiber.Ctx) error {
	return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{
		Error: ErrorDetail{
			Code:    "forbidden",
			Message: "you do not have permission to perform this action",
		},
	})
}
