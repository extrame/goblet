package goblet

import (
	"fmt"
	"testing"
)

type ShouldHidden struct {
	Hidden   string `goblet:"hide"`
	NoHidden string
}

func TestHide(t *testing.T) {
	var tested ShouldHidden
	tested.Hidden = "1"
	tested.NoHidden = "2"
	fmt.Println(autoHide(&tested, nil))
}

func TestHideSlice(t *testing.T) {
	var tested ShouldHidden
	var testedArr = []ShouldHidden{tested}
	tested.Hidden = "1"
	tested.NoHidden = "2"
	fmt.Println(autoHide(testedArr, nil))
}

func TestHideMatcher(t *testing.T) {
	fmt.Println(hideMatcher.FindStringSubmatch("hide(/1/2/3,/4/5/6)"))
	fmt.Println(hideMatcher.MatchString("hide(/1/2/3,/4/5/6)"))
	fmt.Println(hideMatcher.FindStringSubmatch("hide/1/2/3,/4/5/6)"))
	fmt.Println(hideMatcher.MatchString("hide/1/2/3,/4/5/6)"))
	fmt.Println(hideMatcher.FindAllStringSubmatch("hide(/1/2/3)", -1))
	fmt.Println(hideMatcher.FindAllStringSubmatch("hide", -1))
	fmt.Println(hideMatcher.MatchString("hide"))
}
