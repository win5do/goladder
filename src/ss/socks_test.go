package ss

import (
	"net"
	"testing"
)

const (
	ADDR = ":54321"
)

func makeConn(t *testing.T) (sclient, sserver *Sconn) {
	listen, err := net.Listen("tcp", ADDR)
	if err != nil {
		t.Fatal(err)
	}
	defer listen.Close()

	client, err := net.Dial("tcp", ADDR)
	if err != nil {
		t.Fatal(err)
	}

	iv := randIv()
	sclient, err = newSconn(client, KEY, iv)
	if err != nil {
		t.Fatal(err)
	}

	server, err := listen.Accept()
	if err != nil {
		t.Fatal(err)
	}
	sserver, err = newSconn(server, KEY, iv)
	if err != nil {
		t.Fatal(err)
	}
	return
}

func TestReadAndWrite(t *testing.T) {
	client, server := makeConn(t)
	defer client.Close()
	defer server.Close()

	_, err := client.EncryptWrite([]byte(SRC))
	if err != nil {
		t.Fatal(err)
	}

	buf := make([]byte, len(SRC))
	n, err := server.DecryptRead(buf)
	if err != nil {
		t.Fatal(err)
	}
	if n != len(SRC) || string(buf) != SRC {
		t.Error("数据不对")
	}
}

func TestEncryptCopy(t *testing.T) {
	client, server := makeConn(t)
	defer client.Close()
	defer server.Close()

	_, err := client.Write([]byte(SRC))
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		EncryptCopy(server, server)
	}()

	buf := make([]byte, 1024)
	n, err := client.DecryptRead(buf)
	if err != nil {
		t.Fatal(err)
	}

	str := string(buf[:n])
	t.Log(str)
	if str != SRC {
		t.Error("数据不对")
	}
}

func TestDecryptCopy(t *testing.T) {
	client, server := makeConn(t)
	defer client.Close()
	defer server.Close()

	_, err := client.EncryptWrite([]byte(SRC))
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		DecryptCopy(server, server)
	}()

	buf := make([]byte, 1024)
	n, err := client.Read(buf)
	if err != nil {
		t.Fatal(err)
	}

	str := string(buf[:n])
	t.Log(str)
	if n != len(SRC) || str != SRC {
		t.Error("数据不对")
	}
}
