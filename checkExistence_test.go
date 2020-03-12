package gabby

import (
	"net"
	"sync"
	"testing"
	"time"

	"github.com/google/gopacket/layers"
)

func TestCheckExistence(t *testing.T) {
	e, err := New()
	if err != nil {
		t.Fatal(err)
	}
	srcHWAddr, err := net.ParseMAC("00-00-00-00-00-00")
	srcIP := net.ParseIP("192.168.13.1")
	dstHWAddr, err := net.ParseMAC("ff-ff-ff-ff-ff-ff")
	dstIP := net.ParseIP("192.169.13.89")
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
	c := &Context{
		Arp:    &arp,
		State:  CHECKING,
		Engine: e,
	}

	c.handlers = append(c.handlers, CheckExistenceMiddleware)
	c.handlers = append(c.handlers, func(c *Context) {})

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		c.Start()
		wg.Done()
	}()
	wg.Wait()
	if c.State != UNUSED {
		t.Fatalf("checkin failed")
	}

	dstIP = net.ParseIP("192.168.13.1")
	arp.DstProtAddress = []byte(dstIP)
	c.index = 0
	c.recvReply = make(chan interface{})

	wg.Add(1)
	go func() {
		c.Start()
		wg.Done()
	}()

	time.Sleep(1 * time.Second)
	c.recvReply <- struct{}{}

	wg.Wait()
	if c.State != USED {
		t.Fatalf("missed existence")
	}
}
