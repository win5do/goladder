package core

import (
	"net"
	"testing"
	"time"
)

const ADR = ":54321"

func makeConn(t *testing.T) (local, remote sconn) {
	listen, err := net.Listen("tcp", ADR)
	defer listen.Close()
	if err != nil {
		t.Fatal(err)
	}

	conn, err := net.Dial("tcp", ADR)
	if err != nil {
		t.Fatal(err)
	}
	local = newSconn(KEY, conn)

	dst, err := listen.Accept()
	if err != nil {
		t.Fatal(err)
	}
	remote = newSconn(KEY, dst)
	return
}

func TestReadAndWrite(t *testing.T) {
	local, remote := makeConn(t)
	defer local.Close()
	defer remote.Close()

	_, err := local.encryptWrite([]byte(SRC))
	if err != nil {
		t.Fatal(err)
	}

	buf := make([]byte, len(SRC))
	n, err := remote.decryptRead(buf)
	if err != nil {
		t.Fatal(err)
	}
	if n != len(SRC) || string(buf) != SRC {
		t.Error("数据不对")
	}
}

func TestEncryptCopy(t *testing.T) {
	local, remote := makeConn(t)
	go func() {
		time.Sleep(time.Second * 1)
		local.Close()
		remote.Close()
	}()

	_, err := local.Write([]byte(SRC))
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		_, err = remote.encryptCopy(local)
		if err != nil {
			t.Fatal(err)
		}
	}()

	encrypted, err := encrypt([]byte(SRC), hashKey(KEY))
	if err != nil {
		t.Fatal(err)
	}

	for {
		buf := make([]byte, len(encrypted))
		n, err := local.Read(buf)
		if err != nil {
			break
		}

		if n == 0 {
			continue
		}

		str := string(buf)
		t.Log(str)
		if n != len(encrypted) || str != string(encrypted) {
			t.Error("数据不对")
		}
	}
}

func TestDecryptCopy(t *testing.T) {
	local, remote := makeConn(t)
	go func() {
		time.Sleep(time.Second * 1)
		local.Close()
		remote.Close()
	}()

	_, err := local.encryptWrite([]byte(SRC))
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		_, err = remote.decryptCopy(local)
		if err != nil {
			t.Fatal(err)
		}
	}()

	for {
		buf := make([]byte, len(SRC))
		n, err := local.Read(buf)
		if err != nil {
			break
		}

		if n == 0 {
			continue
		}

		str := string(buf)
		t.Log(str)
		if n != len(SRC) || str != SRC {
			t.Error("数据不对")
		}
	}
}
