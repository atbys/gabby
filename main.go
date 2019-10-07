package gomamiso

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

const (
	REQUEST = iota
	REPLY
	INIT
)

type Action struct {
	hooked bool
	exec   func()
}

type Engine struct {
	Device       string
	snapshot_len int32
	promiscuous  bool
	err          error
	timeout      time.Duration
	handle       *pcap.Handle
	action       []*Action
}

func DefaultAction() []*Action {
	var action []*Action
	for i := 0; i < 3; i++ {
		action = append(action, &Action{
			hooked: false,
			exec:   nil,
		})
	}
	return action
}

func Default() *Engine {
	engine := &Engine{
		Device:       "",
		snapshot_len: 1024,
		promiscuous:  false,
		timeout:      30 * time.Second,
		action:       DefaultAction(),
	}

	return engine
}

func (engine *Engine) SetDevice(name string) {
	engine.Device = name
}

func (engine *Engine) SetHook(point int, fn func()) {
	engine.action[point].exec = fn
	engine.action[point].hooked = true
}

func (engine *Engine) Run() {
	//Open device
	if engine.Device == "" {
		log.Println("Please Set Device in below")
		FindDevice()
		return
	}

	engine.handle, engine.err = pcap.OpenLive(engine.Device, engine.snapshot_len, engine.promiscuous, engine.timeout)
	if engine.err != nil {
		log.Fatal(engine.err)
	}
	pakcetSource := gopacket.NewPacketSource(engine.handle, engine.handle.LinkType())
	packets := pakcetSource.Packets()
	defer close(packets)

	if engine.action[INIT].hooked {
		engine.action[INIT].exec()
	}

	go engine.PacketCapture(engine.handle, packets)

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()

}

func (engine *Engine) PacketCapture(handle *pcap.Handle, packets chan gopacket.Packet) {
	for packet := range packets {
		arpLayer := packet.Layer(layers.LayerTypeARP)
		if arpLayer == nil {
			continue
		}
		arp := arpLayer.(*layers.ARP)

		var isRequest = arp.Operation == layers.ARPRequest
		var isReply = arp.Operation == layers.ARPReply

		if isRequest && engine.action[REQUEST].hooked {
			engine.action[REQUEST].exec()
		} else if isReply && engine.action[REPLY].hooked {
			engine.action[REPLY].exec()
		} else {
			fmt.Println("unknown type")
		}
	}
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
