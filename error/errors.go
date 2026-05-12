package ge

const (
	ERROR_NOSUCHROUTER = iota + 10000
	ERROR_FallbackWhenNoStaicFile
	ERROR_CheckedAndStillNotExists = 20000
)

type Error struct {
	Method   string
	Code     int
	fallback *fallbackWhenNoStaicFile
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

func NewFallbackWhenNoStaicFile(file string) *Error {
	return &Error{
		Code:   ERROR_FallbackWhenNoStaicFile,
		Method: file,
	}
}

type fallbackWhenNoStaicFile struct {
	Index  string
	Render string
	Layout string
}
