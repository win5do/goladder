package main

import "goladder/src/core"

func main() {
	core.ListenServer(core.Config{
		nil,
		[]core.ServerConfig{
			{":9999", "123456", nil},
			{":9998", "123456", nil},
		},
	})
}