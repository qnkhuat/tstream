package streamer

import (
	//"compress/gzip"
	"encoding/json"
	//"fmt"
	"github.com/qnkhuat/tstream/pkg/message"
	"log"
	//"os"
	"sync"
	"time"
)

type Transporter struct {
	out            chan<- message.Wrapper
	queue          [][]byte
	lock           sync.Mutex
	blockID        uint
	blockStartTime time.Time
}

func NewTransporter(out chan<- message.Wrapper) *Transporter {
	var queue [][]byte
	return &Transporter{
		out:     out,
		queue:   queue,
		blockID: 0,
	}
}

func (t *Transporter) Start() {
	for _ = range time.Tick(3 * time.Second) {
		//if len(t.queue) == 1 {
		log.Printf("sending, queue len: %d", len(t.queue))

		//f, _ := os.Create(fmt.Sprintf("gzip_%d.gz", t.blockID))
		//defer f.Close()
		//gz := gzip.NewWriter(f)
		//defer gz.Close()

		//fraw, _ := os.Create(fmt.Sprintf("raw_%d", t.blockID))
		//defer fraw.Close()

		//if _, err := gz.Write(dataArray); err != nil {
		//	log.Fatal(err)
		//}
		//fraw.Write(dataArray)

		dataByte, _ := json.Marshal(t.queue)
		payload := message.Wrapper{
			Type: message.TWrite,
			Data: dataByte,
		}

		t.out <- payload
		t.newBlock()

		//}
	}
}

/***
  Note: 5 seconds of parrot generate 62Kb of raw bytes. With gzip the data is just 977B
***/
func (t *Transporter) Write(data []byte) (int, error) {
	// log.Printf("Transporter Receiving message: %d", len(data))

	// log.Printf("data: %s, %v", string(data), time.Since(t.blockStartTime).Milliseconds())

	t.lock.Lock()
	byteData, _ := json.Marshal(message.TermWrite{Data: data, Time: time.Since(t.blockStartTime).Milliseconds()})
	t.queue = append(t.queue, byteData)
	t.lock.Unlock()
	//if time.Since(t.blockStartTime) > 3*time.Second {

	return len(data), nil
}

func (t *Transporter) newBlock() {
	t.blockID += 1
	t.queue = t.queue[:0] // clear it after backup
	t.blockStartTime = time.Now()
}
