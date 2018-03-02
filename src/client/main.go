package main

import "goladder/src/core"

func main() {
	core.Listen("8888", core.HandleClient)
}
