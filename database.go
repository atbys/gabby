package gomamiso

import (
	"fmt"

	_ "github.com/lib/pq"
)

func (engine *Engine) ShowDB() error {
	rows, err := engine.db.Query("SELECT * from network_host")
	if err != nil {
		return err
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

	return nil
}
