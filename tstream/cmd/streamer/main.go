package main

import (
	"flag"
	"github.com/qnkhuat/tstream/internal/logging"
	"github.com/qnkhuat/tstream/pkg/streamer"
	"log"
)

func main() {
	logging.Config("/tmp/tstream.log", "STREAMER: ")
	log.Println("YOOOOOOOOOOOOOO")
	var server = flag.String("server", "0.0.0.0:3000", "Server endpoint")
	var sessionID = "qnkhuat"
	s := streamer.New(*server, sessionID)
	err := s.Start()
	if err != nil {
		log.Panicf("Failed to start stream: %s", err)
	}
}
