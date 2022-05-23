package plugin

import (
	"fmt"
	"testing"
)

type TestedError struct {
	JsonError
	Msg string
}

func TestJsonErr(t *testing.T) {
	var e TestedError
	renderJsonError(e)
	var pe = &e
	renderJsonError(pe)
}

func renderJsonError(e JsonErrorRender) {
	fmt.Println(e, e.RespondAsJson())
}
