package ss

import (
	"net"
	"time"
	"log"
	"fmt"
	"math/rand"
	"io"
	"bufio"
	"regexp"
)

const (
	TIMEOUT = 5 * time.Second
)

func ListenClient(config Config) {
	listener, err := net.Listen("tcp", config.Client)
	defer listener.Close()
	if err != nil {
		log.Fatal(err)
	}

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
	defer server.Close()
	LogErr(err)

	sserver := newSconn(server, oneServer.Password)
	sclient := newSconn(client, oneServer.Password)

	buf := make([]byte, 256)

	n, err := io.ReadAtLeast(sclient, buf, 2)
	if err != nil {
		LogErr(err)
	}

	if buf[0] == 5 {
		// socks5

		// 把剩余的读出来
		total := int(buf[1]) + 2
		if n < total {
			_, err = io.ReadAtLeast(sclient, buf, total-2)
			LogErr(err)
		}

		// 不需要验证
		_, err = sclient.Write([]byte{5, 0})
		LogErr(err)

		// 双向转发
		go func() {
			_, err = sserver.decryptCopy(sclient)
			LogErr(err)
		}()

		_, err = sclient.encryptCopy(sserver)
		LogErr(err)
	} else {
		// http

		err := sclient.SetReadDeadline(time.Now().Add(time.Second))
		if err != nil {
			return
		}

		bufrd := bufio.NewReader(sclient)

		// GET / HTTP/1.1
		line1, err := bufrd.ReadBytes('\n')
		if err != nil {
			return
		}
		ok, err := regexp.Match(`(?i)HTTP`, line1)
		if !ok || err != nil {
			return
		}

		// host: google.com
		//line2, err := bufrd.ReadBytes('\n')
		//if err != nil {
		//	return
		//}
		//ok, err = regexp.Match(`(?i)host:`, line1)
		//if !ok || err != nil {
		//	return
		//}
	}

	if string(buf[:3]) == "HTT" {
		// http请求
	} else if buf[0] == 5 {
		// socks5
		// 不需要验证
		_, err = sclient.Write([]byte{5, 0})
		LogErr(err)

		// 双向转发
		go func() {
			_, err = sserver.decryptCopy(sclient)
			LogErr(err)
		}()

		_, err = sclient.encryptCopy(sserver)
		LogErr(err)
	} else {
		return
	}
}

// 均衡负载 随机一个服务器连接
// 如果不可用启用备用服务器
func balanceDial(config []ServerConfig) (oneServer ServerConfig, conn net.Conn, err error) {
	randomServer, backup := weightRandom(config)
	serverConn, err := net.DialTimeout("tcp", randomServer.Adr, TIMEOUT)
	if err != nil && backup != (ServerConfig{}) {
		backupConn, err := net.DialTimeout("tcp", backup.Adr, TIMEOUT)
		if err != nil {
			log.Println("无可用服务器:", err)
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
	for _, i := range w {
		if i.Weight.(string) == "backup" {
			backup = i
		} else if i.Weight.(int) > 0 {
			sum += i.Weight.(int)
		}
	}

	seed := rand.NewSource(time.Now().UnixNano())
	rd := rand.New(seed)
	r := rd.Float64()
	r *= float64(sum)
	fmt.Println("random:", r)

	scale := 0
	for _, i := range w {
		scale += i.Weight.(int)
		if r < float64(scale) {
			oneServer = i
			return
		}
	}

	oneServer = w[len(w)-1]
	return
}
