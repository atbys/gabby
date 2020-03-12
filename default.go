package gabby

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"time"
)

var (
	BroadcastMAC = net.HardwareAddr{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
	ZeroIP       = net.IP{0, 0, 0, 0}
)

const (
	PROVE_MIN     float32 = 1
	PROVE_MAX     float32 = 2
	PROVE_NUM     int     = 3
	ANNOUNCE_WAIT         = 2.0
)

const (
	UNUSED = iota
	USED
	CHECKING
)

func Default() (*Engine, error) {
	e, err := New()
	if err != nil {
		return nil, err
	}

	f, _ := os.Create("log")
	e.logger = log.New(f, "log: ", log.Lshortfile)
	e.RegistHandle(REQUEST_FROM_ROUTER, defaultInitMiddleware)
	e.RegistHandle(REQUEST_FROM_HOST, defaultInitMiddleware)

	e.RegistHandle(REQUEST_FROM_HOST, CheckExistenceMiddleware)
	e.RegistHandle(REQUEST_FROM_ROUTER, CheckExistenceMiddleware)
	return e, nil
}

func defaultInitMiddleware(c *Context) {
	c.goroutineNum += 1
	c.Next()
	c.goroutineNum -= 1
	c.Result <- Result{
		pid:   c.pid,
		dstIP: c.DstIPaddr.String(),
	}
}

func defaultLogMiddleware(c *Context) {
	//preprocess
	//fmt.Printf("\x1b[31m%s\x1b[0mARP packet get!!\n", "[+]")

	c.Engine.logger.Printf("[%v] %v(%v) is USED\n", time.Now(), c.SrcIPaddr, c.SrcMACaddr)
	c.Next()
	if c.State == UNUSED {
		c.Engine.logger.Printf("[%v] %v(%v) is UNUSED\n", time.Now(), c.DstIPaddr, c.DstMACaddr)
	} else {
		c.Engine.logger.Printf("[%v] %v(%v) is USED\n", time.Now(), c.DstIPaddr, c.DstMACaddr)
	}
	//fmt.Printf("\x1b[31m%s\x1b[0mNext packet waiting...\n", "[+]")
	//postprocess
}

func CheckExistenceMiddleware(c *Context) {
	fmt.Println("[DEBUG] check existence start")
	rand.Seed(time.Now().UnixNano())

	c.State = CHECKING

	//MyMACAddr, err := net.ParseMAC("f4:5c:89:bf:e1:09")
	MyMACAddr, err := net.ParseMAC(c.Engine.Config.Device.Hwaddr)
	if err != nil {
		log.Fatalf("parse mac: %v", err)
	}
	var randFloat float32
SENDLOOP:
	for i := 0; i < PROVE_NUM; i++ {
		if i == PROVE_NUM-1 {
			randFloat = ANNOUNCE_WAIT
		} else {
			randFloat = PROVE_MIN + rand.Float32()*(PROVE_MAX-PROVE_MIN)
		}
		var err error
		if c.Engine.IsVLAN {
			err = c.Engine.SendRequestARPPacketWithVLAN(BroadcastMAC, net.IP(c.Arp.DstProtAddress).To4(), MyMACAddr, ZeroIP, c.VlanID)
		} else {
			err = c.Engine.SendRequestARPPacket(BroadcastMAC, net.IP(c.Arp.DstProtAddress).To4(), MyMACAddr, ZeroIP)
		}

		if err != nil {
			log.Fatalf("send arp: %v", err)
		}
		t := time.NewTimer(time.Duration(randFloat) * time.Second)

		select {
		case <-t.C:
			continue
		case <-c.recvReply:
			c.State = USED
			break SENDLOOP
		}
	}

	if c.State != USED {
		c.State = UNUSED
	}

	c.Next()

}
