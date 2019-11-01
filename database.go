package gabby

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/lib/pq"
)

type HOST struct {
	ipaddr    string
	macaddr   string
	timestamp string
}

type Database struct {
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
		return err
	}
	defer db.Close()

	statement := fmt.Sprintf("ALTER TABLE network_host ADD %s varchar(20);", columnName)
	_, err = db.Query(statement)
	if err != nil {
		return err
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

func (engine *Engine) CreateInsertStatement(column Column) {
	var columnNames, columnValues []string

	defautStatement := "INSERT INTO network_host (%s) VALUES (%s);"
	for k, v := range column {
		columnNames = append(columnNames, k)
		columnValues = append(columnValues, v)
	}
	columnNamesStr := strings.Join(columnNames, ", ")
	columnValuesStr := strings.Join(columnValues, ", ")

	engine.dbinfo.InsertStatement = fmt.Sprintf(defautStatement, columnNamesStr, columnValuesStr)

}

func (engine *Engine) Insert(column Column) {
	var err error
	db := engine.dbinfo.db

	db, err = sql.Open("postgres", engine.dbinfo.OpenParameter)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	engine.CreateInsertStatement(column)
	_, err = db.Query(engine.dbinfo.InsertStatement)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("SQL is executed.")
}
