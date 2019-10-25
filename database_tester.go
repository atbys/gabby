//+build ignore

package main

import (
	"log"

	"github.com/atbys/gomamiso"
)

func main() {
	engine := gomamiso.Default()
	err := engine.Init()
	if err != nil {
		log.Fatal(err)
	}

	column := gomamiso.Column{
		"ipaddr":    "'192.168.3.1'",
		"macaddr":   "'11-11-11-11-11-11'",
		"timestamp": "current_timestamp",
	}
	engine.Insert(column)

	err = engine.ShowDB()
	if err != nil {
		log.Fatal(err)
	}

	err = engine.AddColumn("status")
	if err != nil {
		log.Fatal(err)
	}
}
