//+build ignore

package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/lib/pq"
)

type Column map[string]string

type HOST struct {
	ipaddr    string
	macaddr   string
	timestamp string
}

func CreateInsertStatement(column Column) string {
	var columnNames, columnValues []string

	defautStatement := "INSERT INTO network_host (%s) VALUES (%s);"
	for k, v := range column {
		columnNames = append(columnNames, k)
		columnValues = append(columnValues, v)
	}
	columnNamesStr := strings.Join(columnNames, ", ")
	columnValuesStr := strings.Join(columnValues, ", ")

	InsertStatement := fmt.Sprintf(defautStatement, columnNamesStr, columnValuesStr)

	return InsertStatement
}

func main() {
	db, err := sql.Open("postgres", "host=127.0.0.1 port=5432 dbname=exampledb user=miso password=hello sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	column := Column{
		"ipaddr":    "'192.168.3.1'",
		"macaddr":   "'11-11-11-11-11-11'",
		"timestamp": "current_timestamp",
	}
	InsertStatement := CreateInsertStatement(column)
	fmt.Println("This SQL is executed:", InsertStatement)

	_, err = db.Query(InsertStatement)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("SQL is executed.")
}
