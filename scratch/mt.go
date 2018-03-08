package main

import (
	"fmt"
)

type st struct {
	s string
}

func main() {
	st := st{}
	fmt.Println(st.s == "")
}
