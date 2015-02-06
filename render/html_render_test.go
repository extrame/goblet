package render

import (
	"fmt"
	"testing"
)

func TestSlice(t *testing.T) {
	slice := []string{"123", "234", "345", "456", "567"}
	fmt.Print(Slice(slice, 2))
}
