package main

import "goladder/src/core"

func main() {
	s := core.Socks{
		"123456",
		":8888",
		"45.77.145.34:9999",
	}
	s.ListenClient()
}