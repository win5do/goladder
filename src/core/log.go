package core

import (
	"log"
	"os"
	"io"
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func LogFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func LogErr(err error) {
	if err != nil {
		if err != io.EOF {
			log.Println(err)
		}
		return
	}
}
