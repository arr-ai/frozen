package errors

import "github.com/go-errors/errors"

// Wrap wraps errors. Translates backend nil *Error to nil error.
func Wrap(e any, skip int) error {
	if err := errors.Wrap(e, skip+1); err != nil { //nolint:revive
		return err
	}
	return nil
}

// WrapPrefix wraps errors. Translates backend nil *Error to nil error.
func WrapPrefix(e any, prefix string, skip int) error {
	if err := errors.WrapPrefix(e, prefix, skip+1); err != nil { //noline:revive
		return err
	}
	return nil
}

func Errorf(format string, args ...any) error {
	return errors.Errorf(format, args...)
}

type InternalError string

func (e InternalError) Error() string {
	return string(e)
}

const (
	// ErrWTF is panicked from code that should never be reached.
	ErrWTF = InternalError("should never be called!")

	// ErrUnimplemented is panicked from functions that aren't implemented yet.
	// They shouldn't happened outside frozen development.
	ErrUnimplemented = InternalError("not implemented")

	// ErrConsistencyCheck is panicked when an internal consistency check fails.
	ErrConsistencyCheck = InternalError("consistency check failure")
)
