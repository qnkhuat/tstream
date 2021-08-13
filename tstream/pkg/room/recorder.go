package room

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/qnkhuat/tstream/pkg/message"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// 20 min of asciiquarium generate 10mins of playback
type Recorder struct {
	Interval       time.Duration
	dir            string
	currentblockID uint
	lock           sync.Mutex
}

func NewRecorder(dir string) (*Recorder, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
	}
	return &Recorder{
		dir:            dir,
		currentblockID: 0,
	}, nil
}

func (re *Recorder) WriteMsg(msg message.Wrapper) error {
	path := filepath.Join(re.dir, fmt.Sprintf("%d.gz", re.currentblockID))

	f, err := CreateGZ(path)
	if err != nil {
		log.Printf("Failed to create file: %s", err)
		return err
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to encode message")
		return err
	}

	err = WriteGZ(f, data)
	if err != nil {
		log.Printf("Failed to write message: %s", err)
		return err
	}

	CloseGZ(f)
	re.currentblockID += 1
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
	return nil
}

func CloseGZ(f F) {
	f.bf.Flush()
	// Close the gzip first.
	f.gf.Close()
	f.f.Close()
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
