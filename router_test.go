package goblet

import (
	"fmt"
	"testing"

	"github.com/extrame/goblet/internal/ctrl"
	"github.com/extrame/goblet/internal/matcher"
)

func TestAnchor(t *testing.T) {
	anchor := matcher.Root(ctrl.NewStatic())
	anchor.Add("/", &ctrl.Group{&ctrl.Basic{Name: "root"}, true})
	anchor.Add("/stat/days", &ctrl.Group{&ctrl.Basic{Name: "days"}, true})
	anchor.Add("/sc", &ctrl.Html{})
	anchor.Add("/sec", &ctrl.Html{&ctrl.Basic{Name: "right"}})
	//match 3 letters so root will be matched
	a, _ := anchor.Match("/stat/days/2018-04-19.json", 3)
	if a.Opt.String() != "Group(root) with ignore case" {
		t.Error("Group not match, got ", a.Opt.String())
	}
	//match 15 letters so days will be matched
	c, _ := anchor.Match("/stat/days/2018-04-19.json", 15)
	if c.Opt.String() != "Group(days) with ignore case" {
		t.Error("Group not match, got ", c.Opt.String())
	}
	b, _ := anchor.Match("/sec", 4)
	if b.Opt.String() != "Html(right)" {
		t.Error("Html not match, got ", b.Opt.String())
	}
	t.Logf("%T,%v\n", b.Opt, b.Opt)
}

func TestAnchorShort(t *testing.T) {
	anchor := matcher.Root(ctrl.NewStatic())
	anchor.Add("/", &ctrl.Static{})
	fmt.Println(anchor)
	anchor.Add("/seeed", &ctrl.Group{})
	fmt.Println(anchor)
	anchor.Add("/sec", &ctrl.Html{&ctrl.Basic{Name: "right"}})
	fmt.Println(anchor)
	a, _ := anchor.Match("/sec", 4)
	t.Logf("%T,%v\n", a.Opt, a.Opt)
}

func TestAnchorShortAndSame(t *testing.T) {
	anchor := matcher.Root(ctrl.NewStatic())
	anchor.Add("/", &ctrl.Static{})
	fmt.Println(anchor)
	anchor.Add("/see", &ctrl.Html{&ctrl.Basic{Name: "right"}})
	fmt.Println(anchor)
	anchor.Add("/seeed", &ctrl.Group{})
	fmt.Println(anchor)
	a, _ := anchor.Match("/see", 4)
	t.Logf("%T,%v\n", a.Opt, a.Opt)
}

func TestAnchorOfTwoRest(t *testing.T) {
	anchor := matcher.Root(ctrl.NewStatic())
	anchor.Add("/", &ctrl.Static{})
	anchor.Add("/first", &ctrl.Rest{&ctrl.Basic{Name: "first"}})
	anchor.Add("/first/second", &ctrl.Group{&ctrl.Basic{Name: "second"}, true})
	anchor.Add("/first/three", &ctrl.Rest{&ctrl.Basic{Name: "three"}})
	a, _ := anchor.Match("/first/2/tag", 11)
	t.Logf("%T,%v\n", a.Opt, a.Opt)
	b, _ := anchor.Match("/first/second/222", 17)
	t.Logf("%T,%v\n", b.Opt, b.Opt)
	c, _ := anchor.Match("/first/three", 12)
	t.Logf("%T,%v\n", c.Opt, c.Opt)
}
