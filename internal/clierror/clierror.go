package clierror

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/sailpoint-oss/sailpoint-cli/internal/redact"
)

const (
	ExitGeneral  = 1
	ExitUsage    = 2
	ExitNotFound = 3
	ExitAPI      = 4
	ExitCanceled = 130
)

type Error struct {
	Message  string
	Hint     string
	Code     int
	Category string
}

func (e *Error) Error() string {
	if e.Hint == "" {
		return e.Message
	}
	return fmt.Sprintf("%s\nHint: %s", e.Message, e.Hint)
}

func New(message string, code int) *Error {
	return &Error{Message: message, Code: code}
}

func Usage(message string) *Error {
	return &Error{Message: message, Code: ExitUsage, Category: "usage"}
}

func NotFound(resource, name, hint string) *Error {
	return &Error{
		Message:  fmt.Sprintf("%s not found: %s", resource, name),
		Hint:     hint,
		Code:     ExitNotFound,
		Category: "not_found",
	}
}

func Canceled(action string) *Error {
	if action == "" {
		action = "operation"
	}
	return &Error{
		Message:  action + " canceled",
		Code:     ExitCanceled,
		Category: "canceled",
	}
}

func APIStatus(statusCode int, status string, body []byte) *Error {
	message := status
	if message == "" {
		message = http.StatusText(statusCode)
	}
	if len(body) > 0 {
		message = fmt.Sprintf("%s: %s", message, summarizeBody(body))
	}
	return &Error{
		Message:  "API request failed with status " + message,
		Code:     ExitAPI,
		Category: "api",
	}
}

func ExitCode(err error) int {
	if err == nil {
		return 0
	}

	var cliErr *Error
	if errors.As(err, &cliErr) && cliErr.Code != 0 {
		return cliErr.Code
	}

	return ExitGeneral
}

func summarizeBody(body []byte) string {
	summary := strings.TrimSpace(redact.Bytes(body))
	if len(summary) > 500 {
		return summary[:500] + "..."
	}
	return summary
}
