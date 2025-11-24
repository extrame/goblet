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
	a, _, _ := anchor.Match("/stat/days/2018-04-19.json", 3)
	if a.Opt.String() != "Group(root) with ignore case" {
		t.Error("Group not match, got ", a.Opt.String())
	}
	//match 15 letters so days will be matched
	c, _, _ := anchor.Match("/stat/days/2018-04-19.json", 15)
	if c.Opt.String() != "Group(days) with ignore case" {
		t.Error("Group not match, got ", c.Opt.String())
	}
	b, _, _ := anchor.Match("/sec", 4)
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
	a, _, _ := anchor.Match("/sec", 4)
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
	a, _, _ := anchor.Match("/see", 4)
	t.Logf("%T,%v\n", a.Opt, a.Opt)
}

func TestAnchorOfTwoRest(t *testing.T) {
	anchor := matcher.Root(ctrl.NewStatic())
	anchor.Add("/", &ctrl.Static{})
	anchor.Add("/first", &ctrl.Rest{&ctrl.Basic{Name: "first"}})
	anchor.Add("/first/second", &ctrl.Group{&ctrl.Basic{Name: "second"}, true})
	anchor.Add("/first/three", &ctrl.Rest{&ctrl.Basic{Name: "three"}})
	a, _, _ := anchor.Match("/first/2/tag", 11)
	if a.Opt.String() != "Rest(first)" {
		t.Error("Rest not match, got ", a.Opt.String())
	}
	b, _, _ := anchor.Match("/first/second/222", 17)
	if b.Opt.String() != "Group(second) with ignore case" {
		t.Error("Group not match, got ", b.Opt.String())
	}
	c, _, _ := anchor.Match("/first/three", 12)
	if c.Opt.String() != "Rest(three)" {
		t.Error("Rest not match, got ", c.Opt.String())
	}
}

func TestAnchorOfOneParamsRest(t *testing.T) {
	anchor := matcher.Root(ctrl.NewStatic())
	anchor.Add("/", &ctrl.Static{})
	anchor.Add("/first", &ctrl.Rest{&ctrl.Basic{Name: "first"}})
	anchor.Add("/first/second", &ctrl.Group{&ctrl.Basic{Name: "second"}, true})
	anchor.Add("/first/three", &ctrl.Rest{&ctrl.Basic{Name: "three"}})
	anchor.Add("/four/:param1/three", &ctrl.Rest{&ctrl.Basic{Name: "four"}})
	a, _, _ := anchor.Match("/first/2/tag", 11)
	if a.Opt.String() != "Rest(first)" {
		t.Error("Rest not match, got ", a.Opt.String())
	}
	b, _, _ := anchor.Match("/first/second/222", 17)
	if b.Opt.String() != "Group(second) with ignore case" {
		t.Error("Group not match, got ", b.Opt.String())
	}
	c, _, _ := anchor.Match("/first/three", 12)
	if c.Opt.String() != "Rest(three)" {
		t.Error("Rest not match, got ", c.Opt.String())
	}
	d, _, params := anchor.Match("/four/test/three", 16)
	if d == nil {
		t.Error("Rest not match, got nil")
	} else if d.Opt == nil {
		t.Error("Rest not match, got nil")
	} else if d.Opt.String() != "Rest(four)" {
		t.Error("Rest not match, got ", d.Opt.String())
	} else if params["param1"] != "test" {
		t.Error("params not match, got ", params["param1"])
	}
}

func TestAnchorOfTwoParamsRest(t *testing.T) {
	anchor := matcher.Root(ctrl.NewStatic())
	anchor.Add("/", &ctrl.Static{})
	anchor.Add("/first", &ctrl.Rest{&ctrl.Basic{Name: "first"}})
	anchor.Add("/first/second", &ctrl.Group{&ctrl.Basic{Name: "second"}, true})
	anchor.Add("/first/three", &ctrl.Rest{&ctrl.Basic{Name: "three"}})
	anchor.Add("/four/:param1/three", &ctrl.Rest{&ctrl.Basic{Name: "four"}})
	anchor.Add("/four/:param1/five/:param2", &ctrl.Rest{&ctrl.Basic{Name: "five"}})
	a, _, _ := anchor.Match("/first/2/tag", 11)
	if a.Opt.String() != "Rest(first)" {
		t.Error("Rest not match, got ", a.Opt.String())
	}
	b, _, _ := anchor.Match("/first/second/222", 17)
	if b.Opt.String() != "Group(second) with ignore case" {
		t.Error("Group not match, got ", b.Opt.String())
	}
	c, _, _ := anchor.Match("/first/three", 12)
	if c.Opt.String() != "Rest(three)" {
		t.Error("Rest not match, got ", c.Opt.String())
	}
	d, _, params := anchor.Match("/four/test/three", 16)
	if d.Opt.String() != "Rest(four)" {
		t.Error("Rest not match, got ", d.Opt.String())
	}
	if params["param1"] != "test" {
		t.Error("params not match, got ", params["param1"])
	}
	e, _, params := anchor.Match("/four/test/five/123", 20)
	if e.Opt.String() != "Rest(five)" {
		t.Error("Rest not match, got ", e.Opt.String())
	}
	if params["param1"] != "test" || params["param2"] != "123" {
		t.Error("params not match, got ", params)
	}
}

func TestAnchorOfTwoParamsWithTailRest(t *testing.T) {
	anchor := matcher.Root(ctrl.NewStatic())
	anchor.Add("/", &ctrl.Static{})
	anchor.Add("/first", &ctrl.Rest{&ctrl.Basic{Name: "first"}})
	anchor.Add("/first/second", &ctrl.Group{&ctrl.Basic{Name: "second"}, true})
	anchor.Add("/first/three", &ctrl.Rest{&ctrl.Basic{Name: "three"}})
	anchor.Add("/four/:param1/three", &ctrl.Rest{&ctrl.Basic{Name: "four"}})
	anchor.Add("/four/:param1/five/:param2", &ctrl.Rest{&ctrl.Basic{Name: "five"}})
	a, _, _ := anchor.Match("/first/2/tag", 11)
	if a.Opt.String() != "Rest(first)" {
		t.Error("Rest not match, got ", a.Opt.String())
	}
	b, _, _ := anchor.Match("/first/second/222", 17)
	if b.Opt.String() != "Group(second) with ignore case" {
		t.Error("Group not match, got ", b.Opt.String())
	}
	c, _, _ := anchor.Match("/first/three", 12)
	if c.Opt.String() != "Rest(three)" {
		t.Error("Rest not match, got ", c.Opt.String())
	}
	d, _, params := anchor.Match("/four/test/three", 16)
	if d.Opt.String() != "Rest(four)" {
		t.Error("Rest not match, got ", d.Opt.String())
	}
	if params["param1"] != "test" {
		t.Error("params not match, got ", params["param1"])
	}
	e, _, params := anchor.Match("/four/test/five/123", 20)
	if e.Opt.String() != "Rest(five)" {
		t.Error("Rest not match, got ", e.Opt.String())
	}
	if params["param1"] != "test" || params["param2"] != "123" {
		t.Error("params not match, got ", params)
	}
	e, _, params = anchor.Match("/four/test/five/123/six", 20)
	if e.Opt.String() != "Rest(five)" {
		t.Error("Rest not match, got ", e.Opt.String())
	}
	if params["param1"] != "test" || params["param2"] != "123" {
		t.Error("params not match, got ", params)
	}
}
