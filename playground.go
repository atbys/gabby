//+build ignore

package main

import (
	"fmt"
	"log"
	"time"

	"github.com/google/gopacket/layers"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

var (
	device       string = "enp0s31f6"
	snapshot_len int32  = 60
	promiscuous  bool   = false
	err          error
	timeout      time.Duration = 30 * time.Second
	handle       *pcap.Handle
)

func main() {
	// Open device
	handle, err = pcap.OpenLive(device, snapshot_len, promiscuous, timeout)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	// Use the handle as a packet source to process all packets
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		// Process packet here
		arplayer := packet.Layer(layers.LayerTypeARP)
		if arplayer == nil {
			continue
		}

		arp := *arplayer.(*layers.ARP)

		if arp.Operation == layers.ARPRequest {
			fmt.Printf("[ARP Request] IP: %v MAC: %v\n", arp.SourceProtAddress, arp.SourceHwAddress)
		} else if arp.Operation == layers.ARPReply {
			fmt.Printf("[ARP Reply] IP: %v MAC: %v\n", arp.SourceProtAddress, arp.SourceHwAddress)

		}
	}
}
