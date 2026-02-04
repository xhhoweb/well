package apperr

import "errors"

// Business Error Codes
const (
	CodeSuccess         = 0
	CodeBadRequest      = 400
	CodeUnauthorized    = 401
	CodeForbidden       = 403
	CodeNotFound        = 404
	CodeInternalError   = 500
	CodeDatabaseError   = 1001
	CodeCacheError      = 1002
	CodeThreadNotFound  = 2001
	CodeThreadCreateErr = 2002
	CodeThreadUpdateErr = 2003
	CodeThreadDeleteErr = 2004
)

// Business Errors
var (
	ErrThreadNotFound  = errors.New("thread not found")
	ErrThreadCreateErr = errors.New("thread create error")
	ErrThreadUpdateErr = errors.New("thread update error")
	ErrThreadDeleteErr = errors.New("thread delete error")
	ErrInvalidParams   = errors.New("invalid parameters")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrForbidden       = errors.New("forbidden")
)

// AppError Application Error with code and message
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *AppError) Error() string {
	return e.Message
}

// NewAppError Create new application error
func NewAppError(code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// WrapError Wrap error with code
func WrapError(err error, code int) *AppError {
	if err == nil {
		return nil
	}
	if ae, ok := err.(*AppError); ok {
		return ae
	}
	return &AppError{
		Code:    code,
		Message: err.Error(),
	}
}
