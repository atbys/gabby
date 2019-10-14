package gomamiso

import (
	"bufio"
	"database/sql"
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
		promiscuous:  true,
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

func (engine *Engine) ShowDB() {
	rows, err := engine.db.Query("SELECT * from network_host")
	if err != nil {
		fmt.Println(err)
	}

	var hs []HOST
	for rows.Next() {
		var h HOST
		rows.Scan(&h.IP, &h.MAC, &h.TIME)
		hs = append(hs, h)
	}

	for _, e := range hs {
		fmt.Printf("%+v\n", e)
	}
}
func (engine *Engine) Init() {
	var err error
	engine.db, err = sql.Open("postgres", "host=127.0.0.1 port=5432 dbname=exampledb sslmode=disable")
	if err != nil {
		log.Fatal(err)
		engine.db.Close()
	}

	//Open device
	if engine.Device == "" {
		log.Println("Please Set Device in below")
		FindDevice()
		os.Exit(1)
	}

	engine.handle, err = pcap.OpenLive(engine.Device, engine.snapshot_len, engine.promiscuous, engine.timeout)
	if err != nil {
		log.Fatal(err)
	}
}

func (engine *Engine) Deinit() {
	engine.db.Close()
}

func (engine *Engine) Run() {
	engine.Init()
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
	engine.ShowDB()
	engine.Deinit()
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
