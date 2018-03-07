package core

type Config struct {
	client string
	server []ServerConfig
}

type ServerConfig struct {
	adr      string
	password string
	weight	int
}
