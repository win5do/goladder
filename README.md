# goladder

[![Build Status](https://travis-ci.org/win5do/goladder.svg?branch=master)](https://travis-ci.org/win5do/goladder)

重复造轮子，用go实现的ss程序，用于科学上网

本项目仅供技术学习，请勿用作其他用途

# 安装
```shell
cd $GOPATH
git clone https://github.com/win5do/goladder.git
# 本地端
go install goladder/src/local
# 服务器端
go install goladder/src/server
```


# 参数
local和server只有一个config参数，用于指定配置文件

默认为./local_config.json和./server_config.json

```
local -config=./local_config.json

server -config=./server_config.json
```

# 配置
配置示例见cmd/local及cmd/server下json配置文件

local端配置
```
{
  "local": ":8888", // 本地监听端口，同时支持socks5和http
  "server": [  // server可配置多个
    {
      "addr": "127.0.0.1:9998",
      "password": "123456", // 密码需要与服务器相同
      "weight": 60  // 权重，用于多服务器负载均衡，单服务器可忽略，默认为100
    },
    {
      "addr": "127.0.0.1:9999",
      "password": "123456",
      "weight": "backup" // 权重设为backup时为备用服务器，主服务器挂了启用备用服务器
    }
  ]
}
```
server端配置

```
{
  "server": [ // 可同时监听多个端口
    {
      "addr": ":9998", // 只需填写:端口
      "password": "123456" // 密码需要与local对应
    },
    {
      "addr": ":9999",
      "password": "123456"
    }
  ]
}
```


# todo

- udp转发
- pac过滤