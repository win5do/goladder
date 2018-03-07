package main

import "goladder/src/core"

func main() {
	core.ListenClient(core.Config{
		":8888",
		[]core.ServerConfig{
			{":9999", "123456", 60},
			{":9998", "123456", 40},
		},
	})
}
