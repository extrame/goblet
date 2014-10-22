package goblet

import (
	"fmt"
	"testing"
)

func TestAnchor(t *testing.T) {
	anchor := &Anchor{0, "/", "", []*Anchor{}, &HtmlBlockOption{}}
	anchor.add("/", &CommonBlokOption{})
	anchor.add("/sec", &CommonBlokOption{})
	anchor.add("/dec", &HtmlBlockOption{})
	anchor.add("/sic", &CommonBlokOption{})
	anchor.match("/sed/", 5)
	fmt.Println(anchor.match("/deceeee", 4))
}
