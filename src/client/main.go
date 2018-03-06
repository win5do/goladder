package main

import "goladder/src/core"

func main() {
	s := core.Socks{
		"123456",
		":8888",
		":9999",
	}
	s.ListenClient()
}