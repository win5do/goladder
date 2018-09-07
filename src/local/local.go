package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"goladder/src/ss"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
)

func main() {
	config := ss.CliFlag("./local_config.json")
	configStruct := ss.ParseConfigFile(config)
	ListenLocal(configStruct)
}

func ListenLocal(config ss.Config) {
	listener, err := net.Listen("tcp", config.Local)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
		} else {
			go handleClientConn(conn, config)
		}
	}
}

// 处理客户端tcp连接
func handleClientConn(client net.Conn, config ss.Config) {
	defer client.Close()
	serverConf, server, err := balanceDial(config.Server)
	if err != nil {
		log.Println(err)
		return
	}
	defer server.Close()
	var sserver *ss.Sconn

	// 偷看前两位确定是socks5还是http
	clientRd := bufio.NewReader(client)
	peekBuf, err := clientRd.Peek(2)
	if err != nil {
		log.Println(err)
		return
	}

	/*
		+----+----------+----------+
		|VER | NMETHODS | METHODS  |
		+----+----------+----------+
		| 1  |    1     | 1 to 255 |
		+----+----------+----------+
	*/
	if peekBuf[0] == 5 {
		// socks5

		// 把剩余的读出来
		total := int(peekBuf[1]) + 2
		buf := make([]byte, total)
		_, err = io.ReadFull(clientRd, buf)
		if err != nil {
			log.Println(err)
			return
		}

		/*
			+----+--------+
			|VER | METHOD |
			+----+--------+
			| 1  |   1    |
			+----+--------+
		*/

		// 不需要验证
		_, err = client.Write([]byte{5, 0})
		if err != nil {
			log.Println(err)
			return
		}

		// 和服务器建立ss.Sconn
		sserver, err = ss.InitSconn(server, serverConf.Password)
		if err != nil {
			log.Println(err)
			return
		}
	} else {
		// http
		req, err := http.ReadRequest(clientRd)
		if err != nil {
			log.Println(err)
			return
		}

		// 和服务器建立ss.Sconn
		sserver, err = ss.InitSconn(server, serverConf.Password)
		if err != nil {
			log.Println(err)
			return
		}

		err = handleClientHttp(req, sserver)
		if err != nil {
			log.Println(err)
			return
		}
	}

	// 双向转发
	go func() {
		_, err = ss.DecryptCopy(client, sserver)
		if err != nil {
			sserver.Close()
			client.Close()
			if err != io.EOF {
				log.Println(err)
			}
			return
		}
	}()

	_, err = ss.EncryptCopy(sserver, client)
	if err != nil {
		if err != io.EOF {
			log.Println(err)
		}
		return
	}
}

// 转发http 把http包装成socks5协议
func handleClientHttp(req *http.Request, sserver *ss.Sconn) (err error) {
	// host不含端口 可能为domain、ip
	host := req.URL.Hostname()
	hostType, err := ss.HostType(host)
	if err != nil {
		return
	}

	var hostBuf []byte
	if hostType == "domain" {
		l := uint8(len(host))
		hostBuf = []byte{3, l}
		hostBuf = append(hostBuf, []byte(host)...)
	} else if hostType == "ipv4" || hostType == "ipv6" {
		ipAddr, er := net.ResolveIPAddr("ip", host)
		if er != nil {
			err = er
			return
		}
		hostBuf = ipAddr.IP
	}

	portStr := req.URL.Port()
	port := 80
	// 默认端口为80 portStr == ""
	if portStr != "" {
		port, err = strconv.Atoi(portStr)
		if err != nil {
			return
		}
	}
	portBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(portBuf, uint16(port))

	socksBuf := []byte{5, 1, 0}
	socksBuf = append(socksBuf, hostBuf...)
	socksBuf = append(socksBuf, portBuf...)

	/**
	+----+-----+-------+------+----------+----------+
	|VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
	+----+-----+-------+------+----------+----------+
	| 1  |  1  | X'00' |  1   | Variable |    2     |
	+----+-----+-------+------+----------+----------+
	*/
	log.Println("连接信息：", socksBuf)
	_, err = sserver.EncryptWrite(socksBuf)
	if err != nil {
		return
	}

	// 服务端回应 无用 不需要转发
	replyBuf := make([]byte, 10)
	_, err = sserver.DecryptReadFull(replyBuf)
	if err != nil {
		return
	}

	// 将req加密转发到服务端
	reqWt := bytes.NewBuffer([]byte{})
	err = req.WriteProxy(reqWt)
	if err != nil {
		return
	}
	_, err = sserver.EncryptWrite(reqWt.Bytes())
	if err != nil {
		return
	}

	return
}

// 均衡负载 随机一个服务器连接
// 如果不可用启用备用服务器
func balanceDial(config []ss.ServerConfig) (oneServer ss.ServerConfig, conn net.Conn, err error) {
	randomServer, backup := randomServer(config)
	serverConn, err := net.DialTimeout("tcp", randomServer.Addr, ss.TIMEOUT)
	if err != nil && backup != (ss.ServerConfig{}) {
		backupConn, errb := net.DialTimeout("tcp", backup.Addr, ss.TIMEOUT)
		if errb != nil {
			err = errb
		}
		oneServer = backup
		conn = backupConn
	} else {
		oneServer = randomServer
		conn = serverConn
	}
	return
}

// 随机服务器
func randomServer(sarr []ss.ServerConfig) (oneServer ss.ServerConfig, backup ss.ServerConfig) {
	l := len(sarr)
	if l <= 1 {
		oneServer = sarr[0]
		return
	}

	var warr []ss.Weight

	for k, i := range sarr {
		switch t := i.Weight.(type) {
		case string:
			if t == "backup" {
				backup = i
			}
		case int:
			if t > 0 {
				w := ss.Weight{
					k,
					t,
				}
				warr = append(warr, w)
			}
		}
	}

	r := ss.WeightRandom(warr) // 返回下标
	oneServer = sarr[r]
	return
}
