package main

import (
	"fmt"
)

func main() {
	b := make([]byte, 5)
	b1 := b[:3]
	b[4] = 5
	b[2] = 3
	fmt.Println(b1, len(b1), cap(b1))
}
