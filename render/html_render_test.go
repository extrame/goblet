package render

import (
	"fmt"
	"html/template"
	"os"
	"testing"
)

func TestSlice(t *testing.T) {
	slice := []string{"123", "234", "345", "456", "567"}
	fmt.Print(Slice(slice, 2))
}

func TestRepeat(t *testing.T) {
	fmt.Print(Repeat(2))
}

func TestModel(t *testing.T) {
	var root = ` {{ template "test" }}`
	rt, _ := template.New("").Parse(root)
	ra, _ := rt.Clone()
	var a = `hello from a`
	ra.New("test").Parse(a)

	rb, _ := rt.Clone()
	var b = `hello from b`
	rb.New("test").Parse(b)

	ra.Execute(os.Stdout, nil)
	rb.Execute(os.Stdout, nil)
}
