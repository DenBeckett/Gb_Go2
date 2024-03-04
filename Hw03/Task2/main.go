package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM)
	i := 1
	for {

		select {
		case <-exit:
			fmt.Println("Graceful shutdown!")
			return
		default:
			fmt.Println("i^2 =", i*i)
			time.Sleep(250 * time.Millisecond)
			i++
		}

	}

}
