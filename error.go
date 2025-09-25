package goblet

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"

	"log/slog"

	gorandom "github.com/extrame/go-random"
	"github.com/extrame/goblet/config"
)

func (s *Server) wrapError(w http.ResponseWriter, err interface{}, withStack bool) {
	if withStack {
		w.WriteHeader(500)
	}
	if s.Env() == config.ProductEnv {
		errKey := gorandom.RandomNumeric(10)
		slog.Error("Error happened",
			"error", err,
			"key", errKey,
			"type", fmt.Sprintf("%T", err))
		if withStack {

			log.Print(string(debug.Stack()))
		}
		html := fmt.Sprintf(`<body><h4>Internal Error(%s)</h4><br/>The Random Key is %s</body>`, errKey, errKey)
		w.Write([]byte(html))
	} else {
		w.Write([]byte("<body><h4>"))
		w.Write([]byte(fmt.Sprintf("%T,%v", err, err)))
		w.Write([]byte("</h4>"))
		if withStack {
			w.Write([]byte("<pre>"))
			w.Write([]byte(debug.Stack()))
			w.Write([]byte("</pre>"))
		}
		w.Write([]byte("</body>"))
	}
}

type internalInterruptedError struct {
	reason string
}

func (e *internalInterruptedError) Error() string {
	return e.reason
}

// Interrupted creates a new interrupted error with custom message, interrupted error will make
// goblet stop processing and return the error to client, like in any Pre function
func Interrupted(reason string) error {
	return &internalInterruptedError{
		reason: reason,
	}
}

func (ctx *Context) WrapError(err error, info string) error {
	return fmt.Errorf("[%s]err:%s", info, err)
}
