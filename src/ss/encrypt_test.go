package ss

import (
	"testing"
	"fmt"
)

const (
	KEY = "123456"
	SRC = "321"
)

func TestEncrypt(t *testing.T) {
	enc, err := encrypt([]byte(SRC), hashKey(KEY))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("byte:", enc)
	fmt.Printf("string:%s\n", enc)

	painText, err := decrypt(enc, hashKey(KEY))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(painText))
	if string(painText) != SRC {
		t.Error("加解密错误")
	}
}
