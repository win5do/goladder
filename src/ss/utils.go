package ss

import (
	"regexp"
	"errors"
)

// 判断host的类型 host不包含端口
func hostType(host string) (string, error) {
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
