//+build ignore

package main

import (
	"fmt"
	"time"
)

func main() {
	Done := make(chan interface{})

	go func(){
		time.Sleep(1 * time.Second)
		Done <- struct{}{}
	}()

	<-Done
	fmt.Println("Done")
}