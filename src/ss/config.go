package ss

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
)

type Config struct {
	Local  string
	Server []ServerConfig
	Udp    bool
}

type ServerConfig struct {
	Addr     string
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

	Conf = config
	return config
}

func CliFlag(config string) string {
	flag.StringVar(&config, "config", config, "配置文件相对地址，默认为"+config)
	flag.Parse()
	return config
}
