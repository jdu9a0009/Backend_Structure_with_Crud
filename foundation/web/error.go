package web

// FieldError is used to indicate an error with a specific request field.
type FieldError struct {
	Field string `json:"field"`
	Error string `json:"error"`
}

// ErrorResponse is the form area for API responses from failures in the API.
type ErrorResponse struct {
	Error  string       `json:"error"`
	Fields []FieldError `json:"fields,omitempty"`
	Data   interface{}  `json:"data"`
	Status bool         `json:"status"`
}

// MobileErrorResponse is the form area for API responses from failures in the API.
type MobileErrorResponse struct {
	Error  interface{} `json:"error"`
	Data   interface{} `json:"data"`
	Status bool        `json:"status"`
}

// Error used to pass an error during the request through the
// application with web specific context
type Error struct {
	Err    error
	Status int
	Fields []FieldError
}

// Error implements the error interface. It uses the default message of the
// wrapped error. This is what will be shown in the services' logs.
func (e *Error) Error() string {
	return e.Err.Error()
}

// NewRequestError wraps a provided error with an HTTP status code. This
// function should be used when handlers encounter expected errors.
func NewRequestError(err error, status int) error {
	return &Error{err, status, nil}
}

// shutdown is type used to help with the graceful termination of the service.
type shutdown struct {
	Message string
}

// Error is the implementation of the error interface.
func (s *shutdown) Error() string {
	return s.Message
}

// NewShutdownError returns an error that causes the framework to signal
// a graceful shutdown.
func NewShutdownError(msg string) error {
	return &shutdown{Message: msg}
}

// IsShutdown checks to see if the shutdown error is contained
// in the specified error value.
func IsShutdown(err error) bool {
	if _, ok := Cause(err).(*shutdown); ok {
		return true
	}
	return false
}

// Cause returns the underlying cause of the error, if possible.
// An error value has a cause if it implements the following
// interface:
//
//	type causer interface {
//	       Cause() error
//	}
//
// If the error does not implement Cause, the original error will
// be returned. If the error is nil, nil will be returned without further
// investigation.
func Cause(err error) error {
	type causer interface {
		Cause() error
	}

	for err != nil {
		cause, ok := err.(causer)
		if !ok {
			break
		}
		err = cause.Cause()
	}
	return err
}
