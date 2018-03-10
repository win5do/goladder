package ss

import (
	"log"
	"os"
	"io"
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func logErr(err error) {
	if err != io.EOF {
		log.Println(err)
	}
}
