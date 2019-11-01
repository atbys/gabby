//+build ignore

package main

import (
	"fmt"
	"log"

	"github.com/atbys/gomamiso"
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
	e, err := gomamiso.New()
	if err != nil {
		log.Fatal(err)
	}

	e.SetHook(gomamiso.INIT, startFunc)
	e.SetHook(gomamiso.REQUEST, getRequestPacket)
	e.SetHook(gomamiso.REPLY, getReplyPacket)
	err = e.Run()
	if err != nil {
		log.Fatal(err)
	}
}
