package local

import (
	"log"
	"net"
	"os"

	"goladder/src/ss"
)

func listenUdp() {
	packetConn, err := net.ListenPacket("udp", ss.Conf.Local)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		packetConn.Close()
		os.Exit(0)
	}()

	for {
		buf := make([]byte, 4096)
		_, addr, err := packetConn.ReadFrom(buf)
		if err != nil {
			log.Println(err)
		} else {
			go handleClientUdp(packetConn, addr, buf)
		}
	}
}

func handleClientUdp(packetConn net.PacketConn, addr net.Addr, buf []byte) {

}
