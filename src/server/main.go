package main

import "goladder/src/ss"

func main() {
	config := ss.CliFlag("./server_config.json")
	configStruct := ss.ParseConfigFile(config)
	ss.ListenServer(configStruct)
}
