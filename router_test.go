package goblet

import (
	"fmt"
	"testing"

	"github.com/extrame/goblet/internal/ctrl"
)

func TestAnchor(t *testing.T) {
	anchor := &anchor{0, "/", "", []*anchor{}, &ctrl.Html{}}
	anchor.add("/", &ctrl.Group{&ctrl.Basic{Name: "root"}, true})
	anchor.add("/stat/days", &ctrl.Group{&ctrl.Basic{Name: "days"}, true})
	anchor.add("/sc", &ctrl.Html{})
	anchor.add("/sec", &ctrl.Html{&ctrl.Basic{Name: "right"}})
	//match 3 letters so root will be matched
	a, _ := anchor.match("/stat/days/2018-04-19.json", 3)
	if a.opt.String() != "Group(root) with ignore case" {
		t.Error("Group not match, got ", a.opt.String())
	}
	//match 15 letters so days will be matched
	c, _ := anchor.match("/stat/days/2018-04-19.json", 15)
	if c.opt.String() != "Group(days) with ignore case" {
		t.Error("Group not match, got ", c.opt.String())
	}
	b, _ := anchor.match("/sec", 4)
	if b.opt.String() != "Html(right)" {
		t.Error("Html not match, got ", b.opt.String())
	}
	t.Logf("%T,%v\n", b.opt, b.opt)
}

func TestAnchorShort(t *testing.T) {
	anchor := &anchor{0, "/", "", []*anchor{}, &ctrl.Html{}}
	anchor.add("/", &ctrl.Static{})
	fmt.Println(anchor)
	anchor.add("/seeed", &ctrl.Group{})
	fmt.Println(anchor)
	anchor.add("/sec", &ctrl.Html{&ctrl.Basic{Name: "right"}})
	fmt.Println(anchor)
	a, _ := anchor.match("/sec", 4)
	t.Logf("%T,%v\n", a.opt, a.opt)
}

func TestAnchorShortAndSame(t *testing.T) {
	anchor := &anchor{0, "/", "", []*anchor{}, &ctrl.Html{}}
	anchor.add("/", &ctrl.Static{})
	fmt.Println(anchor)
	anchor.add("/see", &ctrl.Html{&ctrl.Basic{Name: "right"}})
	fmt.Println(anchor)
	anchor.add("/seeed", &ctrl.Group{})
	fmt.Println(anchor)
	a, _ := anchor.match("/see", 4)
	t.Logf("%T,%v\n", a.opt, a.opt)
}

func TestAnchorOfTwoRest(t *testing.T) {
	anchor := &anchor{0, "/", "", []*anchor{}, &ctrl.Html{}}
	anchor.add("/", &ctrl.Static{})
	anchor.add("/first", &ctrl.Rest{&ctrl.Basic{Name: "first"}})
	anchor.add("/first/second", &ctrl.Group{&ctrl.Basic{Name: "second"}, true})
	anchor.add("/first/three", &ctrl.Rest{&ctrl.Basic{Name: "three"}})
	a, _ := anchor.match("/first/2/tag", 11)
	t.Logf("%T,%v\n", a.opt, a.opt)
	b, _ := anchor.match("/first/second/222", 17)
	t.Logf("%T,%v\n", b.opt, b.opt)
	c, _ := anchor.match("/first/three", 12)
	t.Logf("%T,%v\n", c.opt, c.opt)
}
