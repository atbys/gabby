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

func (self *Database) AddColumn(columnName string) error {
	var err error
	db := self.db
	db, err = sql.Open("postgres", self.OpenParameter)
	if err != nil {
		return err
	}
	defer db.Close()

	statement := fmt.Sprintf("ALTER TABLE hosts ADD %s varchar(20);", columnName)
	_, err = db.Query(statement)
	if err != nil {
		return err
	}

	self.ColumnName = append(self.ColumnName, columnName)
	return nil
}

func (self *Database) ShowDB() error {
	var err error
	db := self.db
	db, err = sql.Open("postgres", self.OpenParameter)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT * from hosts;")
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

func (self *Database) createInsertStatement(column Column) {
	var columnNames, columnValues []string

	defautStatement := "INSERT INTO hosts (%s) VALUES (%s);"
	for k, v := range column {
		columnNames = append(columnNames, k)
		columnValues = append(columnValues, v)
	}
	columnNamesStr := strings.Join(columnNames, ", ")
	columnValuesStr := strings.Join(columnValues, ", ")

	self.InsertStatement = fmt.Sprintf(defautStatement, columnNamesStr, columnValuesStr)

}

func (self *Database) Insert(column Column) {
	var err error
	db := self.db

	db, err = sql.Open("postgres", self.OpenParameter)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	self.createInsertStatement(column)
	_, err = db.Query(self.InsertStatement)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("SQL is executed.")
}
