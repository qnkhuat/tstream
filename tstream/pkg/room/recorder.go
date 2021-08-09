package room

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"github.com/qnkhuat/tstream/pkg/message"
	"log"
	"os"
	"sync"
	"time"
)

// 20 min of asciiquarium generate 10mins of playback
type Recorder struct {
	In             chan message.Wrapper
	Interval       time.Duration
	path           string
	currentblockID uint
	lock           sync.Mutex
	f              F
}

func NewRecorder(path string) *Recorder {
	f, _ := CreateGZ(path)
	return &Recorder{
		f:    f,
		path: path,
	}
}

func (re *Recorder) WriteMsg(msg message.Wrapper) error {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to encode message")
		return err
	}
	err = WriteGZ(re.f, data)
	if err != nil {
		log.Printf("Failed to write message: %s", err)
		return err
	}
	return nil
}

type F struct {
	f  *os.File
	gf *gzip.Writer
	bf *bufio.Writer
}

func CreateGZ(path string) (F, error) {
	var f F
	fi, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		log.Printf("Failed to create file: %s", err)
		return f, err
	}
	gf := gzip.NewWriter(fi)
	bf := bufio.NewWriter(gf)
	f = F{fi, gf, bf}
	return f, nil
}

func WriteGZ(f F, data []byte) error {
	n, err := f.bf.Write(data)
	if n != len(data) || err != nil {
		log.Printf("Failed to write data %s", err)
		return err
	}
	log.Printf("Wrote %d bytes", n)
	return nil
}

//func ReadGzFile(filename string) ([]byte, error) {
//	fi, err := os.Open(filename)
//	if err != nil {
//		return nil, err
//	}
//	defer fi.Close()
//
//	fz, err := gzip.NewReader(fi)
//	if err != nil {
//		return nil, err
//	}
//	defer fz.Close()
//
//	s, err := ioutil.ReadAll(fz)
//	if err != nil {
//		return nil, err
//	}
//	return s, nil
//}
