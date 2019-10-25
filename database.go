package gomamiso

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

type HOST struct {
	ipaddr    string
	macaddr   string
	timestamp string
}

type DatabaseInfo struct {
	db              *sql.DB
	OpenParameter   string
	ColumnName      []string
	InsertStatement string
}

type Column map[string]string

func (engine *Engine) AddColumn(columnName string) error {
	var err error
	db := engine.dbinfo.db
	db, err = sql.Open("postgres", engine.dbinfo.OpenParameter)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	statement := fmt.Sprintf("ALTER TABLE network_host ADD COLUMN %s;", columnName)
	_, err = db.Query(statement)
	if err != nil {
		log.Fatal(err)
	}

	engine.dbinfo.ColumnName = append(engine.dbinfo.ColumnName, columnName)
	return nil
}

func (engine *Engine) ShowDB() error {
	var err error
	db := engine.dbinfo.db
	db, err = sql.Open("postgres", engine.dbinfo.OpenParameter)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT * from network_host;")
	if err != nil {
		return err
	}

	var hs []HOST
	for rows.Next() {
		var h HOST
		rows.Scan(&h.ipaddr, &h.macaddr, &h.timestamp)
		hs = append(hs, h)
	}

	for _, e := range hs {
		fmt.Printf("%+v\n", e)
	}

	return nil
}

func (engine *Engine) CreateInsertStatement() {
	defautStatement := "INSERT INTO network_host TABLE (%s) VALUES (%s);"
	columnNames := ""
	columnValues := ""
	for i, name := range engine.dbinfo.ColumnName {
		columnNames += name + ", "
		columnValues += "$" + string(i+1) + ", "
	}

	engine.dbinfo.InsertStatement = fmt.Sprintf(defautStatement, columnNames, columnValues)

}

//TODO
func (engine *Engine) Insert() {
	var err error
	db := engine.dbinfo.db

	db, err = sql.Open("postgres", engine.dbinfo.OpenParameter)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ipaddr := "192.168.2.1"
	macaddr := "ff-ff-ff-ff-ff-ff"

	statement := "INSERT INTO network_host VALUES ($1, $2, current_timestamp);"
	stmt, err := db.Prepare(statement)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(ipaddr, macaddr)
	for rows.Next() {
		h := HOST{}
		rows.Scan(&h.ipaddr, &h.macaddr, &h.timestamp)
		fmt.Println(h)
	}
}
