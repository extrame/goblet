package goblet

import (
	"log"
	"runtime/debug"
)

//Use for defer unsafe go runtime
func SafeGo() {
	if err := recover(); err != nil {
		log.Printf("%T,%v", err, err)
		log.Print(string(debug.Stack()))
	}
}
