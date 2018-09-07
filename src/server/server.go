package main

import (
	"goladder/src/ss"
	"log"
	"os"
	"os/signal"
)

func main() {
	config := ss.CliFlag("./server_config.json")
	configStruct := ss.ParseConfigFile(config)
	ListenServer(configStruct)
}

func ListenServer(config ss.Config) {
	for _, i := range config.Server {
		go ss.ListenTcp(i)
		if ss.UpiFlag {
			go ss.ListenUdp(i)
		}
	}
	waitSignal()
}

func waitSignal() {
	var sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan)
	for sig := range sigChan {
		log.Printf("caught signal %v, exit", sig)
	}
}
