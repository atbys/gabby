//+build ignore

package main

import (
	"log"

	"github.com/atbys/gabby"
)

func main() {
	engine, err := gabby.New()
	if err != nil {
		log.Fatal(err)
	}

	column := gabby.Column{
		"ipaddr":    "'192.168.3.1'",
		"macaddr":   "'11-11-11-11-11-11'",
		"timestamp": "current_timestamp",
	}
	engine.DB.Insert(column)

	err = engine.DB.ShowDB()
	if err != nil {
		log.Fatal(err)
	}
}
