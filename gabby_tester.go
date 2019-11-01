//+build ignore

package main

import (
	"fmt"
	"log"

	"github.com/atbys/gabby"
)

func startFunc(name string) {
	fmt.Println("Start system")
}

func getRequestPacket(...interface{}) {
	fmt.Println("Get request packet")
}

func getReplyPacket(...interface{}) {
	fmt.Println("Get reply packet")
}

func main() {
	e, err := gabby.New()
	if err != nil {
		log.Fatal(err)
	}

	e.SetHook(gabby.INIT, startFunc)
	e.SetHook(gabby.REQUEST, getRequestPacket)
	e.SetHook(gabby.REPLY, getReplyPacket)
	e.Run()
}
