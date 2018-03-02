package core

import (
	"log"
	"os"
)

func init() {
	log.SetOutput(os.Stdout)
}

func LogFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func LogErr(err error) {
	if err != nil {
		log.Println(err)
		return
	}
}
