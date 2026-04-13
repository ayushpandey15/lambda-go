package response

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type SuccessResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type ErrorResponse struct {
	Status  string `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ValidationField struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationErrorResponse struct {
	Status  string            `json:"status"`
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Errors  []ValidationField `json:"errors"`
}

func WriteSucessResponse(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, SuccessResponse{
		Status:  "success",
		Message: message,
		Data:    data,
	})
}

func WriteErrorResponse(c *gin.Context, err *AppError) {
	c.JSON(err.Status, ErrorResponse{
		Status:  "error",
		Code:    err.Code,
		Message: err.Message,
	})
}

func WriteValidationError(c *gin.Context, err error) {
	fields := parseValidationErrors(err)

	c.JSON(http.StatusBadRequest, ValidationErrorResponse{
		Status:  "error",
		Code:    "INVALID_INPUT",
		Message: "Validation failed",
		Errors:  fields,
	})
}

func parseValidationErrors(err error) []ValidationField {
	var fields []ValidationField

	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, fe := range ve {
			fields = append(fields, ValidationField{
				Field:   fe.Field(),
				Message: formatValidationMessage(fe),
			})
		}
	} else {
		fields = append(fields, ValidationField{
			Field:   "unknown",
			Message: err.Error(),
		})
	}

	return fields
}

func formatValidationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fe.Field())
	case "email":
		return fmt.Sprintf("%s must be a valid email", fe.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s", fe.Field(), fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s", fe.Field(), fe.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of [%s]", fe.Field(), fe.Param())
	case "numeric":
		return fmt.Sprintf("%s must be numeric", fe.Field())
	case "url":
		return fmt.Sprintf("%s must be a valid URL", fe.Field())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", fe.Field(), fe.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", fe.Field(), fe.Param())
	default:
		return fmt.Sprintf("%s failed on %s validation", fe.Field(), fe.Tag())
	}
}
