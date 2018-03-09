package main

import "goladder/src/ss"

func main() {
	config := ss.CliFlag("./client_config.json")
	configStruct := ss.ParseConfigFile(config)
	ss.ListenClient(configStruct)
}