//+build ignore

package main

import (
	"log"
	"net"

	"github.com/atbys/gabby"
)

func main() {
	e, err := gabby.Default()
	if err != nil {
		log.Fatalf("constructer error")
	}

	dstHW := net.HardwareAddr{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
	dstIP := net.IP{192, 168, 13, 1}
	srcHW, err := net.ParseMAC("f4:5c:89:bf:e1:09")
	if err != nil {
		log.Fatalf("parse error")
	}
	srcIP := net.IP{192, 168, 13, 2}

	err = e.SendRequestARPPacket(dstHW, dstIP, srcHW, srcIP)
	if err != nil {
		log.Fatalf("send error")
	}
}
