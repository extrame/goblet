package goblet

import (
	"github.com/valyala/fasthttp"
	"log"
	"net/http"
	"runtime/debug"
)

func ErrorWrap(ctx *fasthttp.RequestCtx) {
	if e := recover(); e != nil {
		log.Print("panic:", e, "\n", string(debug.Stack()))
		ctx.SetStatusCode(http.StatusInternalServerError)
		if err, ok := e.(error); ok {
			ctx.Write([]byte(err.Error()))
		}
	}
}
