package errors

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
