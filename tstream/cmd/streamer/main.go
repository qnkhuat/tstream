/*
Inner working of streamer program
- When a streamer start his streaming
	1. We open up a pty, in this pty we run a shell program (pty.StartShell)
	- It's easier to think that we now have a client(the pty) and server (the shell program). The roles of each is:
		- Client is just a text based displayer
		- Server is what process users input. For example: run vim, execute scripts, cat, grep ...
	- In order for users to use the shell like a regular shell we have to:
		- Forward all users input to the server : pty -> os.stdin
		- Forward server output to the client: os.stdout -> pty
	- This is infact how our terminal work as well, the terminal program is just the client. and the shell in os is what process the client inputs
	2. Open a websocket connection to server
	- Forward pty stdout to server via a websocket
	- This will then be broadcast to viewers to display the terminal
*/

package main

import (
	"flag"
	"github.com/qnkhuat/tstream/internal/logging"
	"github.com/qnkhuat/tstream/pkg/streamer"
	"log"
)

func main() {
	logging.Config("/tmp/tstream.log", "STREAMER: ")
	log.Println("Streamer started")
	var server = flag.String("server", "0.0.0.0:3000", "Server endpoint")
	var session = flag.String("session", "qnkhuat", "Session name")
	flag.Parse()

	log.Printf("Got session: %s\n", *session)
	s := streamer.New(*server, *session)
	err := s.Start()
	if err != nil {
		log.Panicf("Failed to start stream: %s", err)
	}
}
