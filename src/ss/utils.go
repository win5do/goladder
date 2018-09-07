package ss

import (
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"time"
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

type Weight struct {
	Value  int
	Weight int
}

// 根据权重随机
func WeightRandom(w []Weight) int {
	l := len(w)
	if l < 1 {
		return 0
	}

	if l == 1 {
		return w[0].Value
	}

	sum := 0

	for _, i := range w {
		sum += i.Weight
	}

	seed := rand.NewSource(time.Now().UnixNano())
	rd := rand.New(seed)
	r := rd.Float64()
	r *= float64(sum)
	fmt.Println("random:", r)

	scale := 0
	for _, i := range w {
		scale += i.Weight
		if r < float64(scale) {
			return i.Value
		}
	}
	return w[len(w)-1].Value
}
