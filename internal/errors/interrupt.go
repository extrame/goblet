package errors

type InternalInterruptedError struct {
	Reason string
}

func (e *InternalInterruptedError) Error() string {
	return e.Reason
}
