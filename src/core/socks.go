package core

import "net"

type Socks struct {
	conn *net.Conn
}

func NewSocks(conn *net.Conn) Socks {
	return Socks{
		conn,
	}
}

func (*Socks) Encode(b []byte) (n int, err error) {
	return
}

func (*Socks) Decode(b []byte) (n int, err error) {
	return
}
