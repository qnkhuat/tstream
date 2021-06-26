package ptyMaster

import (
	ptyDevice "github.com/creack/pty"
	"golang.org/x/term"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

type PtyMaster struct {
	cmd               *exec.Cmd
	f                 *os.File
	terminalInitState *term.State
}

// *** Getter/Setters ****
func (pty *PtyMaster) F() *os.File {
	return pty.f
}

func New() *PtyMaster {
	return &PtyMaster{}
}

func (pty *PtyMaster) StartShell() error {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "bash"
	}

	pty.cmd = exec.Command(shell)
	pty.cmd.Env = os.Environ()
	return pty.StartCommand()
}

func (pty *PtyMaster) StartCommand() error {
	f, err := ptyDevice.Start(pty.cmd)
	if err != nil {
		return err
	}
	pty.f = f

	// Save the initial state of the terminal, before making it RAW. Note that this terminal is the
	// terminal under which the tty-share command has been started, and it's identified via the
	// stdin file descriptor (0 in this case)
	// We need to make this terminal RAW so that when the command (passed here as a string, a shell
	// usually), is receiving all the input, including the special characters:
	// so no SIGINT for Ctrl-C, but the RAW character data, so no line discipline.
	// Read more here: https://www.linusakesson.net/programming/tty/
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	return nil
}

func (pty *PtyMaster) Stop() error {
	signal.Ignore(syscall.SIGWINCH)

	pty.cmd.Process.Signal(syscall.SIGTERM)
	// TODO: Find a proper way to close the running command. Perhaps have a timeout after which,
	// if the command hasn't reacted to SIGTERM, then send a SIGKILL
	// (bash for example doesn't finish if only a SIGTERM has been sent)
	pty.cmd.Process.Signal(syscall.SIGKILL)

	err := pty.f.Close()
	return err
}

func (pty *PtyMaster) Wait() error {
	return pty.cmd.Wait()
}
