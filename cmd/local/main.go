package main

import (
	"goladder/src/local"
	"goladder/src/ss"
)

func main() {
	config := ss.CliFlag("./local_config.json")
	ss.ParseConfigFile(config)
	local.ListenLocal()
}
