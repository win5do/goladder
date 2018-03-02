package main

import "fmt"

func main() {
	b1 := []byte{5}
	b2 := []byte{0x05}
	fmt.Println(b1[0] == 0x05)
	fmt.Println(b2[0] == 5)
}
