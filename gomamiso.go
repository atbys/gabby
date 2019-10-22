package gomamiso

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"

	_ "github.com/lib/pq"
)

const (
	REQUEST = iota
	REPLY
	INIT
	HOOK_POINT_NUM
)

type Action struct {
	hooked bool
	exec   func()
}

type Engine struct {
	Device       string
	snapshot_len int32
	promiscuous  bool
	timeout      time.Duration
	handle       *pcap.Handle
	action       []*Action
	db           *sql.DB
}

type HOST struct {
	IP   string
	MAC  string
	TIME string
}

func New() *Engine {
	return &Engine{}
}

func ClearAction() []*Action {
	var action []*Action
	for i := 0; i < HOOK_POINT_NUM; i++ {
		action = append(action, &Action{
			hooked: false,
			exec:   nil,
		})
	}
	return action
}

func Default() *Engine {
	engine := &Engine{
		Device:       "enp0",
		snapshot_len: 1024,
		promiscuous:  true,
		timeout:      30 * time.Second,
		action:       ClearAction(),
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

func (engine *Engine) Init() error {
	var err error
	engine.db, err = sql.Open("postgres", "host=127.0.0.1 port=5432 dbname=hellosql sslmode=disable")
	if err != nil {
		return err
	}

	//Open device
	if engine.Device == "" {
		log.Println("Please Set Device in below")
		FindDevice()
		engine.db.Close()
		return errors.New("cannot open device")
	}

	engine.handle, err = pcap.OpenLive(engine.Device, engine.snapshot_len, engine.promiscuous, engine.timeout)
	if err != nil {
		engine.db.Close()
		return err
	}

	return nil
}

func (engine *Engine) Deinit() {
	engine.db.Close()
}

func (engine *Engine) Run() error {
	err := engine.Init()
	if err != nil {
		return err
	}
	defer engine.Deinit()

	if false {
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
	
	return nil
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
