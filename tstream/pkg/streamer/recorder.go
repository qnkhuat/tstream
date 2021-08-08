/***
Recorder service for streamer.
It receives package from pty and manage when to send to server
Each message is a TermBlock. Inside termblock is multiple TermWrite message during a time interval
***/
package streamer

import (
	"bytes"
	"compress/gzip"
	//"encoding/base64"
	"encoding/json"
	"github.com/qnkhuat/tstream/pkg/message"
	"log"
	"sync"
	"time"
)

type Recorder struct {
	lock sync.Mutex

	// A queue to store message
	queue [][]byte

	// Channel to send message to
	out chan<- message.Wrapper

	// duration of each termwriteblock to send
	blockDuration time.Duration

	// delay of stream
	delay time.Duration

	currentBlock *Block
}

func NewRecorder(blockDuration time.Duration, delay time.Duration, out chan<- message.Wrapper) *Recorder {

	if delay < blockDuration {
		// delay should be larger than blockduraiton about .5 seconds for transmission time
		// this will ensure a smooth stream
		log.Printf("Block duration(%d) should smaller than delay(%d)", blockDuration, delay)
		blockDuration = delay
	}
	currentBlock := NewBlock(blockDuration, delay)
	return &Recorder{
		blockDuration: blockDuration,
		out:           out,
		currentBlock:  currentBlock,
		delay:         delay,
	}
}

func (r *Recorder) Start() {
	if r.out == nil {
		log.Printf("No output channel for recorder")
		return
	}

	// First message
	time.Sleep(r.delay)
	go r.Send()

	// Send all message in queue after each block duration
	for _ = range time.Tick(r.blockDuration) {
		go r.Send()
	}
}

// send all message in block and create a new one
func (r *Recorder) Send() {
	if r.currentBlock.NQueue() == 0 {
		r.newBlock()
		return
	}

	payload, err := r.currentBlock.Serialize()
	if err != nil {
		log.Printf("Failed to serialize block")
		r.newBlock()
		return
	}
	r.out <- payload
	r.newBlock()
}

func (r *Recorder) Write(data []byte) (int, error) {
	r.lock.Lock()
	r.currentBlock.AddMessage(data)
	r.lock.Unlock()
	return len(data), nil

}

func (r *Recorder) newBlock() {
	r.lock.Lock()
	r.currentBlock = NewBlock(r.blockDuration, r.delay)
	defer r.lock.Unlock()
}

/*** Block ***/
type Block struct {
	lock sync.Mutex

	// Each data block will have its own start time
	// Any message in queue will be offset to this startime
	startTime time.Time

	// how many milliseconds of data this block contains
	duration time.Duration

	delay time.Duration

	// queue of encoded termwrite message
	queue [][]byte
}

func NewBlock(duration time.Duration, delay time.Duration) *Block {
	var queue [][]byte
	return &Block{
		duration:  duration,
		queue:     queue,
		startTime: time.Now(),
		delay:     delay,
	}
}

func (bl *Block) Serialize() (message.Wrapper, error) {
	var msg message.Wrapper

	// Serialize message queue
	dataByte, err := json.Marshal(bl.queue)
	if err != nil {
		log.Printf("Failed to marshal message: %s", err)
		return msg, err
	}

	// compress with gzip
	// with gzip data often compressed to 1/10 -> 1/8 its original
	// Note: 3 seconds of parrot generate 70Kb of raw bytes. With gzip the data is just 6k
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write(dataByte); err != nil {
		log.Printf("Failed to gzip: %s", err)
		gz.Close()
		return msg, err
	}
	gz.Close()

	// construct return message
	blockMsg := message.TermWriteBlock{
		StartTime: bl.startTime,
		Duration:  bl.duration.Milliseconds(),
		Data:      b.Bytes(),
	}

	blockByte, err := json.Marshal(blockMsg)
	if err != nil {
		log.Printf("Failed to encode termwrite block message")
		return msg, err
	}

	msg = message.Wrapper{
		Type: message.TWriteBlock,
		Data: blockByte,
	}

	return msg, nil
}

func (bl *Block) AddMessage(data []byte) {
	bl.lock.Lock()

	// have to marshal any single termwrite message
	// or else the rendering will screw up
	byteData, _ := json.Marshal(message.TermWrite{
		Data: data,
		// offset of a single message is
		// the different between now and block start time
		// plus the (delay - duration)
		Offset: time.Since(bl.startTime).Milliseconds() + bl.delay.Milliseconds() - bl.duration.Milliseconds(),
	})
	bl.queue = append(bl.queue, byteData)

	bl.lock.Unlock()
}

func (bl *Block) NQueue() int {
	return len(bl.queue)
}
