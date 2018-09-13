package main

import (
	"goladder/src/ss"
)

func main() {
	config := ss.CliFlag("./server_config.json")
	ss.ParseConfigFile(config)
	ListenServer()
}

func ListenServer() {
	for _, i := range ss.Conf.Server {
		go ss.ListenTcp(i)
		if ss.Conf.Udp {
			go ss.ListenUdp(i)
		}
	}
	ss.WaitSignal()
}
