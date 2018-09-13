package ss

import (
	"bufio"
	"errors"
	"log"
	"os"
	"os/signal"
	"regexp"
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
