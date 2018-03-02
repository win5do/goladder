package core

import (
	"net"
	"io"
	"fmt"
	"log"
	"encoding/binary"
)

type Proxy struct {
	src net.Addr
	dst net.Addr
}

func NewProxy(src, dst net.Addr) *Proxy {
	return &Proxy{
		src: src,
		dst: dst,
	}
}

func Listen(port string, cb func(*net.Conn)) {
	listener, err := net.Listen("tcp", ":"+port)
	LogFatal(err)
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
		} else {
			go cb(&conn)
		}
	}
}

func HandleClient(conn *net.Conn) {
	//socks := core.NewSocks(conn)
	c := *conn
	remote, err := net.Dial("tcp", ":9999")
	LogErr(err)
	defer remote.Close()

	go io.Copy(remote, c)
	io.Copy(c, remote)
}

func HandleServer(conn *net.Conn) {
	c := *conn
	defer c.Close()
	buf := make([]byte, 256)
	_, err := c.Read(buf)
	LogErr(err)
	fmt.Println("内容：", buf)

	if buf[0] != 5 {
		// 只支持socks5
		return
	}
	// 不需要验证
	c.Write([]byte{5, 0})

	n, err := c.Read(buf)
	LogErr(err)
	fmt.Println("连接：", buf)
	if buf[1] != 0x01 {
		// 目前只支持 CONNECT
		return
	}
	var ip []byte
	// aType 代表请求的远程服务器地址类型，值长度1个字节，有三种类型
	switch buf[3] {
	case 0x01:
		//	IP V4 address: X'01'
		ip = buf[4 : 4+net.IPv4len]
	case 0x03:
		//	DOMAINNAME: X'03'
		ipAddr, err := net.ResolveIPAddr("ip", string(buf[5:n-2]))
		if err != nil {
			return
		}
		ip = ipAddr.IP
	case 0x04:
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
	c.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

	fmt.Println(dstAddr.String())
	// 连接真正的远程服务
	dst, err := net.Dial("tcp", dstAddr.String())
	defer dst.Close()
	LogErr(err)

	// 进行转发
	go func() {
		io.Copy(dst, c)
	}()
	io.Copy(c, dst)
}
