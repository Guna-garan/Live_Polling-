// Package apperror defines the application's standard error codes and the
// JSON envelope every API error response uses.
package apperror

import "net/http"

// Code is a stable, machine-readable error identifier returned to clients.
type Code string

const (
	CodeInvalidRequest     Code = "INVALID_REQUEST"
	CodeUnauthorized       Code = "UNAUTHORIZED"
	CodeForbidden          Code = "FORBIDDEN"
	CodePollNotFound       Code = "POLL_NOT_FOUND"
	CodePollClosed         Code = "POLL_CLOSED"
	CodePollExpired        Code = "POLL_EXPIRED"
	CodeInvalidOption      Code = "INVALID_OPTION"
	CodeAlreadyVoted       Code = "ALREADY_VOTED"
	CodeUserExists         Code = "USER_EXISTS"
	CodeInvalidCredentials Code = "INVALID_CREDENTIALS"
	CodeRateLimited        Code = "RATE_LIMITED"
	CodeInternalError      Code = "INTERNAL_ERROR"
)

// AppError is an error with an associated HTTP status and public code/message.
// Internal details (wrapped errors) are logged server-side but never leaked
// to the client.
type AppError struct {
	Status  int
	Code    Code
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error { return e.Err }

func New(status int, code Code, message string) *AppError {
	return &AppError{Status: status, Code: code, Message: message}
}

func Wrap(status int, code Code, message string, err error) *AppError {
	return &AppError{Status: status, Code: code, Message: message, Err: err}
}

// Envelope is the JSON body shape for every error response.
type Envelope struct {
	Error EnvelopeBody `json:"error"`
}

type EnvelopeBody struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
}

func (e *AppError) Envelope() Envelope {
	return Envelope{Error: EnvelopeBody{Code: e.Code, Message: e.Message}}
}

// Internal returns a standard 500 that never leaks the underlying error text.
func Internal(err error) *AppError {
	return Wrap(http.StatusInternalServerError, CodeInternalError, "something went wrong", err)
}
