package gabby

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

type Handle func(*Context)

type Handlers []Handle

type Engine struct {
	DeviceName      string
	DeviceHWAddr    net.HardwareAddr
	snapshotLen     int32
	promiscuous     bool
	timeout         time.Duration
	phandle         *pcap.Handle
	RequestHandlers map[string]Handlers
	ReplyHandlers   map[string]Handlers
	DB              Database
	Config          Config
}

func New() (*Engine, error) {
	e := &Engine{
		snapshotLen:     1024,
		promiscuous:     true,
		timeout:         30 * time.Second,
		RequestHandlers: make(map[string]Handlers),
		ReplyHandlers:   make(map[string]Handlers),
	}

	err := e.Init()
	if err != nil {
		return nil, err
	}

	return e, nil
}

func (self *Engine) Init() error {
	err := self.readConfig()
	if err != nil {
		return err
	}

	//Open device
	if self.Config.Device.Name == "" {
		log.Println("Please Set Device in below")
		FindDevice()
		return errors.New("cannot open device")
	}

	self.DeviceName = self.Config.Device.Name

	self.phandle, err = pcap.OpenLive(self.DeviceName, self.snapshotLen, self.promiscuous, self.timeout)
	if err != nil {
		return err
	}

	return nil
}

//entry
func (self *Engine) Run() {
	packetSource := gopacket.NewPacketSource(self.phandle, self.phandle.LinkType())
	packets := packetSource.Packets()
	defer close(packets)

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)

	fmt.Println("START")
	go self.HandleManager(packets)

	<-quit

	fmt.Println("[Interrupt] Exit")
}

func (self *Engine) HandleManager(packets chan gopacket.Packet) {
	processingHandle := make(map[string]*Context)
	myHWaddr, _ := net.ParseMAC(self.Config.Device.Hwaddr)
	result := make(chan Result)
	for {
		select {
		case packet := <-packets:
			// Process packet here
			arpLayer := packet.Layer(layers.LayerTypeARP)
			if arpLayer == nil {
				continue
			}
			arp := arpLayer.(*layers.ARP)
			if bytes.Equal([]byte(myHWaddr), arp.SourceHwAddress) {
				// This is a packet I sent.
				continue
			}
			// log.Printf("IP %v is at %v", net.IP(arp.SourceProtAddress), net.HardwareAddr(arp.SourceHwAddress))

			srcIP := net.IP(arp.SourceProtAddress).String()
			dstIP := net.IP(arp.DstProtAddress).String()

			_, ok := processingHandle[dstIP]
			if ok {
				continue
			}

			c := &Context{
				Arp:    arp,
				index:  0,
				Engine: self,
				Result: result,
			}

			processingHandle[srcIP] = c

			var isRequest = arp.Operation == layers.ARPRequest
			var isReply = arp.Operation == layers.ARPReply

			if isRequest {
				handlers, ok := self.RequestHandlers[srcIP]
				if ok {
					c.handlers = handlers
				} else {
					c.handlers, ok = self.RequestHandlers["ANY"]
				}
			} else if isReply {
				handlers, ok := self.ReplyHandlers[srcIP]
				if ok {
					c.handlers = handlers
				} else {
					c.handlers, ok = self.ReplyHandlers["ANY"]
				}

				c, ok := processingHandle[dstIP]
				if ok {
					c.receiveReply <- struct{}{}
				}
			} else {
				log.Fatal(srcIP)
			}

			go c.Start()

		case r := <-result:
			delete(processingHandle, r.addr)
			fmt.Println(processingHandle)
		}
	}
}

func (self *Engine) Request(addr string, handle Handle) {
	self.RequestHandlers[addr] = append(self.RequestHandlers[addr], handle)
}

func (self *Engine) Reply(addr string, handle Handle) {
	self.ReplyHandlers[addr] = append(self.ReplyHandlers[addr], handle)
}

func (self *Engine) Use(middleware Handle) {
	self.RequestHandlers["ANY"] = append(self.RequestHandlers["ANY"], middleware)
	self.ReplyHandlers["ANY"] = append(self.ReplyHandlers["ANY"], middleware)
}

func (self *Engine) SendRequestARPPacket(dstHWAddr net.HardwareAddr, dstIP net.IP, srcHWAddr net.HardwareAddr, srcIP net.IP) error {
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
		SourceProtAddress: []byte(srcIP),
		DstHwAddress:      []byte(dstHWAddr),
		DstProtAddress:    []byte(dstIP),
	}

	buf := gopacket.NewSerializeBuffer()

	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}

	err := gopacket.SerializeLayers(buf, opts, &eth, &arp)
	if err != nil {
		return err
	}
	err = self.phandle.WritePacketData(buf.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func (self *Engine) SendReplyARPPacket(dstHWAddr net.HardwareAddr, dstIP net.IP, srcHWAddr net.HardwareAddr, srcIP net.IP) error {
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
		Operation:         layers.ARPReply,
		SourceHwAddress:   []byte(srcHWAddr),
		SourceProtAddress: []byte(srcIP),
		DstHwAddress:      []byte(dstHWAddr),
		DstProtAddress:    []byte(dstIP),
	}

	buf := gopacket.NewSerializeBuffer()

	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}

	err := gopacket.SerializeLayers(buf, opts, &eth, &arp)
	if err != nil {
		return err
	}
	err = self.phandle.WritePacketData(buf.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func FindDevice() {
	devices, err := pcap.FindAllDevs()
	if err != nil {
		log.Fatal(err)
	}

	// Print device information
	fmt.Println("Devices found:")
	for _, device := range devices {
		fmt.Println("\nName: ", device.Name)
		fmt.Println("Description: ", device.Description)
		fmt.Println("Devices addresses: ", device.Description)
		for _, address := range device.Addresses {
			fmt.Println("- IP address: ", address.IP)
			fmt.Println("- Subnet mask: ", address.Netmask)
		}
	}
}
