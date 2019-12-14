package gabby

import (
	"github.com/google/gopacket/layers"
)

type Result struct {
	addr string
}

type Context struct {
	Arp            *layers.ARP
	index          int
	Engine         *Engine
	handlers       Handlers
	receiveReply   chan interface{}
	State          int
	goroutineNum   int
	ReceiveWaitNum *int
	Result         chan Result
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
