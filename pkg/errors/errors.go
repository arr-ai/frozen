package errors

import "github.com/go-errors/errors"

// Wrap wraps errors. Translates backend nil *Error to nil error.
func Wrap(e any, skip int) error {
	if err := errors.Wrap(e, skip+1); err != nil {
		return err
	}
	return nil
}

// WrapPrefix wraps errors. Translates backend nil *Error to nil error.
func WrapPrefix(e any, prefix string, skip int) error {
	if err := errors.WrapPrefix(e, prefix, skip+1); err != nil {
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
	// WTF is panicked from code that should never be reached.
	WTF = InternalError("should never be called!")

	// Unimplemented is panicked from functions that aren't implemented yet.
	// They shouldn't happened outside frozen development.
	Unimplemented = InternalError("not implemented")

	// ConsistencyCheck is panicked when an internal consistency check fails.
	ConsistencyCheck = InternalError("consistency check failure")
)
