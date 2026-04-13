package response

import "net/http"

type AppError struct {
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *AppError) Error() string {
	return e.Message
}

func (e *AppError) WithMessage(msg string) *AppError {
	return &AppError{
		Status:  e.Status,
		Code:    e.Code,
		Message: msg,
	}
}

var (
	ErrInternal          = &AppError{http.StatusInternalServerError, "INTERNAL", "Internal server error."}
	ErrUnauthorized      = &AppError{http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized access."}
	ErrNotImplemented    = &AppError{http.StatusNotImplemented, "NOT_IMPLEMENTED", "Resource method not implemented."}
	ErrInvalidInput      = &AppError{http.StatusBadRequest, "INVALID_INPUT", "Invalid input in request."}
	ErrNotFound          = &AppError{http.StatusNotFound, "NOT_FOUND", "No such resource exists."}
	ErrNotAllowed        = &AppError{http.StatusForbidden, "NOT_ALLOWED", "Operation not allowed."}
	ErrNoAccess          = &AppError{http.StatusForbidden, "NO_ACCESS", "Access not allowed."}
	ErrAlreadyExists     = &AppError{http.StatusConflict, "ALREADY_EXISTS", "Resource already exists."}
	ErrSizeLimit         = &AppError{http.StatusRequestEntityTooLarge, "SIZE_LIMIT", "Input size exceeds allowed limits."}
	ErrRateLimit         = &AppError{http.StatusTooManyRequests, "RATE_LIMIT", "Request rate exceeds allowed limits."}
	ErrRateLimitLDAP     = &AppError{http.StatusTooManyRequests, "RATE_LIMIT_LDAP", "Request rate exceeds allowed limits for ldap token."}
	ErrInternalRateLimit = &AppError{http.StatusTooManyRequests, "INTERNAL_RATE_LIMIT_EXCEED", "Please login into matrix or BMS and reload again."}
	ErrOriginNotAllowed  = &AppError{http.StatusForbidden, "ORIGIN_NOT_ALLOWED", "Request from this origin is not allowed, Please contact to admin."}
	ErrDeviceIDNotFound  = &AppError{http.StatusNotFound, "DEVICE_ID_NOT_FOUND", "device uuid not found"}
	ErrSendingEmail      = &AppError{http.StatusNotFound, "ERROR_WHILE_SENDING_EMAIL", "Error while sending email."}
	ErrInvalidAuth       = &AppError{http.StatusForbidden, "INVALID_AUTH", "Invalid Auth."}
	ErrDatabaseTimeout   = &AppError{http.StatusServiceUnavailable, "DATABASE_TIMEOUT", "Unable to connect database."}
	ErrInvalidOTP        = &AppError{http.StatusServiceUnavailable, "INVALID_OTP", "Invalid OTP."}
	ErrSessionExpired    = &AppError{http.StatusUnauthorized, "SESSION_EXPIRED", "Session has been expired."}
	ErrURLNotExists      = &AppError{http.StatusMovedPermanently, "URL_NOT_EXISTS", "Url not exists, please contact saurabhsingh@policybazaar.com."}
	ErrInvalidToken      = &AppError{http.StatusUnauthorized, "UAE_AUTH_INVALID_TOKEN", "Token is invalid."}
)
