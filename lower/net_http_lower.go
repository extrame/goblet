package lower

import (
	"io"
	"net/http"
	"net/url"
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

func (n *netHTTPLowerReqeust) RemoteAddr() string {
	return n.req.RemoteAddr
}

func (n *netHTTPLowerReqeust) Body() io.Reader {
	return n.req.Body
}

func (n *netHTTPLowerReqeust) URL() *url.URL {
	return n.req.URL
}

type netHTTPLowerWriter http.ResponseWriter
