package main

import (
	"goladder/src/server"
	"goladder/src/ss"
)

func main() {
	config := ss.CliFlag("./server_config.json")
	ss.ParseConfigFile(config)
	server.ListenServer()
}
