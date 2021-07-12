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
	"bufio"
	"flag"
	"fmt"
	"github.com/manifoldco/promptui"
	"github.com/qnkhuat/tstream/internal/cfg"
	"github.com/qnkhuat/tstream/internal/logging"
	"github.com/qnkhuat/tstream/pkg/streamer"
	"log"
	"os"
	"os/user"
	"regexp"
)

func main() {
	// Check if current process is under a tstream session
	if len(os.Getenv(cfg.STREAMER_ENVKEY_SESSIONID)) > 0 {
		fmt.Printf("This terminal is currently running under session: %s\nType 'exit' to stop the current session!\n", os.Getenv(cfg.STREAMER_ENVKEY_SESSIONID))
		os.Exit(1)
	}

	logging.Config("/tmp/tstream.log", "STREAMER: ")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "To Stream: just type in `tstream`.\n\nAdvanced config:\n")
		flag.PrintDefaults()
	}

	var server = flag.String("server", "https://server.tstream.club", "Server endpoint")
	var client = flag.String("client", "https://tstream.club", "TStream client url")

	flag.Parse()
	user, err := user.Current()
	if err != nil {
		log.Fatalf(err.Error())
	}

	username := user.Username
	validateUsername := func(input string) error {
		var validUsername = regexp.MustCompile(`^[a-z][a-z0-9]*[._-]?[a-z0-9]+$`)
		if validUsername.MatchString(input) && len(input) > 3 && len(input) < 20 {
			return nil
		} else {
			return fmt.Errorf("Invalid username")
		}
	}

	validateTitle := func(input string) error {
		if len(input) > 1 {
			return nil
		} else {
			return fmt.Errorf("Title must not be empty")
		}
	}

	promptUsername := promptui.Prompt{
		Label:    "Username",
		Default:  username,
		Validate: validateUsername,
	}

	promptTitle := promptui.Prompt{
		Label:    "Stream title",
		Validate: validateTitle,
	}

	username, err = promptUsername.Run()
	if err != nil {
		os.Exit(1)
	}
	title, err := promptTitle.Run()
	if err != nil {
		os.Exit(1)
	}

	s := streamer.New(*client, *server, username, title)

	statusCode := s.RequestAddRoom()
	log.Printf("Got status code: %d", statusCode)
	if statusCode == 400 {
		fmt.Printf("Detected a session is streaming with the same username\nProceed to stream from this terminal? (y/n): ")
		confirm, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		if confirm[0] != 'y' {
			os.Exit(1)
		}
	} else if statusCode == 401 {
		fmt.Printf("Username: %s is currently used by other streamer. Please use a different username!\n", username)
		os.Exit(1)
	}
	err = s.Start()
	if err != nil {
		log.Printf("Failed to start tstream : %s", err)
	}
}
