package local

import (
	"bytes"
	"encoding/binary"
	"log"
	"net"
	"net/http"
	"strconv"

	"goladder/src/ss"
)

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
