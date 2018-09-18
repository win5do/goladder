package ss

import (
	"testing"
	"log"
)

const (
	FILE1 = "../../cmd/local/local_config.json"
	FILE2 = "../../cmd/server/server_config.json"
)

func TestParseConfigFile(t *testing.T) {
	c1 := ParseConfigFile(FILE1)
	log.Print(c1)
	c2 := ParseConfigFile(FILE2)
	log.Print(c2)
}
