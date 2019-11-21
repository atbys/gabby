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

func SampleMiddleware(c *gabby.Context) {
	fmt.Printf("\x1b[31m%s\x1b[0mARP packet get!!\n", "[+]")
	c.Next()
	fmt.Printf("\x1b[31m%s\x1b[0mNext packet waiting...\n", "[+]")
}

func SampleMiddleware2(c *gabby.Context) {
	fmt.Println("-----------------")
	c.Next()
	fmt.Println("+++++++++++++++++")
}
func SampleReplyHandle(c *gabby.Context) {
	fmt.Printf("\x1b[32m%s\x1b[0m Source IP: %s\n", "[Reply]", net.IP(c.Arp.SourceProtAddress).String())
}

func main() {
	e, err := gabby.New()
	if err != nil {
		log.Fatal(err)
	}

	e.Use(SampleMiddleware)
	e.Use(SampleMiddleware2)
	e.Request("ANY", SampleRequestHandle)
	e.Reply("ANY", SampleReplyHandle)

	e.Run()
}
