package gabby

import (
	"net"
	"fmt"
)

func Default() (*Engine, error){
	e, err := New()
	if err != nil {
		return nil, err
	}
	e.Use(defaultLogMiddleware)
	e.Use(defaultDBMiddleware)
	return e, nil
}

func defaultDBMiddleware(c *Context) {
	//preprocess
	column := Column{
		"ipaddr": "'" + net.IP(c.Arp.SourceProtAddress).String() + "'",
		"macaddr": "'" + net.HardwareAddr(c.Arp.SourceHwAddress).String() + "'",
		"timestamp": "current_timestamp",
	}
	c.engine.DB.Insert(column)

	c.Next()

	//postprocess
	c.engine.DB.ShowDB()
}

func defaultLogMiddleware(c *Context) {
	//preprocess
	fmt.Printf("\x1b[31m%s\x1b[0mARP packet get!!\n", "[+]")
	c.Next()
	fmt.Printf("\x1b[31m%s\x1b[0mNext packet waiting...\n", "[+]")
	//postprocess
}

