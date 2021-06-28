/*
Wrapper around the pty (https://dev.to/napicella/linux-terminals-tty-pty-and-shell-192e)
Used to control (start, stop) and communicate with the terminal
*/

// Most the code are taken from : https://github.com/elisescu/tty-share/blob/master/pty_master.go
package ptyMaster

import (
	ptyDevice "github.com/creack/pty"
	"golang.org/x/term"
	//"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	//"time"
)

type PtyMaster struct {
	cmd               *exec.Cmd
	f                 *os.File
	terminalInitState *term.State
}

type onWindowChangedCB func(int, int)

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

//func (pty *PtyMaster) Refresh() {
//	// We wanna force the app to re-draw itself, but there doesn't seem to be a way to do that
//	// so we fake it by resizing the window quickly, making it smaller and then back big
//	winSize, err := pty.f.GetsizeFull(0)
//	winSize.Rows -= 1
//
//	if err != nil {
//		return
//	}
//
//	pty.SetWinSize(winSize)
//	winSize.Rows += 1
//
//	go func() {
//		time.Sleep(time.Millisecond * 50)
//		pty.SetWinSize(winSize)
//	}()
//}

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

	//err = pty.MakeRaw()
	//if err != nil {
	//	return err
	//}

	//pty.Restore()

	// Set the initial window size
	//winSize, err := term.GetFullSize(0)
	//if err != nil {
	//	log.Printf("Failed to get wisize: %s", err)
	//}
	//pty.SetWinSize(winSize)
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

func (pty *PtyMaster) Restore() {
	term.Restore(0, pty.terminalInitState)
}

func (pty *PtyMaster) Wait() error {
	return pty.cmd.Wait()
}

//func (pty *PtyMaster) SetWinSize(rows, cols int) {
//	ptyDevice.Setsize(pty.ptyFile, rows, cols)
//}
//
//func onWindowChanges(wcCB onWindowChangedCB) {
//	wcChan := make(chan os.Signal, 1)
//	signal.Notify(wcChan, syscall.SIGWINCH)
//	// The interface for getting window changes from the pty slave to its process, is via signals.
//	// In our case here, the tty-share command (built in this project) is the client, which should
//	// get notified if the terminal window in which it runs has changed. To get that, it needs to
//	// register for SIGWINCH signal, which is used by the kernel to tell process that the window
//	// has changed its dimentions.
//	// Read more here: https://www.linusakesson.net/programming/tty/
//	// Shortly, ioctl calls are used to communicate from the process to the pty slave device,
//	// and signals are used for the communiation in the reverse direction: from the pty slave
//	// device to the process.
//
//	for {
//		select {
//		case <-wcChan:
//			cols, rows, err := term.GetSize(0)
//			if err == nil {
//				wcCB(cols, rows)
//			} else {
//				log.Warnf("Can't get window size: %s", err.Error())
//			}
//		}
//	}
//}
//
//func (pty *PtyMaster) SetWinChangeCB(winChangedCB onWindowChangedCB) {
//	// Start listening for window changes
//	go onWindowChanges(func(cols, rows int) {
//		// TODO:policy: should the server decide here if we care about the size and set it
//		// right here?
//		pty.SetWinSize(rows, cols)
//
//		// Notify the PtyMaster user of the window changes, to be sent to the remote side
//		winChangedCB(cols, rows)
//	})
//}
