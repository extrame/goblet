package goblet

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"runtime/debug"
)

func WrapError(w *fasthttp.RequestCtx, err interface{}, withStack bool) {
	w.SetStatusCode(500)
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
