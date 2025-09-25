package errors

type InternalInterruptedError struct {
	reason string
}

func (e *InternalInterruptedError) Error() string {
	return e.reason
}

// Interrupted creates a new interrupted error with custom message, interrupted error will make
// goblet stop processing and return the error to client, like in any Pre function
func Interrupted(reason string) error {
	return &InternalInterruptedError{
		reason: reason,
	}
}
