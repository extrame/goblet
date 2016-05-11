package lower

import (
	"net/http"
)

type netHTTPLowerReqeust struct {
	req    *http.Request
	writer *http.ResponseWriter
}

func (n *netHTTPLowerReqeust) ReqMethod() string {
	return n.req.Method
}

func (n *netHTTPLowerReqeust) ReqReferer() string {
	return n.req.Referer()
}

func (n *netHTTPLowerReqeust) FormValue(key string) string {
	return n.req.FormValue(key)
}

func (n *netHTTPLowerReqeust) QueryValue(key string) string {
	return n.req.URL.Query().Get(key)
}

type netHTTPLowerWriter http.ResponseWriter
