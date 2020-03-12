//+build ignore

package main

import (
	"log"
	"os"
)

func main() {
	f, _ := os.Create("Hello.txt")
	logger := log.New(f, "Logger: ", log.Lshortfile)

	logger.Println("Hello")
}
