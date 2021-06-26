package main

import (
	//"flag"
	"fmt"
	"github.com/qnkhuat/tstream/pkg/streamer"
	"io"
	"os"
)

func main() {
	var server = flag.String("server", "0.0.0.0:3000", "Server endpoint")
	s := streamer.New(server)
	err := s.Start()
	if err != nil {
		log.Panicf("Failed to start stream: %s", err)
	}

	pty := ptyMaster.New()
	pty.StartShell()
	fmt.Println("yo man")
	go func() { io.Copy(pty.F(), os.Stdin) }()
	io.Copy(os.Stdout, pty.F())
}
