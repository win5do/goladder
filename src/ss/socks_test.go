package ss

import (
	"net"
	"testing"
)

const (
	ADR = ":54321"
)

func makeConn(t *testing.T) (client, server sconn) {
	listen, err := net.Listen("tcp", ADR)
	defer listen.Close()
	if err != nil {
		t.Fatal(err)
	}

	conn, err := net.Dial("tcp", ADR)
	if err != nil {
		t.Fatal(err)
	}
	client = newSconn(conn, KEY)

	dst, err := listen.Accept()
	if err != nil {
		t.Fatal(err)
	}
	server = newSconn(dst, KEY)
	return
}

func TestReadAndWrite(t *testing.T) {
	client, server := makeConn(t)
	defer client.Close()
	defer server.Close()

	_, err := client.encryptWrite([]byte(SRC))
	if err != nil {
		t.Fatal(err)
	}

	buf := make([]byte, len(SRC))
	n, err := server.decryptRead(buf)
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
		_, err = server.encryptCopy(server)
		if err != nil {
			t.Fatal(err)
		}
	}()

	buf := make([]byte, 1024)
	n, err := client.Read(buf)
	if err != nil {
		t.Fatal(err)
	}

	painText, err := decrypt(buf[:n], hashKey(KEY))
	if err != nil {
		t.Fatal(err)
	}

	str := string(painText)
	t.Log(str)
	if str != SRC {
		t.Error("数据不对")
	}
}

func TestDecryptCopy(t *testing.T) {
	client, server := makeConn(t)
	defer client.Close()
	defer server.Close()

	_, err := client.encryptWrite([]byte(SRC))
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		_, err = server.decryptCopy(server)
		if err != nil {
			t.Fatal(err)
		}
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
