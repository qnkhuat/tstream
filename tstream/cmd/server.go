/*
Roles of server:
- Receive the stdout of users terminal then broadcast to all users via websocket
*/

package main

import (
	"flag"
	"fmt"
	"github.com/qnkhuat/tstream/internal/cfg"
	"github.com/qnkhuat/tstream/internal/logging"
	"github.com/qnkhuat/tstream/pkg/server"
	"log"
	"os"
	"path/filepath"
)

func main() {
	logging.Config("/tmp/tstream.log", "SERVER: ")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Just type `server` to turn on server.\n\nAdvanced config:\n")
		flag.PrintDefaults()
		fmt.Printf("\nFind a bug? Create an issue at: https://github.com/qnkhuat/tstream\n")
	}

	var dbPath = flag.String("db", ".db", "Path to database")
	var host = flag.String("host", "localhost:3000", "Host address to serve server")
	var version = flag.Bool("version", false, fmt.Sprintf("TStream server version: %s", cfg.SERVER_VERSION))
	var playbackDir = flag.String("playback", ".tstream/", "Directory to save playback files")

	flag.Parse()

	fmt.Printf("TStream server v%s\n", cfg.SERVER_VERSION)

	if *version {
		fmt.Printf("TStream server %s\nGithub: https://github.com/qnkhuat/tstream\n", cfg.SERVER_VERSION)
		os.Exit(0)
		return
	}

	log.Printf("is abs: %s", filepath.IsAbs(*playbackDir))
	absPlaybackDir, err := filepath.Abs(*playbackDir)
	if err != nil {
		log.Printf("Failed to create playback dir: %s", err)
		return
	}
	fmt.Printf("Saving playback at: %s\n", absPlaybackDir)
	log.Printf("Saving playback at: %s", absPlaybackDir)
	s, err := server.New(*host, *dbPath, absPlaybackDir)
	if err != nil {
		fmt.Printf("Failed to create server: %s", err)
		log.Printf("Failed to create server: %s", err)
		return
	}
	s.Start()
	return
}
