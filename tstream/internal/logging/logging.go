package logging

import (
	"log"
	"os"
)

func Config(dest, prefix string) {
	f, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(f)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetPrefix(prefix)
}
