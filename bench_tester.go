//+build ignore

package main

import (
	//"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/atbys/gabby"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func main() {
	srcHWAddr, err := net.ParseMAC("00-00-00-00-00-00")
	srcIP := net.ParseIP("192.168.13.1")
	dstHWAddr, err := net.ParseMAC("ff-ff-ff-ff-ff-ff")
	dstIP := net.ParseIP("192.169.13.89")

	eth := layers.Ethernet{
		SrcMAC:       srcHWAddr,
		DstMAC:       dstHWAddr,
		EthernetType: layers.EthernetTypeARP,
	}

	arp := layers.ARP{
		AddrType:          layers.LinkTypeEthernet,
		Protocol:          layers.EthernetTypeIPv4,
		HwAddressSize:     6,
		ProtAddressSize:   4,
		Operation:         layers.ARPRequest,
		SourceHwAddress:   []byte(srcHWAddr),
		SourceProtAddress: []byte(srcIP.To4()),
		DstHwAddress:      []byte(dstHWAddr),
		DstProtAddress:    []byte(dstIP.To4()),
	}

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}
	err = gopacket.SerializeLayers(buf, opts, &eth, &arp)
	if err != nil {
		log.Fatal(err)
	}

	p := gopacket.NewPacket(buf.Bytes(), layers.LinkTypeEthernet, gopacket.Default)

	e, err := gabby.Default()
	if err != nil {
		log.Fatal(err)
	}
	e.Request("ANY", func(c *gabby.Context){})
	e.Reply("ANY", func(c *gabby.Context){})

	packets := make(chan gopacket.Packet, 1024)
	go e.HandleManager(packets)

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)

	for i := 1; i < 256; i++ {
		for j:=1; j < 256; j++{
			arp.DstProtAddress = net.IP{192, 168, byte(j), byte(i)}.To4()
			err = gopacket.SerializeLayers(buf, opts, &eth, &arp)
			if err != nil {
				log.Fatal(err)
			}
			p = gopacket.NewPacket(buf.Bytes(), layers.LinkTypeEthernet, gopacket.Default)
			packets <- p
		}
	}

	<-quit
}
