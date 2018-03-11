package ss

import (
	"testing"
	"fmt"
)

const (
	KEY = "123456"
	SRC = "321"
)

func TestScipher(t *testing.T) {
	iv := randIv()
	k := hashKey(KEY)
	scipher, err := NewScipher(k, iv)
	if err != nil {
		t.Fatal(err)
	}

	src := []byte(SRC)
	buf := make([]byte, len(src))
	scipher.encrypt(buf, src)
	fmt.Println("byte:", buf)
	fmt.Printf("string:%s\n", buf)

	scipher.decrypt(buf, buf)
	fmt.Println("byte:", buf)
	fmt.Printf("string:%s\n", buf)

	if string(buf) != SRC {
		t.Error("加解密错误")
	}
}
