package goblet

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/extrame/goblet/v2/internal/errors"
)

// 错误包装函数
func errorWrap(w http.ResponseWriter) {
	if e := recover(); e != nil {
		log.Print("panic:", e, "\n", string(debug.Stack()))
		w.WriteHeader(http.StatusInternalServerError)
		if err, ok := e.(error); ok {
			w.Write([]byte(err.Error()))
		}
	}
}

// Interrupted creates a new interrupted error with custom message, interrupted error will make
// goblet stop processing and return the error to client, like in any Pre function
func Interrupted(reason string) error {
	return &errors.InternalInterruptedError{
		Reason: reason,
	}
}
