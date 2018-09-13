package local

import (
	"bufio"
	"io"
	"log"
	"net"
	"net/http"
	"os"

	"goladder/src/ss"
)

func ListenLocal() {

	go listenTcp()

	if ss.Conf.Udp {
		go listenUdp()
	}

	ss.WaitSignal()
}

func listenTcp() {
	listener, err := net.Listen("tcp", ss.Conf.Local)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		listener.Close()
		os.Exit(0)
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
		} else {
			go handleClientConn(conn)
		}
	}
}

// 处理客户端tcp连接
func handleClientConn(client net.Conn) {
	defer client.Close()
	serverConf, server, err := ss.BalanceDial(ss.Conf.Server)
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
	} else if ss.IsHttp(clientRd) {
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
	} else {
		return
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
