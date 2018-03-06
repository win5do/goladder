package main

import (
	"log"
	"time"
)

func main() {
	go func() {
		log.Println(1)
	}()

	time.Sleep(1 * time.Second)
}
