//+build ignore

package main

import (
	"github.com/atbys/gomamiso"
)

func main(){
	engine := gomamiso.Default()
	engine.Init()
	defer engine.Deinit()

	engine.Insert()
}