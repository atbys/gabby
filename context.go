package gabby

import (
	"github.com/google/gopacket/layers"
	"net"
)

type Result struct {
	addr string
}

type Context struct {
	Arp            *layers.ARP
	FromRouter     bool
	index          int
	VlanID         uint16
	Engine         *Engine
	handlers       Handlers
	receiveReply   chan interface{}
	State          int
	goroutineNum   int
	ReceiveWaitNum *int
	Result         chan Result
	SetAddr        string
	DstIPaddr      net.IP
	SrcIPaddr      net.IP
	DstMACaddr     net.HardwareAddr
	SrcMACaddr     net.HardwareAddr
}

func (self *Context) Start() {
	for self.index < len(self.handlers) {
		self.handlers[self.index](self)
		self.index++
	}
}

func (self *Context) Next() {
	self.index++
	self.handlers[self.index](self)
}
