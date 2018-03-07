package core

import (
	"net"
	"time"
	"log"
	"fmt"
	"math/rand"
)

const TIMEOUT = 5 * time.Second

func ListenClient(config Config) {
	listener, err := net.Listen("tcp", config.client)
	LogFatal(err)
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
	// 均衡负载 随机一个
	oneServer := weightRandom(config.server)
	server, err := net.DialTimeout("tcp", oneServer.adr, TIMEOUT)
	defer server.Close()
	LogErr(err)

	sserver := newSconn(server, oneServer.password)
	sclient := newSconn(client, oneServer.password)

	go func() {
		_, err = sserver.decryptCopy(sclient)
		LogErr(err)
	}()

	_, err = sclient.encryptCopy(sserver)
	LogErr(err)
}

// 根据权重随机
func weightRandom(w []ServerConfig) ServerConfig {
	l := len(w)

	if l <= 1 {
		return w[0]
	}

	sum := 0

	for _, i := range w {
		sum += i.weight
	}

	seed := rand.NewSource(time.Now().UnixNano())
	rd := rand.New(seed)
	r := rd.Float64()
	r *= float64(sum)
	fmt.Println("random:", r)

	scale := 0
	for _, i := range w {
		scale += i.weight
		if r < float64(scale) {
			return i
		}
	}
	return w[len(w)-1]
}
