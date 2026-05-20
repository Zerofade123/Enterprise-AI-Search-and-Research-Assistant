package errors

import "fmt"

type Code string

const (
	CodeValidation Code = "VALIDATION_ERROR"
	CodeNotFound   Code = "NOT_FOUND"
	CodeConflict   Code = "CONFLICT"
	CodeUnauthorized Code = "UNAUTHORIZED"
	CodeForbidden  Code = "FORBIDDEN"
	CodeInternal   Code = "INTERNAL"
)

type AppError struct {
	Code    Code
	Op      string
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Op, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Op, e.Message)
}

func Wrap(op string, code Code, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Op:      op,
		Message: message,
		Err:     err,
	}
}
