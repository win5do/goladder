package server

import (
	"crypto/aes"
	"io"
	"log"
	"net"

	"goladder/src/ss"
)

func ListenServer() {
	for _, i := range ss.Conf.Server {
		go listenTcp(i)
		if ss.Conf.Udp {
			go listenUdp(i)
		}
	}
	ss.WaitSignal()
}

func listenTcp(oneServer ss.ServerConfig) {
	listener, err := net.Listen("tcp", oneServer.Addr)
	if err != nil {
		log.Println(err)
		return
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
		} else {
			go handleTcpConn(conn, oneServer)
		}
	}
}

// 处理tcp连接
func handleTcpConn(client net.Conn, oneServer ss.ServerConfig) {
	defer client.Close()

	iv := make([]byte, aes.BlockSize)
	_, err := io.ReadFull(client, iv)
	if err != nil {
		log.Println(err)
		return
	}

	sclient, err := ss.NewSconn(client, oneServer.Password, iv)
	if err != nil {
		log.Println(err)
		return
	}

	/**
		+------+----------+----------+
		| ATYP | DST.ADDR | DST.PORT |
		+------+----------+----------+
		|  1   | Variable |    2     |
		+------+----------+----------+
	*/

	buf := make([]byte, 2) // 先读2位 如果为域名 buf[1]为域名长度
	_, err = sclient.DecryptReadFull(buf)
	if err != nil {
		log.Println(err)
		return
	}

	hostType := buf[0]
	remain := ss.ParseSocksRemain(buf, hostType)

	// 根据剩余长度读完剩余 合并到buf
	remainBuf := make([]byte, remain)
	_, err = sclient.DecryptReadFull(remainBuf)
	if err != nil {
		log.Println(err)
		return
	}
	buf = append(buf, remainBuf...)
	log.Printf("read buf = %v", buf)

	dstAddr := ss.ParseSocksAddr(buf, hostType)

	dst, err := net.DialTimeout("tcp", dstAddr, ss.TIMEOUT)
	if err != nil {
		// 连接远程服务器失败
		log.Println(err)
		return
	}
	defer dst.Close()

	// 双向转发
	go func() {
		_, err := ss.EncryptCopy(sclient, dst)
		if err != nil {
			dst.Close()
			sclient.Close()
			if err != io.EOF {
				log.Println(err)
			}
			return
		}
	}()
	_, err = ss.DecryptCopy(dst, sclient)
	if err != nil && err != io.EOF {
		if err != io.EOF {
			log.Println(err)
		}
		return
	}
}
