package ss

import (
	"bufio"
	"encoding/binary"
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"regexp"
	"strconv"
)

// 判断host的类型 host不包含端口
func HostType(host string) (string, error) {
	regDomain := regexp.MustCompile(`^[a-zA-Z0-9][-a-zA-Z0-9]{0,62}(\.[a-zA-Z0-9][-a-zA-Z0-9]{0,62})+\.?$`)
	regIpv4 := regexp.MustCompile(`^((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$`)
	regIpv6 := regexp.MustCompile(`^([\da-fA-F]{1,4}:){7}[\da-fA-F]{1,4}$`)

	if regIpv4.MatchString(host) {
		return "ipv4", nil
	} else if regDomain.MatchString(host) {
		return "domain", nil
	} else if regIpv6.MatchString(host) {
		return "ipv6", nil
	} else {
		return "", errors.New("invalid host")
	}
}

func ParseSocksReqType(buf []byte) string {
	reqType := "" // 请求类型

	if buf[1] == 1 {
		// tcp
		reqType = "tcp"
	} else if buf[1] == 3 {
		// udp
		reqType = "udp"
	} else {
		return ""
	}

	log.Printf("reqType = %v", reqType)
	return reqType
}

func ParseSocksRemain(buf []byte, hostType byte) int {
	// aType 代表请求的远程服务器地址类型hostType,值长度1个字节,有三种类型
	var remain int // 代理信息剩余长度
	if hostType == 1 {
		// ipv4
		remain = 5 // 4+2-1
	} else if hostType == 3 {
		// domain
		remain = int(buf[1]) + 2
	} else if hostType == 4 {
		// ipv6
		remain = 17 // 16+2-1
	} else {
		return 0
	}

	log.Printf("remain bit = %v", remain)
	return remain
}

func ParseSocksAddr(buf []byte, hostType byte) (string) {
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
		return ""
	}

	portInt := int(binary.BigEndian.Uint16(buf[len(buf)-2:]))
	port := strconv.Itoa(portInt)

	dstAddr := host + ":" + port

	return dstAddr
}

func WaitSignal() {
	var sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan)
	for sig := range sigChan {
		log.Printf("caught signal %v, exit", sig)
		os.Exit(0)
	}
}

func IsHttp(rd *bufio.Reader) bool {
	line := []byte{}

	for i := 1; ; i++ {
		buf, err := rd.Peek(i)
		if err != nil {
			return false
		}

		if buf[i-1] == '\n' {
			line = buf
			break
		}
	}

	ok, err := regexp.Match("HTTP", line)
	if err != nil {
		return false
	}
	return ok
}

func PortBuff(port string) ([]byte, error) {
	portBuf := make([]byte, 2)
	portInt, err := strconv.Atoi(port)
	if err != nil {
		return nil, err
	}
	binary.BigEndian.PutUint16(portBuf, uint16(portInt))

	return portBuf, nil
}
