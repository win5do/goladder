package ss

import (
	"log"
	"net"
)

type udpNatMap map[string]net.Conn

var UdpNat udpNatMap

func init() {
	UdpNat = make(udpNatMap)
}

func genKey(client, server string) string {
	return client + "->" + server
}

func (nat udpNatMap) Put(client, server string, conn net.Conn) {
	key := genKey(client, server)
	nat[key] = conn
}

func (nat udpNatMap) Get(client, server string) net.Conn {
	key := genKey(client, server)
	if conn, ok := nat[key]; ok {
		return conn
	}
	return nil
}

func (nat udpNatMap) GenConn(client, server string) net.Conn {
	conn := nat.Get(client, server)

	if conn == nil {
		conn, err := net.DialTimeout("udp", server, TIMEOUT)
		if err != nil {
			log.Println(err)
			return nil
		}
		nat.Put(client, server, conn)
	}

	return conn
}
