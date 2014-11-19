package goblet

import (
	"fmt"
	"testing"
)

func TestAnchor(t *testing.T) {
	anchor := &Anchor{0, "/", "", []*Anchor{}, &HtmlBlockOption{}}
	anchor.add("/", &GroupBlockOption{})
	anchor.add("/sec", &GroupBlockOption{})
	anchor.add("/dec", &HtmlBlockOption{})
	anchor.add("/sec", &GroupBlockOption{})
	anchor.match("/sed/", 5)
	fmt.Println(anchor.match("/deceeee", 4))
}
