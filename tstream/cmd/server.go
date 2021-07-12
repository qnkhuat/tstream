/*
Roles of server:
- Receive the stdout of users terminal then broadcast to all users via websocket
*/

package main

import (
	"flag"
	"fmt"
	"github.com/qnkhuat/tstream/internal/logging"
	"github.com/qnkhuat/tstream/pkg/server"
	"log"
)

func main() {
	logging.Config("/tmp/tstream.log", "SERVER: ")
	var db_path = flag.String("db", ".db", "Path to database")
	var host = flag.String("host", "localhost:3000", "Host address to serve server")

	flag.Parse()

	s, err := server.New(*host, *db_path)
	if err != nil {
		fmt.Printf("Failed to create server: %s", err)
		log.Printf("Failed to create server: %s", err)
		return
	}
	s.Start()
	return
}
