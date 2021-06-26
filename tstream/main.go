//package main
//
//import (
//	"encoding/json"
//	"flag"
//	"fmt"
//	ptyDevice "github.com/creack/pty"
//	"github.com/gorilla/mux"
//	"github.com/gorilla/websocket"
//	"io"
//	"log"
//	"net/http"
//	"os"
//	"os/exec"
//	"os/signal"
//	"syscall"
//	"time"
//)
//
//type MsgWrapper struct {
//	Type string
//	Data []byte
//}
//
//func InitLog(dest, prefix string) {
//	f, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
//	if err != nil {
//		log.Fatalf("error opening file: %v", err)
//	}
//	log.SetOutput(f)
//	log.SetFlags(log.LstdFlags | log.Lshortfile)
//	log.SetPrefix(prefix)
//}
//
//// upgrade an http request to websocket
//var httpUpgrader = websocket.Upgrader{
//	ReadBufferSize:  1024,
//	WriteBufferSize: 1024,
//}
//
//func handleWebSocket(w http.ResponseWriter, r *http.Request) {
//	log.Println("Connecting")
//	httpUpgrader.CheckOrigin = func(r *http.Request) bool { return true }
//	conn, err := httpUpgrader.Upgrade(w, r, nil)
//	defer conn.Close()
//	if err != nil {
//		log.Panicf("Failed to upgrade to websocket: %s", err)
//	}
//
//	// TODO: turn this into a struct that implement READ/WRITE
//	for { // why this function run with a low frequency?
//		buf := make([]byte, 1024)
//		read, err := pty.f.Read(buf)
//
//		msg := MsgWrapper{
//			Data: buf[:read],
//			Type: "Write",
//		}
//
//		payload, err := json.Marshal(msg)
//		if err != nil {
//			log.Panicf("Failed to Encode msg: ", err)
//		}
//		if err != nil {
//			conn.WriteMessage(websocket.TextMessage, []byte(err.Error()))
//			log.Panicf("Unable to read from pty/cmd: %s", err)
//			return
//		}
//		conn.WriteMessage(websocket.BinaryMessage, payload)
//		log.Println("Sent a buffer")
//	}
//}
//
//func healthHandler(w http.ResponseWriter, r *http.Request) {
//	log.Printf("health check")
//	fmt.Fprintf(w, "I'm fine, go away: %s\n", time.Now().String())
//}
//
//var pty *PtyMaster
//
//func main() {
//	InitLog("log", "")
//	var listen = flag.String("listen", "0.0.0.0:3000", "Host:port to listen on")
//	log.Println("Start server")
//	flag.Parse()
//
//	pty = NewPtyMater()
//	pty.StartShell()
//	log.Println("Shell started")
//	// Copy stdin to the pty and the pty to stdout.
//	// NOTE: The goroutine will keep reading until the next keystroke before returning.
//	go func() { io.Copy(pty.f, os.Stdin) }()
//	go func() { io.Copy(os.Stdout, pty.f) }()
//
//	router := mux.NewRouter()
//	router.HandleFunc("/health", healthHandler) // Terminal session
//	router.HandleFunc("/ws", handleWebSocket)   // Terminal session
//
//	httpServer := &http.Server{Addr: *listen, Handler: router}
//	go func() {
//		log.Printf("Http Server is serving at %s", *listen)
//		if err := httpServer.ListenAndServe(); err != nil {
//			log.Panicf("Something went wrong with the webserver: %s", err)
//		}
//	}()
//
//	pty.Wait()
//	pty.Stop()
//	httpServer.Close()
//	fmt.Fprintf(os.Stdout, "Bye!\n")
//	log.Println("Stopped")
//}
