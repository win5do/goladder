package ss

import (
	"encoding/binary"
	"net"
	"log"
	"sync"
	"crypto/aes"
	"io"
	"strconv"
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
			go handleServerConn(conn, oneServer)
		}
	}
}

// 处理服务端连接
func handleServerConn(client net.Conn, oneServer ServerConfig) {
	defer client.Close()

	iv := make([]byte, aes.BlockSize)
	_, err := io.ReadFull(client, iv)
	if err != nil {
		log.Println(err)
		return
	}

	sclient, err := newSconn(client, oneServer.Password, iv)
	if err != nil {
		log.Println(err)
		return
	}

	/**
		+----+-----+-------+------+----------+----------+
		|VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
		+----+-----+-------+------+----------+----------+
		| 1  |  1  | X'00' |  1   | Variable |    2     |
		+----+-----+-------+------+----------+----------+
	*/
	buf := make([]byte, 5) // 先读5位 如果为域名 buf[4]为域名长度
	_, err = sclient.decryptReadFull(buf)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("debug:", buf)

	reqType := "" // 请求类型

	if buf[1] == 1 {
		// tcp
		reqType = "tcp"
	} else if buf[1] == 3 {
		// udp
		reqType = "udp"
	} else {
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
	_, err = sclient.decryptReadFull(remainBuf)
	if err != nil {
		log.Println(err)
		return
	}
	buf = append(buf, remainBuf...)
	log.Println("debug:", buf)

	var ip net.IP
	var host string

	if hostType == 1 {
		// ipv4
		ip = buf[4:8]
		host = ip.String()
	} else if hostType == 3 {
		// domain
		host = string(buf[5 : 5+buf[4]])
	} else if hostType == 4 {
		// ipv6
		ip = buf[4:20]
		host = ip.String()
	} else {
		return
	}

	portInt := int(binary.BigEndian.Uint16(buf[len(buf)-2:]))
	port := strconv.Itoa(portInt)

	dstAddr := host + ":" + port

	// 连接真正的远程服务
	// 响应客户端连接结果
	/**
		+----+-----+-------+------+----------+----------+
		|VER | REP |  RSV  | ATYP | BND.ADDR | BND.PORT |
		+----+-----+-------+------+----------+----------+
		| 1  |  1  | X'00' |  1   | Variable |    2     |
		+----+-----+-------+------+----------+----------+
	*/

	if reqType == "tcp" {
		dst, err := net.DialTimeout("tcp", dstAddr, TIMEOUT)
		if err != nil {
			// 连接远程服务器失败
			log.Println(err)
			return
		}
		defer dst.Close()

		// 连接远程服务器成功
		_, err = sclient.encryptWrite([]byte{5, 0, 0, 1, 0, 0, 0, 0, 0, 0})
		if err != nil {
			log.Println(err)
			return
		}

		// 双向转发
		go func() {
			_, err := encryptCopy(sclient, dst)
			if err != nil {
				dst.Close()
				sclient.Close()
				if err != io.EOF {
					log.Println(err)
				}
				return
			}
		}()
		_, err = decryptCopy(dst, sclient)
		if err != nil && err != io.EOF {
			if err != io.EOF {
				log.Println(err)
			}
			return
		}
	} else if reqType == "udp" {
		// todo
	}
}

// udp转发
func handleServerUdp() {}
