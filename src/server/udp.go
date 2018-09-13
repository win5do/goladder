package server

import (
	"log"
	"net"

	"goladder/src/ss"
)

func listenUdp(oneServer ss.ServerConfig) {
	pkConn, err := net.ListenPacket("udp", oneServer.Addr)
	if err != nil {
		log.Println(err)
		return
	}
	defer pkConn.Close()

	for {
		buf := make([]byte, 4096)
		_, addr, err := pkConn.ReadFrom(buf)
		if err != nil {
			return
		} else {
			go handleUdpPacket(pkConn, addr, buf)
		}
	}
}

// udp转发
func handleUdpPacket(pkConn net.PacketConn, addr net.Addr, buf []byte) {
	// 把头解出来 拿到报文和地址
	net.Dial("udp", "")

	pkConn.WriteTo(buf, addr)
}
