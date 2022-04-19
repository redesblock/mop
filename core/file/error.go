package file

// AbortError should be returned whenever a file operation is terminated
// before it has completed.
type AbortError struct {
	err error
}

// NewAbortError creates a new AbortError instance.
func NewAbortError(err error) error {
	return &AbortError{
		err: err,
	}
}

// Unwrap returns an underlying error.
func (e *AbortError) Unwrap() error {
	return e.err
}

// Error implement standard go error interface.
func (e *AbortError) Error() string {
	return e.err.Error()
}

// HashError should be returned whenever a file operation is terminated
// before it has completed.
type HashError struct {
	err error
}

// NewHashError creates a new HashError instance.
func NewHashError(err error) error {
	return &HashError{
		err: err,
	}
}

// Unwrap returns an underlying error.
func (e *HashError) Unwrap() error {
	return e.err
}

// Error implement standard go error interface.
func (e *HashError) Error() string {
	return e.err.Error()
}
