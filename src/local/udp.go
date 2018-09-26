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
		// udp头最短10位
		if err != nil || len(buf) < 10 {
			log.Println(err)
		} else {
			go handleClientUdp(packetConn, addr, buf)
		}
	}
}

func handleClientUdp(clientConn net.PacketConn, clientAddr net.Addr, buf []byte) {
	serverConf, err := ss.BalanceServer(ss.Conf.Server, "udp")
	if err != nil {
		log.Println(err)
		return
	}

	server := ss.UdpNat.GenConn(clientAddr.String(), serverConf.Addr)
	if server == nil {
		return
	}
	defer server.Close()

	sserver, err := ss.InitSconn(server, serverConf.Password)
	if err != nil {
		log.Println(err)
		return
	}

	_, err = sserver.EncryptWrite(buf[3:])
	if err != nil {
		log.Println(err)
		return
	}

	
}

func piplineUdp()  {
	// 读客户端服务端消息发给
}


