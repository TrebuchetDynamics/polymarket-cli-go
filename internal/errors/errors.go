package errors

import "fmt"

// Code is a machine-readable error code.
type Code string

const (
	// Transport errors
	CodeTimeout          Code = "NET-001"
	CodeConnectionFailed Code = "NET-002"
	CodeRateLimited      Code = "NET-003"
	CodeCircuitOpen      Code = "NET-004"

	// Auth errors
	CodeMissingSigner    Code = "AUTH-001"
	CodeMissingCreds     Code = "AUTH-002"
	CodeInvalidSignature Code = "AUTH-003"
	CodeUnauthorized     Code = "AUTH-004"

	// CLOB errors
	CodeOrderNotFound     Code = "CLOB-001"
	CodeInsufficientFunds Code = "CLOB-002"
	CodeInvalidOrder      Code = "CLOB-003"
	CodeInvalidTokenID    Code = "CLOB-004"

	// Validation errors
	CodeMissingField    Code = "VAL-001"
	CodeInvalidValue    Code = "VAL-002"
	CodeBatchSizeExceed Code = "VAL-003"

	// Safety errors
	CodeLiveDisabled    Code = "SAFETY-001"
	CodePreflightFailed Code = "SAFETY-002"
	CodeNotAuthorized   Code = "SAFETY-003"

	// Gamma errors
	CodeMarketNotFound Code = "GAMMA-001"
	CodeEventNotFound  Code = "GAMMA-002"
)

// Error wraps a Code with an upstream message.
type Error struct {
	Code     Code   `json:"code"`
	Message  string `json:"message"`
	HTTPCode int    `json:"http_code,omitempty"`
	Err      error  `json:"-"`
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error { return e.Err }

func New(code Code, msg string) *Error {
	return &Error{Code: code, Message: msg}
}

func Wrap(code Code, msg string, err error) *Error {
	return &Error{Code: code, Message: msg, Err: err}
}

func WithHTTP(code Code, msg string, httpCode int) *Error {
	return &Error{Code: code, Message: msg, HTTPCode: httpCode}
}
