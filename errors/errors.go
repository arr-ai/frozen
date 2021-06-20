package errors

import "github.com/go-errors/errors"

// Wrap errors, returning a nil error if errors.Wrap returns a nil *Error.
func Wrap(e interface{}, skip int) error { // nolint:unparam
	// nolint:revive
	if err := errors.Wrap(e, skip+1); err != nil {
		return err
	}
	return nil
}

type InternalError string

func (e InternalError) Error() string {
	return string(e)
}

const (
	// WTF is panicked from code that should never be reached.
	WTF = InternalError("should never be called!")

	// Unimplemented is panicked from functions that aren't implemented yet.
	// They shouldn't happened outside frozen development.
	Unimplemented = InternalError("not implemented")
)
