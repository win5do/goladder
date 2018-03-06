package core

import (
	"net"
	"log"
	"fmt"
	"encoding/binary"
)

type Socks struct {
	Key    string
	Listen string // 监听端口
	Remote string
}

func (s Socks) ListenClient() {
	listen(s, handleClient)
}

func (s Socks) ListenServer() {
	listen(s, handleServer)
}

func listen(s Socks, cb func(sconn, Socks)) {
	listener, err := net.Listen("tcp", s.Listen)
	LogFatal(err)
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
		} else {
			sconn := newSconn(s.Key, conn)
			go cb(sconn, s)
		}
	}
}

// 处理客户端连接
func handleClient(sc sconn, s Socks) {
	defer sc.Close()
	remote, err := net.Dial("tcp", s.Remote)
	defer remote.Close()
	LogErr(err)

	sremote := newSconn(s.Key, remote)

	go func() {
		_, err = sremote.decryptCopy(sc)
		LogErr(err)
	}()

	_, err = sc.encryptCopy(sremote)
	LogErr(err)
}

// 处理服务端连接
func handleServer(sc sconn, s Socks) {
	defer sc.Close()
	buf := make([]byte, 256)
	_, err := sc.decryptRead(buf)
	LogErr(err)
	fmt.Println("内容：", buf)

	if buf[0] != 5 {
		// 只支持socks5
		return
	}
	// 不需要验证
	_, err = sc.encryptWrite([]byte{5, 0})
	LogErr(err)

	/**
		+----+-----+-------+------+----------+----------+
		|VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
		+----+-----+-------+------+----------+----------+
		| 1  |  1  | X'00' |  1   | Variable |    2     |
		+----+-----+-------+------+----------+----------+
	*/
	n, err := sc.decryptRead(buf)
	LogErr(err)
	fmt.Println("连接：", buf)
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
	_, err = sc.encryptWrite([]byte{5, 0, 0, 1, 0, 0, 0, 0, 0, 0})
	LogErr(err)

	fmt.Println(dstAddr.String())
	// 连接真正的远程服务
	dst, err := net.Dial("tcp", dstAddr.String())
	defer dst.Close()
	LogErr(err)
	sdst := newSconn(s.Key, dst)

	// 进行转发
	go func() {
		_, err := sdst.encryptCopy(sc)
		LogErr(err)
	}()
	_, err = sc.decryptCopy(sdst)
	LogErr(err)
}
