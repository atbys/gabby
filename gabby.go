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

const (
	REQUEST_FROM_ROUTER = iota
	REQUEST_FROM_HOST
	USED_PACKET
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
	// RequestFromRouterHandlers Handlers
	// RequestFromHostHandlers Handlers
	// UsedHandles		Handlers
	registedHandlers map[int]Handlers
	DB               Database
	Config           Config
	IsVLAN           bool
	logger           *log.Logger
}

func New() (*Engine, error) {
	e := &Engine{
		snapshotLen:      1024,
		promiscuous:      true,
		timeout:          30 * time.Second,
		RequestHandlers:  make(map[string]Handlers),
		ReplyHandlers:    make(map[string]Handlers),
		registedHandlers: make(map[int]Handlers),
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

func (self *Engine) analizePacket(packet gopacket.Packet) (*Context, int) {
	myHWaddr, _ := net.ParseMAC(self.Config.Device.Hwaddr)

	ethLayer := packet.Layer(layers.LayerTypeEthernet)
	if ethLayer == nil {
		return nil, 0
	}
	eth := ethLayer.(*layers.Ethernet)

	arpLayer := packet.Layer(layers.LayerTypeARP)
	if arpLayer == nil {
		return nil, 0
	}
	arp := arpLayer.(*layers.ARP)
	if bytes.Equal([]byte(myHWaddr), []byte(eth.SrcMAC)) {
		// This is a packet I sent.
		return nil, 0
	}

	srcIP := net.IP(arp.SourceProtAddress).String()
	//dstIP := net.IP(arp.DstProtAddress).String()

	c := &Context{
		Arp:        arp,
		index:      0,
		Engine:     self,
		DstIPaddr:  net.IP(arp.DstProtAddress),
		SrcIPaddr:  net.IP(arp.SourceProtAddress),
		DstMACaddr: eth.DstMAC,
		SrcMACaddr: eth.SrcMAC,
		pid:        net.IP(arp.SourceProtAddress).String() + net.IP(arp.DstProtAddress).String(),
		recvReply:  make(chan interface{}),
	}

	if self.IsVLAN {
		dot1qLayer := packet.Layer(layers.LayerTypeDot1Q)
		if dot1qLayer == nil {
			return nil, 0
		}
		dot1q := dot1qLayer.(*layers.Dot1Q)
		c.VlanID = dot1q.VLANIdentifier
	}

	//var isRequest = arp.Operation == layers.ARPRequest
	var isReply = arp.Operation == layers.ARPReply

	isFromRouter := isFromRouter(self.Config.Routers, srcIP)

	if isReply {
		return c, USED_PACKET
	} else {
		if isFromRouter {
			return c, REQUEST_FROM_ROUTER
		} else if bytes.Equal(arp.SourceProtAddress, arp.DstProtAddress) {
			return c, USED_PACKET
		} else {
			return c, REQUEST_FROM_HOST
		}
	}
}

func isFromRouter(routers []RouterConfig, srcIP string) bool {
	for _, router := range routers {
		if router.RouterIP == srcIP {
			return true
		}
	}

	return false
}

func (self *Engine) HandleManager(packets chan gopacket.Packet) {
	blockDupList := make(map[string]*Context)
	waitRecvList := make(map[string]*Context)
	result := make(chan Result, 1024)

	for {
		select {
		case packet := <-packets:
			// Process packet here
			// log.Printf("IP %v is at %v", net.IP(arp.SourceProtAddress), net.HardwareAddr(arp.SourceHwAddress))
			c, packetClass := self.analizePacket(packet)
			if c == nil {
				break
			}

			if packetClass == USED_PACKET {
				waitC, ok := waitRecvList[c.SrcIPaddr.String()]
				fmt.Println("hey")
				fmt.Println(waitRecvList)
				if ok {
					waitC.recvReply <- struct{}{}
				}
				fmt.Println("yo")
				delete(waitRecvList, c.SrcIPaddr.String())
			} else {
				if _, ok := blockDupList[c.pid]; ok {
					break
				}
				log.Println("add")
				c.Result = result
				waitRecvList[c.DstIPaddr.String()] = c
				blockDupList[c.pid] = c
			}

			h, ok := self.registedHandlers[packetClass]
			if ok {
				c.handlers = h
			} else {
				break
			}

			go c.Start()

		case r := <-result:
			log.Println("delete")
			delete(blockDupList, r.pid)
			delete(waitRecvList, r.dstIP)
			//fmt.Println(processingHandle)
		}
	}
}

func (self *Engine) RegistHandle(packetClass int, handle Handle) {

	self.registedHandlers[packetClass] = append(self.registedHandlers[packetClass], handle)
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

func (self *Engine) SendRequestARPPacketWithVLAN(dstHWAddr net.HardwareAddr, dstIP net.IP, srcHWAddr net.HardwareAddr, srcIP net.IP, vlanid uint16) error {
	eth := layers.Ethernet{
		SrcMAC:       srcHWAddr,
		DstMAC:       dstHWAddr,
		EthernetType: layers.EthernetTypeDot1Q,
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

	dot1q := layers.Dot1Q{
		VLANIdentifier: vlanid,
		Type:           layers.EthernetTypeARP,
	}

	buf := gopacket.NewSerializeBuffer()

	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}

	err := gopacket.SerializeLayers(buf, opts, &eth, &dot1q, &arp)
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

func (self *Engine) SendReplyARPPacketWithVLAN(dstHWAddr net.HardwareAddr, dstIP net.IP, srcHWAddr net.HardwareAddr, srcIP net.IP, vlanid uint16) error {
	eth := layers.Ethernet{
		SrcMAC:       srcHWAddr,
		DstMAC:       dstHWAddr,
		EthernetType: layers.EthernetTypeARP,
	}

	dot1q := layers.Dot1Q{
		Priority:       0,
		DropEligible:   false,
		VLANIdentifier: vlanid,
		Type:           layers.EthernetTypeARP,
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

	err := gopacket.SerializeLayers(buf, opts, &eth, &dot1q, &arp)
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
