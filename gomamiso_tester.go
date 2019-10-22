//+build ignore

package main

import (
	"fmt"
	"github.com/atbys/gomamiso"
)

func main() {
	e := gomamiso.New()
	e.SetDevice("en0")
	fmt.Println(e.Device)
}
