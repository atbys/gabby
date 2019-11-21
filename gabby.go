package gabby

import (
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
	snapshotLen     int32
	promiscuous     bool
	timeout         time.Duration
	phandle         *pcap.Handle
	RequestHandlers map[string]Handlers
	ReplyHandlers   map[string]Handlers
	dbinfo          Database
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
	info, err := ReadAndSetInfo()
	if err != nil {
		return err
	}
	self.dbinfo.OpenParameter = fmt.Sprintf(
		"host=127.0.0.1 port=5432 user=%s password=%s dbname=%s sslmode=disable",
		info.Database.User,
		info.Database.Password,
		"exampledb",
	)

	self.dbinfo.ColumnName = append(self.dbinfo.ColumnName, "ipaddr", "macaddr", "timestamp")
	//Open device
	if info.Device.Name == "" {
		log.Println("Please Set Device in below")
		FindDevice()
		return errors.New("cannot open device")
	}

	self.DeviceName = info.Device.Name

	self.phandle, err = pcap.OpenLive(self.DeviceName, self.snapshotLen, self.promiscuous, self.timeout)
	if err != nil {
		return err
	}

	return nil
}

func (self *Engine) Run() {
	pakcetSource := gopacket.NewPacketSource(self.phandle, self.phandle.LinkType())
	packets := pakcetSource.Packets()
	defer close(packets)

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)

	fmt.Println("START")
	go self.PacketAnalyze(packets)

	<-quit

	fmt.Println("[Interrupt] Exit")
}

func (self *Engine) PacketAnalyze(packets chan gopacket.Packet) {
	for packet := range packets {
		arpLayer := packet.Layer(layers.LayerTypeARP)
		if arpLayer == nil {
			continue
		}
		arp := arpLayer.(*layers.ARP)

		var isRequest = arp.Operation == layers.ARPRequest
		var isReply = arp.Operation == layers.ARPReply

		srcIP := net.IP(arp.SourceProtAddress).String()

		c := &Context{
			Arp:    arp,
			index:  0,
			engine: self,
		}

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
		} else {
			log.Fatalf("Unknown type ARP")
		}

		c.Start()
	}
}

func (self *Engine) Request(addr string, handle Handle) {
	self.RequestHandlers[addr] = append(self.RequestHandlers[addr], handle)
	fmt.Println("Request handlers num =", len(self.RequestHandlers[addr]))
}

func (self *Engine) Reply(addr string, handle Handle) {
	self.ReplyHandlers[addr] = append(self.ReplyHandlers[addr], handle)
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
