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
	"fmt"
	"github.com/manifoldco/promptui"
	"github.com/qnkhuat/tstream/internal/logging"
	"github.com/qnkhuat/tstream/pkg/streamer"
	"log"
	"os/user"
	"regexp"
	"strings"
)

func main() {
	logging.Config("/tmp/tstream.log", "STREAMER: ")

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

	validateServer := func(input string) error {
		if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
			return nil
		} else {
			return fmt.Errorf("Invalide server url")
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

	promptServer := promptui.Prompt{
		Label:    "TStream Server",
		Default:  "https://server.tstream.club",
		Validate: validateServer,
	}

	username, err = promptUsername.Run()
	title, err := promptTitle.Run()
	server, err := promptServer.Run()

	s := streamer.New("https://tstream.club", server, username, title)
	err = s.Start()
	if err != nil {
		log.Panicf("Failed to start stream: %s", err)
	}
}
