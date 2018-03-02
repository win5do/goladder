package main

import "goladder/src/core"

func main() {
	core.Listen("9999", core.HandleServer)
}
