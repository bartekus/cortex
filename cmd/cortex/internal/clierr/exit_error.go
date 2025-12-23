package clierr

import (
	"errors"
	"fmt"
)

type ExitCoder interface {
	error
	ExitCode() int
}

// ExitError is an error that carries an explicit process exit code.
// It supports wrapping via Unwrap so errors.Is/As work as expected.
type ExitError struct {
	code  int
	msg   string
	cause error
}

func (e *ExitError) Error() string {
	// Keep this stable and user-facing; don't include code here.
	// Include cause only if present, in a deterministic way.
	if e.cause == nil {
		return e.msg
	}
	return fmt.Sprintf("%s: %v", e.msg, e.cause)
}

func (e *ExitError) ExitCode() int { return e.code }

// Unwrap enables errors.Is/As to traverse the underlying cause.
func (e *ExitError) Unwrap() error { return e.cause }

// Code returns the exit code (optional convenience; keeps fields private).
func (e *ExitError) Code() int { return e.code }

// Message returns the top-level message (optional convenience).
func (e *ExitError) Message() string { return e.msg }

// New creates an ExitError with a message.
func New(code int, msg string) error {
	return &ExitError{code: normalize(code), msg: msg}
}

// Wrap creates an ExitError that wraps an underlying cause.
func Wrap(code int, msg string, cause error) error {
	if cause == nil {
		return New(code, msg)
	}
	return &ExitError{code: normalize(code), msg: msg, cause: cause}
}

// Newf is a formatted variant.
func Newf(code int, format string, args ...any) error {
	return &ExitError{code: normalize(code), msg: fmt.Sprintf(format, args...)}
}

// Wrapf is a formatted variant that wraps.
func Wrapf(code int, cause error, format string, args ...any) error {
	msg := fmt.Sprintf(format, args...)
	return Wrap(code, msg, cause)
}

// ExitCodeOf extracts an exit code from any error, defaulting to 1.
// This keeps main() dumb and avoids duplicating errors.As logic everywhere.
func ExitCodeOf(err error) int {
	if err == nil {
		return 0
	}
	var ec ExitCoder
	if errors.As(err, &ec) {
		return ec.ExitCode()
	}
	return 1
}

func normalize(code int) int {
	// Exit code 0 means success; errors should never be 0.
	if code <= 0 {
		return 1
	}
	return code
}
