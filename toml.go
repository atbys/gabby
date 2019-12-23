package gabby

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Device   DeviceConfig
	Database Database
	Routers  []RouterConfig `toml:"routers"`
}

type Database struct {
	Host   string `toml:"host""`
	Port   int    `toml:"port"`
	User   string `toml:"user"`
	Pass   string `toml:"pass"`
	Dbname string `toml:"dbname"`
}

type DeviceConfig struct {
	Name   string `toml:"name"`
	Hwaddr string `toml:"hwaddr"`
}

type RouterConfig struct {
	RouterIP  string `toml:"router_ip"`
	RouterMAC string `toml:"router_mac"`
}

func (self *Engine) readConfig() error {
	_, err := toml.DecodeFile("config.toml", &self.Config)
	if err != nil {
		return err
	}

	fmt.Println("[DEBUG] device name is ", self.Config.Device.Name)
	self.DB = self.Config.Database

	return nil
}
