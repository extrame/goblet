package goblet

import (
	"log"
	"net/http"
	"runtime/debug"
)

//错误包装函数
func errorWrap(w http.ResponseWriter) {
	if e := recover(); e != nil {
		log.Print("panic:", e, "\n", string(debug.Stack()))
		w.WriteHeader(http.StatusInternalServerError)
		if err, ok := e.(error); ok {
			w.Write([]byte(err.Error()))
		}
	}
}
