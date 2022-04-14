package file

// ErrAborted should be returned whenever a file operation is terminated
// before it has completed.
type AbortError struct {
	err error
}

// NewErrAbort creates a new ErrAborted instance.
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
