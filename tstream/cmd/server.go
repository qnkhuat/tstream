/*
Roles of server:
- Receive the stdout of users terminal then broadcast to all users via websocket
*/

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/qnkhuat/tstream/internal/cfg"
	"github.com/qnkhuat/tstream/internal/logging"
	"github.com/qnkhuat/tstream/pkg/server"
)

func main() {
	logging.Config("/tmp/tstream.log", "SERVER: ")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Just type `server` to turn on server.\n\nAdvanced config:\n")
		flag.PrintDefaults()
		fmt.Printf("\nFind a bug? Create an issue at: https://github.com/qnkhuat/tstream\n")
	}

	var db_path = flag.String("db", ".db", "Path to database")
	var host = flag.String("host", "localhost:3000", "Host address to serve server")
	var version = flag.Bool("version", false, fmt.Sprintf("TStream server version: %s", cfg.SERVER_VERSION))

	flag.Parse()

	fmt.Printf("TStream server v%s\n", cfg.SERVER_VERSION)

	if *version {
		fmt.Printf("TStream server %s\nGithub: https://github.com/qnkhuat/tstream\n", cfg.SERVER_VERSION)
		os.Exit(0)
		return
	}

	s, err := server.New(*host, *db_path)
	if err != nil {
		fmt.Printf("Failed to create server: %s", err)
		log.Printf("Failed to create server: %s", err)
		return
	}
	s.Start()
	return
}
