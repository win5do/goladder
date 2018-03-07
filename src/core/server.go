package core

import (
	"encoding/binary"
	"net"
	"log"
)

func ListenServer(config Config) {
	for _, i := range config.server {
		go listenServer(i)
	}
}

func listenServer(oneServer ServerConfig) {
	listener, err := net.Listen("tcp", oneServer.adr)
	LogErr(err)
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
		} else {
			sconn := newSconn(conn, oneServer.password)
			go handleServerConn(sconn, oneServer)
		}
	}
}

// 处理服务端连接
func handleServerConn(sserver sconn, oneServer ServerConfig) {
	defer sserver.Close()
	buf := make([]byte, 256)
	_, err := sserver.decryptRead(buf)
	LogErr(err)

	if buf[0] != 5 {
		// 只支持socks5
		return
	}
	// 不需要验证
	_, err = sserver.encryptWrite([]byte{5, 0})
	LogErr(err)

	/**
		+----+-----+-------+------+----------+----------+
		|VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
		+----+-----+-------+------+----------+----------+
		| 1  |  1  | X'00' |  1   | Variable |    2     |
		+----+-----+-------+------+----------+----------+
	*/
	n, err := sserver.decryptRead(buf)
	LogErr(err)
	if buf[1] != 1 {
		// 目前只支持 CONNECT
		return
	}
	var ip []byte
	// aType 代表请求的远程服务器地址类型,值长度1个字节,有三种类型
	switch buf[3] {
	case 1:
		//	IP V4 address: X'01'
		ip = buf[4 : 4+net.IPv4len]
	case 3:
		//	DOMAINNAME: X'03'
		ipAddr, err := net.ResolveIPAddr("ip", string(buf[5:n-2]))
		if err != nil {
			return
		}
		ip = ipAddr.IP
	case 4:
		//	IP V6 address: X'04'
		ip = buf[4 : 4+net.IPv6len]
	default:
		return
	}
	port := buf[n-2:]
	dstAddr := &net.TCPAddr{
		IP:   ip,
		Port: int(binary.BigEndian.Uint16(port)),
	}
	// 响应客户端连接成功
	/**
		+----+-----+-------+------+----------+----------+
		|VER | REP |  RSV  | ATYP | BND.ADDR | BND.PORT |
		+----+-----+-------+------+----------+----------+
		| 1  |  1  | X'00' |  1   | Variable |    2     |
		+----+-----+-------+------+----------+----------+
	*/
	_, err = sserver.encryptWrite([]byte{5, 0, 0, 1, 0, 0, 0, 0, 0, 0})
	LogErr(err)

	// 连接真正的远程服务
	dst, err := net.DialTimeout("tcp", dstAddr.String(), TIMEOUT)
	defer dst.Close()
	LogErr(err)
	sdst := newSconn(dst, oneServer.password)

	// 进行转发
	go func() {
		_, err := sdst.encryptCopy(sserver)
		LogErr(err)
	}()
	_, err = sserver.decryptCopy(sdst)
	LogErr(err)
}
