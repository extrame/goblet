package goblet

import (
	"fmt"
	"net/http"
	"runtime/debug"
)

func WrapError(w http.ResponseWriter, err interface{}, withStack bool) {
	w.WriteHeader(500)
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
