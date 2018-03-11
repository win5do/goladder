package main

import (
	"bufio"
	"net/http"
	"log"
	"net"
	"strconv"
	"encoding/binary"
	"bytes"
	"io/ioutil"
)

func handleClientHttp(clientRd *bufio.Reader, sserver *sconn) (err error) {
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