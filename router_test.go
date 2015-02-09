package goblet

import (
	"fmt"
	"testing"
)

func TestAnchor(t *testing.T) {
	anchor := &Anchor{0, "/", "", []*Anchor{}, &HtmlBlockOption{}}
	anchor.add("/", &GroupBlockOption{})
	anchor.add("/sed", &GroupBlockOption{})
	anchor.add("/dec", &HtmlBlockOption{})
	anchor.add("/sec", &HtmlBlockOption{})
	a, _ := anchor.match("/sed", 4)
	fmt.Printf("%T,%v\n", a.opt, a.opt)
	b, _ := anchor.match("/sec", 4)
	fmt.Printf("%T,%v\n", b.opt, b.opt)
}
