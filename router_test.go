package goblet

import (
	"fmt"
	"testing"
)

func TestAnchor(t *testing.T) {
	anchor := &Anchor{0, "/", "", []*Anchor{}, &HtmlBlockOption{}}
	anchor.add("/", &GroupBlockOption{})
	anchor.add("/sed", &GroupBlockOption{})
	anchor.add("/sc", &HtmlBlockOption{})
	anchor.add("/sec", &HtmlBlockOption{})
	a, _ := anchor.match("/se", 3)
	fmt.Printf("%T,%v\n", a.opt, a.opt)
	b, _ := anchor.match("/sec", 4)
	fmt.Printf("%T,%v\n", b.opt, b.opt)
}

func TestAnchorShort(t *testing.T) {
	anchor := &Anchor{0, "/", "", []*Anchor{}, &HtmlBlockOption{}}
	anchor.add("/", &_staticBlockOption{})
	fmt.Println(anchor)
	anchor.add("/seeed", &GroupBlockOption{})
	fmt.Println(anchor)
	anchor.add("/sec", &HtmlBlockOption{})
	fmt.Println(anchor)
	a, _ := anchor.match("/sec", 4)
	fmt.Printf("%T,%v\n", a.opt, a.opt)
}

func TestAnchorShortAndSame(t *testing.T) {
	anchor := &Anchor{0, "/", "", []*Anchor{}, &HtmlBlockOption{}}
	anchor.add("/", &_staticBlockOption{})
	fmt.Println(anchor)
	anchor.add("/see", &HtmlBlockOption{})
	fmt.Println(anchor)
	anchor.add("/seeed", &GroupBlockOption{})
	fmt.Println(anchor)
	a, _ := anchor.match("/see", 4)
	fmt.Printf("%T,%v\n", a.opt, a.opt)
}
