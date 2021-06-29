/*
Wrapper around the pty
Used to control (start, stop) and communicate with the terminal
*/

// Most the code are taken from : https://github.com/elisescu/tty-share/blob/master/pty_master.go
package ptyMaster

import (
	ptyDevice "github.com/creack/pty"
	//"golang.org/x/term"
	term "golang.org/x/crypto/ssh/terminal"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

type PtyMaster struct {
	cmd               *exec.Cmd
	f                 *os.File
	terminalInitState *term.State
}

type onWindowChangedCB func(*ptyDevice.Winsize)

// *** Getter/Setters ****
func (pty *PtyMaster) F() *os.File {
	return pty.f
}

func New() *PtyMaster {
	return &PtyMaster{}
}

func (pty *PtyMaster) Write(b []byte) (int, error) {
	return pty.f.Write(b)
}

func (pty *PtyMaster) Read(b []byte) (int, error) {
	return pty.f.Read(b)
}

func (pty *PtyMaster) StartShell() error {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "bash"
	}

	pty.cmd = exec.Command(shell)
	pty.cmd.Env = os.Environ()

	err := pty.StartCommand()
	if err != nil {
		return err
	}
	// Set the initial window size
	winSize, _ := GetWinsize(0)
	pty.SetWinsize(winSize)

	pty.SetWinChangeCB(func(ws *ptyDevice.Winsize) {
		pty.SetWinsize(ws)
	})
	return nil
}

func (pty *PtyMaster) StartCommand() error {
	f, err := ptyDevice.Start(pty.cmd)
	if err != nil {
		return err
	}
	pty.f = f
	return nil
}

func (pty *PtyMaster) Stop() error {
	signal.Ignore(syscall.SIGWINCH)

	err := pty.cmd.Process.Signal(syscall.SIGTERM)
	// TODO: Find a proper way to close the running command. Perhaps have a timeout after which,
	// if the command hasn't reacted to SIGTERM, then send a SIGKILL
	// (bash for example doesn't finish if only a SIGTERM has been sent)
	err = pty.cmd.Process.Signal(syscall.SIGKILL)

	err = pty.f.Close()
	return err
}

func (pty *PtyMaster) Restore() {
	term.Restore(0, pty.terminalInitState)
}

func (pty *PtyMaster) Refresh() {
	// We wanna force the app to re-draw itself, but there doesn't seem to be a way to do that
	// so we fake it by resizing the window quickly, making it smaller and then back big
	winSize, err := GetWinsize(0)
	if err != nil {
		return
	}

	winSize.Rows -= 1

	if err != nil {
		return
	}

	pty.SetWinsize(winSize)
	winSize.Rows += 1

	go func() {
		time.Sleep(time.Millisecond * 50)
		pty.SetWinsize(winSize)
	}()
}

func (pty *PtyMaster) Wait() error {
	return pty.cmd.Wait()
}

func (pty *PtyMaster) MakeRaw() error {
	// Save the initial state of the terminal, before making it RAW. Note that this terminal is the
	// terminal under which the tty-share command has been started, and it's identified via the
	// stdin file descriptor (0 in this case)
	// We need to make this terminal RAW so that when the command (passed here as a string, a shell
	// usually), is receiving all the input, including the special characters:
	// so no SIGINT for Ctrl-C, but the RAW character data, so no line discipline.
	// Read more here: https://www.linusakesson.net/programming/tty/
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	pty.terminalInitState = oldState
	return err
}

func (pty *PtyMaster) SetWinsize(ws *ptyDevice.Winsize) {
	ptyDevice.Setsize(pty.f, ws)
}

func onWindowChanges(wcCB onWindowChangedCB) {
	wcChan := make(chan os.Signal, 1)
	signal.Notify(wcChan, syscall.SIGWINCH)
	// The interface for getting window changes from the pty slave to its process, is via signals.
	// In our case here, the tty-share command (built in this project) is the client, which should
	// get notified if the terminal window in which it runs has changed. To get that, it needs to
	// register for SIGWINCH signal, which is used by the kernel to tell process that the window
	// has changed its dimentions.
	// Read more here: https://www.linusakesson.net/programming/tty/
	// Shortly, ioctl calls are used to communicate from the process to the pty slave device,
	// and signals are used for the communiation in the reverse direction: from the pty slave
	// device to the process.

	for {
		select {
		case <-wcChan:
			ws, err := GetWinsize(0)
			if err == nil {
				wcCB(ws)
			} else {
				log.Printf("Can't get window size: %s", err.Error())
			}
		}
	}
}

func (pty *PtyMaster) SetWinChangeCB(winChangedCB onWindowChangedCB) {
	// Start listening for window changes
	go onWindowChanges(func(ws *ptyDevice.Winsize) {
		pty.SetWinsize(ws)

		// Notify the PtyMaster user of the window changes, to be sent to the remote side
		winChangedCB(ws)
	})
}

func GetWinsize(fd int) (*ptyDevice.Winsize, error) {
	cols, rows, err := term.GetSize(fd)
	if err != nil {
		log.Printf("Failed to get winsize: %s", err)
		return nil, err
	}
	ws := &ptyDevice.Winsize{
		Rows: uint16(rows),
		Cols: uint16(cols),
		X:    uint16(0), // not used
		Y:    uint16(0), // not used
	}

	return ws, nil
}
