package main

import "goladder/src/ss"

func main() {
	config := ss.CliFlag("./local_config.json")
	configStruct := ss.ParseConfigFile(config)
	ss.ListenLocal(configStruct)
}