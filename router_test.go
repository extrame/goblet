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
	anchor.add("/sec", &HtmlBlockOption{})
	a, _ := anchor.match("/end", 4)
	fmt.Printf("%T,%v\n", a.opt, a.opt)
}
