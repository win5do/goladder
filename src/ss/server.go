package ss

import (
	"encoding/binary"
	"net"
	"log"
	"sync"
)

var wg sync.WaitGroup

func ListenServer(config Config) {
	for _, i := range config.Server {
		wg.Add(1)
		go listenServer(i)
	}
	wg.Wait()
}

func listenServer(oneServer ServerConfig) {
	defer wg.Done()
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
			sconn := newSconn(conn, oneServer.Password)
			go handleServerConn(sconn, oneServer)
		}
	}
}

// 处理服务端连接
func handleServerConn(sserver sconn, oneServer ServerConfig) {
	defer sserver.Close()

	/**
		+----+-----+-------+------+----------+----------+
		|VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
		+----+-----+-------+------+----------+----------+
		| 1  |  1  | X'00' |  1   | Variable |    2     |
		+----+-----+-------+------+----------+----------+
	*/
	buf := make([]byte, 5) // 先读5位 如果为域名 buf[4]为域名长度
	_, err := sserver.decryptReadFull(buf)
	if err != nil {
		log.Println(err)
		return
	}
	if buf[1] != 1 {
		// 目前只支持 CONNECT
		return
	}

	// aType 代表请求的远程服务器地址类型,值长度1个字节,有三种类型
	hostType := buf[3]
	var remain int // 代理信息剩余长度
	if hostType == 1 {
		// ipv4
		remain = 5 // 4+2-1
	} else if hostType == 3 {
		// domain
		remain = int(buf[4]) + 2
	} else if hostType == 4 {
		// ipv6
		remain = 17 // 16+2-1
	} else {
		return
	}

	log.Println("debug:", hostType, remain)
	// 根据剩余长度读完剩余 合并到buf
	remainBuf := make([]byte, remain)
	_, err = sserver.decryptReadFull(remainBuf)
	if err != nil {
		log.Println(err)
		return
	}
	buf = append(buf, remainBuf...)
	log.Println("debug:", buf)

	var ip []byte
	if hostType == 1 {
		// ipv4
		ip = buf[4:8]
	} else if hostType == 3 {
		// domain
		ipAddr, err := net.ResolveIPAddr("ip", string(buf[5:5+buf[4]]))
		if err != nil {
			return
		}
		ip = ipAddr.IP
	} else if hostType == 4 {
		// ipv6
		ip = buf[4:20]
	} else {
		return
	}

	port := buf[len(buf)-2:]
	dstAddr := &net.TCPAddr{
		IP:   ip,
		Port: int(binary.BigEndian.Uint16(port)),
	}

	// 连接真正的远程服务
	dst, err := net.DialTimeout("tcp", dstAddr.String(), TIMEOUT)
	if err != nil {
		log.Println(err)
		return
	}
	defer dst.Close()

	// 响应客户端连接成功
	/**
		+----+-----+-------+------+----------+----------+
		|VER | REP |  RSV  | ATYP | BND.ADDR | BND.PORT |
		+----+-----+-------+------+----------+----------+
		| 1  |  1  | X'00' |  1   | Variable |    2     |
		+----+-----+-------+------+----------+----------+
	*/
	_, err = sserver.encryptWrite([]byte{5, 0, 0, 1, 0, 0, 0, 0, 0, 0})
	if err != nil {
		log.Println(err)
		return
	}

	sdst := newSconn(dst, oneServer.Password)

	// 双向转发
	go func() {
		_, err := sdst.encryptCopy(sserver)
		if err != nil {
			log.Println(err)
			return
		}
	}()
	_, err = sserver.decryptCopy(sdst)
	if err != nil {
		log.Println(err)
		return
	}
}
