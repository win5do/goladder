package main

import (
	"strings"
	"bufio"
	"fmt"
	"io"
)

func main() {
	srd := strings.NewReader("12345")
	brd := bufio.NewReader(srd)
	peekBuf, err := brd.Peek(2)
	fmt.Println(err, peekBuf)
	buf := make([]byte, 4)
	_, err = io.ReadFull(brd, buf)
	fmt.Println(err, buf)
}
