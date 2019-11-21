//+build ignore

package main

import (
	"fmt"
	"log"
	"net"

	"github.com/atbys/gabby"
)

func SampleRequestHandle(c *gabby.Context) {
	fmt.Print("REQ")
	fmt.Println(net.IP(c.Arp.SourceProtAddress).String())
}

func SampleRequestHandle2(c *gabby.Context) {
	fmt.Println("Sampleeeeeeeeeeeeeeeeeee")
}
func SampleReplyHandle(c *gabby.Context) {
	fmt.Print("REP")
	fmt.Println(net.IP(c.Arp.SourceProtAddress).String())
}

func main() {
	e, err := gabby.New()
	if err != nil {
		log.Fatal(err)
	}

	e.Request("ANY", SampleRequestHandle)
	e.Request("ANY", SampleRequestHandle2)
	e.Reply("ANY", SampleReplyHandle)
	e.Run()
}
