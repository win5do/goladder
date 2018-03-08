package core

import (
	"encoding/json"
	"log"
	"io/ioutil"
)

type Config struct {
	Client string
	Server []ServerConfig
}

type ServerConfig struct {
	Adr      string
	Password string
	Weight   interface{}
}

func ParseConfigFile(filePath string) Config {
	fileByte, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal("打开配置文件出错:", err)
	}

	config := Config{}
	err = json.Unmarshal(fileByte, &config)
	if err != nil {
		log.Fatal("配置文件格式错误:", err)
	}

	servers := config.Server
	if len(servers) < 1 {
		log.Fatal("至少提供一个服务器")
	}

	for k, i := range servers {
		if i.Weight == nil {
			servers[k].Weight = 100
		}
	}
	return config
}
