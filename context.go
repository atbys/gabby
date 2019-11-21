package gabby

import "github.com/google/gopacket/layers"

type Context struct {
	Arp      *layers.ARP
	index    int
	engine   *Engine
	handlers Handlers
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
