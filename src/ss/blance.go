package ss

import (
	"fmt"
	"math/rand"
	"net"
	"time"
)

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

// 均衡负载 随机一个服务器连接
// 如果不可用启用备用服务器
func BalanceServer(servers []ServerConfig, netType string) (ServerConfig, error) {
	var server ServerConfig

	randomServer, backup := RandomServer(servers)

	serverConn, err := net.DialTimeout(netType, randomServer.Addr, TIMEOUT)
	if err != nil && backup != (ServerConfig{}) {
		backupConn, errb := net.DialTimeout(netType, backup.Addr, TIMEOUT)
		if errb != nil {
			err = errb
		}
		server = backup
		backupConn.Close()
	} else {
		server = randomServer
		serverConn.Close()
	}

	return server, nil
}

// 随机服务器
func RandomServer(sarr []ServerConfig) (oneServer ServerConfig, backup ServerConfig) {
	l := len(sarr)
	if l <= 1 {
		oneServer = sarr[0]
		return
	}

	var warr []Weight

	for k, i := range sarr {
		switch t := i.Weight.(type) {
		case string:
			if t == "backup" {
				backup = i
			}
		case int:
			if t > 0 {
				w := Weight{
					k,
					t,
				}
				warr = append(warr, w)
			}
		}
	}

	r := WeightRandom(warr) // 返回下标
	oneServer = sarr[r]
	return
}
