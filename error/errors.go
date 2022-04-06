package ge

const (
	ERROR_NOSUCHROUTER = iota + 10000
	ERROR_CheckedAndStillNotExists
)

type Error struct {
	Method string
	Code   int
}

func (e *Error) Error() string {
	switch e.Code {
	case ERROR_NOSUCHROUTER:
		return "NOSUCHROUTER"
	}
	return "Unknown Error"
}

func NOSUCHROUTER(method string) error {
	return &Error{
		Code:   ERROR_NOSUCHROUTER,
		Method: method,
	}
}

func CheckedAndStillNotExists(method string) error {
	return &Error{
		Code:   ERROR_CheckedAndStillNotExists,
		Method: method,
	}
}

func IsNoSuchRouter(err error) bool {
	myE, ok := err.(*Error)
	return ok && myE.Code == ERROR_NOSUCHROUTER
}

func IsChecked(err error) bool {
	myE, ok := err.(*Error)
	return ok && myE.Code == ERROR_CheckedAndStillNotExists
}
