/***
Recorder service for room
It takes a block of messages and write to gzip files
A file often contains minutes of messages
***/
package room

import (
	"encoding/json"
	"fmt"
	"github.com/qnkhuat/tstream/internal/cfg"
	"github.com/qnkhuat/tstream/pkg/message"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// 20 min of asciiquarium generate 10mins of record
type Recorder struct {
	startTime        time.Time
	dir              string
	Interval         time.Duration
	lock             sync.Mutex
	currentSegmentID int
	currentSegment   *Segment
	manifest         *Manifest
}

func NewRecorder(dir string) (*Recorder, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
	}
	manifestPath := filepath.Join(dir, "manifest.jsonl")
	return &Recorder{
		startTime:        time.Now(),
		currentSegmentID: 0,
		currentSegment:   NewSegment(),
		dir:              dir,
		manifest:         NewManifest(manifestPath),
	}, nil
}

func (re *Recorder) AddMsg(msg message.Wrapper) error {
	// Delay compared to the start time of the block
	re.currentSegment.Add(msg)
	return nil
}

func (re *Recorder) newSegment() {
	re.currentSegment = NewSegment()
}

func (re *Recorder) Start(writeInterval time.Duration, id int) {
	manifestHeader := ManifestHeader{
		Id:              id,
		Version:         cfg.MANIFEST_VERSION,
		Time:            time.Now(),
		SegmentDuration: writeInterval.Milliseconds(),
	}
	err := re.manifest.WriteHeader(manifestHeader)
	if err != nil {
		log.Printf("Failed to write manifest header: %s", err)
		return
	}

	for range time.Tick(writeInterval) {
		re.WriteCurrentSegment()
	}
}

func (re *Recorder) WriteCurrentSegment() error {
	var err error
	re.lock.Lock()

	data, err := re.currentSegment.BytesBuffer()
	if err != nil {
		log.Printf("Failed to gzip data: %s", err)
		re.lock.Unlock()
		return err
	}

	gzFileName := fmt.Sprintf("%d.gz", re.currentSegmentID)
	gzPath := filepath.Join(re.dir, gzFileName)
	err = CreateWriteCloseGZ(gzPath, data)
	if err != nil {
		log.Printf("Failed to write data %s", err)
	} else {
		log.Printf("Writing: %s(%d)", gzPath, re.currentSegment.Nbuffer())
		manifestEntry := ManifestEntry{
			Offset: time.Since(re.startTime).Milliseconds(),
			Id:     re.currentSegmentID,
			Path:   gzFileName,
		}
		err = re.manifest.WriteEntry(manifestEntry)
		if err != nil {
			log.Printf("Failed to write entry: %s", err)
		}
	}

	re.currentSegmentID += 1
	re.currentSegment = NewSegment()
	re.lock.Unlock()
	return err

}

func (re *Recorder) Stop() {
	re.WriteCurrentSegment()
}

/* Segment
** Equivalent to one file that contains recorded content for a duration
**/

type Segment struct {
	startTime time.Time
	lock      sync.Mutex
	buffer    []message.Wrapper
}

func NewSegment() *Segment {
	var buffer []message.Wrapper
	se := &Segment{
		buffer:    buffer,
		startTime: time.Now(),
	}
	return se
}

func (se *Segment) Add(msg message.Wrapper) {
	msg.Delay = time.Since(se.startTime).Milliseconds()
	se.lock.Lock()
	se.buffer = append(se.buffer, msg)
	se.lock.Unlock()
}

func (se *Segment) Nbuffer() int {
	return len(se.buffer)
}

func (se *Segment) BytesBuffer() ([]byte, error) {
	var bytesBuffer []byte

	bytesBuffer, err := json.Marshal(se.buffer)
	if err != nil {
		return bytesBuffer, err
	}

	return bytesBuffer, nil
}

//func (se *Segment) Gzip() ([]byte, error) {
//	se.lock.Lock()
//	defer se.lock.Unlock()
//
//	dataByte, err := json.Marshal(se.buffer)
//	if err != nil {
//		return []byte{}, err
//	}
//
//	var b bytes.Buffer
//	gz := gzip.NewWriter(&b)
//	if _, err := gz.Write(dataByte); err != nil {
//		gz.Close()
//		return []byte{}, err
//	}
//	gz.Close()
//
//	return b.Bytes(), nil
//}

/* Manfiest
** Used to store information about all segments of a playback
** Playback player should use this file to find which files to download
**/

type Manifest struct {
	path string
}

type ManifestHeader struct {
	// Start time of stream
	Version         int       `json:"version"`
	Time            time.Time `json:"time"`
	Id              int       `json:"id"`
	SegmentDuration int64     `json:"segmentDuration"`
}

type ManifestEntry struct {
	// Offset time with the stream header
	Offset int64  `json:"offset"`
	Id     int    `json:"id"`
	Path   string `json:"path"`
}

func NewManifest(path string) *Manifest {
	return &Manifest{path}
}

func (ma *Manifest) WriteHeader(header ManifestHeader) error {
	if _, err := os.Stat(ma.path); os.IsExist(err) {
		// File shouldn't existed
		return err
	}

	f, err := os.Create(ma.path)
	defer f.Close()
	if err != nil {
		return err
	}
	lineByte, _ := json.Marshal(header)
	f.WriteString(string(lineByte) + "\n")
	return nil
}

func (ma *Manifest) WriteEntry(entry ManifestEntry) error {
	if _, err := os.Stat(ma.path); os.IsNotExist(err) {
		// File should existed and contains header
		return err
	}

	f, err := os.OpenFile(ma.path, os.O_APPEND|os.O_WRONLY, 0600)
	defer f.Close()
	if err != nil {
		return err
	}
	lineByte, _ := json.Marshal(entry)
	f.WriteString(string(lineByte) + "\n")
	return nil
}
