//+build ignore

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGINT)

	fmt.Printf("process id: %d\n", os.Getpid)

	for {
		select {
		case <-sigCh:
			fmt.Println("Don't Stop")
		}
	}
}
