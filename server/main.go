package main

import (
  "github.com/gorilla/websocket"
  "github.com/gorilla/mux"
  ptyDevice "github.com/creack/pty"
  "log"
  "golang.org/x/term"
  "flag"
  "net/http"
  "fmt"
  "time"
  "os"
	"os/exec"
  "os/signal"
  "syscall"
  "io"
  "encoding/json"
)

type MsgWrapper struct {
  Type string
  Data []byte
}


type PtyMaster struct {
  cmd *exec.Cmd
  f *os.File
	terminalInitState *term.State
}

func NewPtyMater () *PtyMaster {
  return &PtyMaster{}
}

func (pty *PtyMaster) StartShell() error{
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
	// TODO: Find a proper wai to close the running command. Perhaps have a timeout after which,
	// if the command hasn't reacted to SIGTERM, then send a SIGKILL
	// (bash for example doesn't finish if only a SIGTERM has been sent)
	pty.cmd.Process.Signal(syscall.SIGKILL)

  err := pty.f.Close()
  return err
}

func (pty *PtyMaster) Wait() error {
	return pty.cmd.Wait()
}

func InitLog(dest, prefix string) {
  f, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
  if err != nil {
    log.Fatalf("error opening file: %v", err)
  }
  log.SetOutput(f)
  log.SetFlags(log.LstdFlags | log.Lshortfile)
  log.SetPrefix(prefix)
}

// upgrade an http request to websocket
var httpUpgrader = websocket.Upgrader{
  ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
  log.Println("Connecting")
  httpUpgrader.CheckOrigin = func(r *http.Request) bool { return true }
  conn, err := httpUpgrader.Upgrade(w, r, nil)
  defer conn.Close()
  if err != nil {
    log.Panicf("Failed to upgrade to websocket: %s", err)
  }

  //go func() {
    for {
      buf := make([]byte, 1024)
      read, err := pty.f.Read(buf)
      log.Printf("Read byte: %x", buf)

      msg := MsgWrapper{
        Data: buf[:read],
        Type: "Write",
      }
      
      payload, err := json.Marshal(msg)
      if (err != nil) {
        log.Panicf("Failed to Encode msg: ", err)
      }
      if err != nil {
        conn.WriteMessage(websocket.TextMessage, []byte(err.Error()))
        log.Panicf("Unable to read from pty/cmd: %s", err)
        return
      }
      conn.WriteMessage(websocket.BinaryMessage, payload)
      log.Println("Sent a buffer")
    }
  //}()
}


func healthHandler(w http.ResponseWriter, r *http.Request) {
  log.Printf("health check")
  fmt.Fprintf(w, "I'm fine, go away: %s\n", time.Now().String())
}

var pty *PtyMaster
func main() {
  InitLog("log", "")
  var listen = flag.String("listen", "0.0.0.0:3000", "Host:port to listen on")
  log.Println("Start server")
  flag.Parse()

  pty = NewPtyMater()
  pty.StartShell()
  log.Println("Shell started")
  // Copy stdin to the pty and the pty to stdout.
  // NOTE: The goroutine will keep reading until the next keystroke before returning.
  go func() { io.Copy(pty.f, os.Stdin) }()
  go func() { io.Copy(os.Stdout, pty.f) }()

  router := mux.NewRouter()
  router.HandleFunc("/health", healthHandler) // Terminal session
  router.HandleFunc("/ws", handleWebSocket) // Terminal session

  httpServer := &http.Server{Addr: *listen, Handler:router}
  go func () {
    log.Printf("Http Server is serving at %s", *listen)
    if err := httpServer.ListenAndServe(); err != nil {
      log.Panicf("Something went wrong with the webserver: %s", err)
    }
  }()

  pty.Wait()
  pty.Stop()
  httpServer.Close()
  fmt.Fprintf(os.Stdout, "Bye!\n")
  log.Println("Stopped")
}


