//+build ignore

package main

import (
	"fmt"
	"log"
	"net"

	"github.com/atbys/gabby"
)

func SampleRequestHandle(c *gabby.Context) {
	fmt.Printf("\x1b[32m%s\x1b[0m Source IP: %s\n", "[Request]", net.IP(c.Arp.SourceProtAddress).String())
}

func SampleReplyHandle(c *gabby.Context) {
	fmt.Printf("\x1b[32m%s\x1b[0m Source IP: %s\n", "[Reply]", net.IP(c.Arp.SourceProtAddress).String())
}

func main() {
	e, err := gabby.Default()
	if err != nil {
		log.Fatal(err)
	}

	e.Request("ANY", SampleRequestHandle)
	e.Reply("ANY", SampleReplyHandle)

	e.Run()
}
