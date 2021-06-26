package main

import (
	"github.com/qnkhuat/tstream/internal/logging"
	"github.com/qnkhuat/tstream/pkg/server"
)

func main() {
	logging.Config("/tmp/tstream.log", "SERVER: ")
	s := server.New("0.0.0.0:3000")
	s.Start()
	return
}
