//+build ignore

package gabby

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
)

type HostHistry struct {
	IpAddr    string
	MacAddr   string
	TimeStamp string
}

type HostCurrentState struct {
	Ipaddr  string
	Macaddr string
	//VlanID    int
	State     int
	Timestamp time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

func (self *Database) Connect() (*gorm.DB, error) {
	DBMS := "postgres"
	USER := "user=gabby"
	//PASS := ""
	HOST := "host=127.0.0.1"
	PORT := "port=5432"
	DBNAME := "dbname=network_test"
	SSLMODE := "sslmode=disable"

	CONNECT := HOST + " " + PORT + " " + USER + " " + DBNAME + " " + SSLMODE
	db, err := gorm.Open(DBMS, CONNECT)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (self *Database) SelectAll(table string) error {
	db, err := self.Connect()
	if err != nil {
		return err
	}
	defer db.Close()

	var hostItems []HostHistry
	var hostStateItems []HostCurrentState

	if table == "Host" {
		db.Table("hosts").Find(&hostItems)
		fmt.Println(hostItems)
	} else {
		db.Table("current_map").Find(&hostStateItems)
		fmt.Println(hostStateItems)
	}

	return nil
}

func (self *Database) InsertHost(ip string, mac string) error {
	db, err := self.Connect()
	if err != nil {
		return err
	}
	defer db.Close()

	var h HostHistry
	h.IpAddr = ip
	h.MacAddr = mac
	h.TimeStamp = "current_timestamp"

	db.Table("hosts").Create(&h)

	return nil
}

func (self *Database) InsertHostState(ip string, mac string, state int) error {
	db, err := self.Connect()
	if err != nil {
		return err
	}
	defer db.Close()

	var hs HostCurrentState
	hs.Ipaddr = ip
	hs.Macaddr = mac
	hs.State = state
	//hs.Timestamp = "current_timestamp"

	db.Table("test_map").Create(&hs)

	return nil
}

//where文をそのまま書かせる
func (self *Database) UpadateHostState(ip string, mac string, state int, where string) error {
	db, err := self.Connect()
	if err != nil {
		return err
	}
	defer db.Close()

	var hsBefore HostCurrentState
	db.Table("current_map").Where(where).First(&hsBefore)

	hsAfter := hsBefore

	hsAfter.Ipaddr = ip
	hsAfter.Macaddr = mac
	hsAfter.State = state

	db.Model(&hsBefore).Update(&hsAfter)

	return nil
}
