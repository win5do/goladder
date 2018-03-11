package ss

import (
	"net"
	"time"
	"log"
	"math/rand"
	"io"
	"bufio"
	"net/http"
	"bytes"
	"io/ioutil"
	"encoding/binary"
	"strconv"
)

const (
	TIMEOUT = 3 * time.Second
)

func ListenClient(config Config) {
	listener, err := net.Listen("tcp", config.Client)
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

// 处理客户端连接
func handleClientConn(client net.Conn, config Config) {
	defer client.Close()
	oneServer, server, err := balanceDial(config.Server)
	if err != nil {
		log.Println(err)
		return
	}
	defer server.Close()

	clientRd := bufio.NewReader(client)
	peekBuf, err := clientRd.Peek(2)
	if err != nil {
		log.Println(err)
		return
	}

	iv := randIv()
	sserver, err := newSconn(server, oneServer.Password, iv)
	if err != nil {
		log.Println(err)
		return
	}
	_, err = sserver.Write(iv)
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
	} else {
		// http
		req, err := http.ReadRequest(clientRd)
		if err != nil {
			log.Println(err)
			return
		}

		// host不含端口 可能为domain、ip
		host := req.URL.Hostname()
		hostType, err := hostType(host)
		if err != nil {
			log.Println(err)
			return
		}

		var hostBuf []byte
		if hostType == "domain" {
			l := uint8(len(host))
			hostBuf = []byte{3, l}
			hostBuf = append(hostBuf, []byte(host)...)
		} else if hostType == "ipv4" || hostType == "ipv6" {
			ipAddr, err := net.ResolveIPAddr("ip", host)
			if err != nil {
				log.Println(err)
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
				log.Println(err)
				return
			}
		}
		portBuf := make([]byte, 2)
		binary.BigEndian.PutUint16(portBuf, uint16(port))

		socksBuf := []byte{5, 0, 0}
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
		_, err = sserver.encryptWrite(socksBuf)
		if err != nil {
			log.Println(err)
			return
		}

		// 服务端回应 无用 不需要转发
		replyBuf := make([]byte, 10)
		_, err = sserver.decryptReadFull(replyBuf)
		if err != nil {
			log.Println(err)
			return
		}

		// 将req加密转发
		reqWt := bytes.NewBuffer([]byte{})
		err = req.WriteProxy(reqWt)
		if err != nil {
			log.Println(err)
			return
		}
		reqBuf, err := ioutil.ReadAll(reqWt)
		if err != nil {
			log.Println(err)
			return
		}
		_, err = sserver.encryptWrite(reqBuf)
		if err != nil {
			log.Println(err)
			return
		}
	}

	// 双向转发
	go func() {
		_, err = decryptCopy(client, sserver)
		if err != nil {
			sserver.Close()
			client.Close()
			log.Println(err)
			return
		}
	}()

	_, err = encryptCopy(sserver, client)
	if err != nil {
		log.Println(err)
		return
	}
}

// 均衡负载 随机一个服务器连接
// 如果不可用启用备用服务器
func balanceDial(config []ServerConfig) (oneServer ServerConfig, conn net.Conn, err error) {
	randomServer, backup := weightRandom(config)
	serverConn, err := net.DialTimeout("tcp", randomServer.Addr, TIMEOUT)
	if err != nil && backup != (ServerConfig{}) {
		backupConn, errb := net.DialTimeout("tcp", backup.Addr, TIMEOUT)
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

// 根据权重随机
func weightRandom(w []ServerConfig) (oneServer ServerConfig, backup ServerConfig) {
	l := len(w)
	if l <= 1 {
		oneServer = w[0]
		return
	}

	sum := 0
	wint := []ServerConfig{}

	for _, i := range w {
		switch q := i.Weight.(type) {
		case string:
			if q == "backup" {
				backup = i
			}
		case int:
			if q > 0 {
				sum += q
				wint = append(wint, i)
			}
		}
	}

	seed := rand.NewSource(time.Now().UnixNano())
	rd := rand.New(seed)
	r := rd.Float64()
	r *= float64(sum)

	scale := 0
	for _, i := range wint {
		scale += i.Weight.(int)
		if r < float64(scale) {
			oneServer = i
			return
		}
	}

	oneServer = w[len(w)-1]
	return
}
