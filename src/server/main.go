package main

import "goladder/src/core"

func main() {
	s := core.Socks{
		"123456",
		":9999",
		"",
	}
	s.ListenServer()
}
