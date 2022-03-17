package presenters

import "fmt"

// HTTPError represents an HTTP error that could be wrapping another error, it includes an HTTP code to be sent
type HTTPError struct {
	orig error
	msg  string
	code int
}

// WrapErrorf returns a wrapped error
func WrapErrorf(orig error, httpStatusCode int, format string, a ...interface{}) error {
	return &HTTPError{
		code: httpStatusCode,
		orig: orig,
		msg:  fmt.Sprintf(format, a...),
	}
}

// NewErrorf instantiates a new error
func NewErrorf(httpStatusCode int, format string, a ...interface{}) error {
	return WrapErrorf(nil, httpStatusCode, format, a...)
}

// Error returns the message, when wrapping errors the wrapped error is returned
func (e *HTTPError) Error() string {
	if e.orig != nil {
		return fmt.Sprintf("%s: %v", e.msg, e.orig)
	}

	return e.msg
}

// Unwrap returns the wrapped error, if any
func (e *HTTPError) Unwrap() error {
	return e.orig
}

// Code returns the code representing this error
func (e *HTTPError) Code() int {
	return e.code
}
