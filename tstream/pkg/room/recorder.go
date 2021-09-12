/***
Recorder service for room
It takes a block of messages and write to gzip files
A file often contains minutes of messages
***/
package room

import (
	"encoding/json"
	"fmt"
	"github.com/qnkhuat/tstream/pkg/message"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const MANIFEST_FILENAME = "manifest.json"

/* Manfiest
** Used to store information about all segments of a playback
** Playback player should use this file to find which files to download
**/

type Manifest struct {
	Id              int
	Version         int
	StartTime       time.Time
	StopTime        time.Time
	SegmentDuration int64
	Segments        []ManifestSegment
}

type ManifestSegment struct {
	// Offset time with the stream header
	Offset int64
	Id     int
	Path   string
}

/* Recodrer
**/
type Recorder struct {
	startTime        time.Time
	dir              string
	Interval         time.Duration
	lock             sync.Mutex
	currentSegmentID int
	currentSegment   *Segment
	manifest         *Manifest
}

func NewRecorder(dir string, id int) (*Recorder, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
	}
	//manifestPath := filepath.Join(dir, "manifest.jsonl")
	manifest := &Manifest{
		Version: 0,
		Id:      id,
	}
	return &Recorder{
		startTime:        time.Now(),
		currentSegmentID: 0,
		currentSegment:   NewSegment(),
		dir:              dir,
		manifest:         manifest,
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
	re.manifest.StartTime = time.Now()
	re.manifest.SegmentDuration = writeInterval.Milliseconds()

	for range time.Tick(writeInterval) {
		re.WriteCurrentSegment()
		re.dumpManifest()
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
	}

	manifestSegment := ManifestSegment{
		Offset: re.currentSegment.startTime.Sub(re.startTime).Milliseconds(),
		Id:     re.currentSegmentID,
		Path:   gzFileName,
	}

	re.manifest.Segments = append(re.manifest.Segments, manifestSegment)
	re.currentSegmentID += 1
	re.currentSegment = NewSegment()
	re.lock.Unlock()
	return err
}

func (re *Recorder) dumpManifest() error {
	path := filepath.Join(re.dir, MANIFEST_FILENAME)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	defer f.Close()
	if err != nil {
		return err
	}

	content, err := json.Marshal(re.manifest)
	if err != nil {
		return err
	}
	f.WriteString(string(content))
	return nil
}

func (re *Recorder) Stop() {
	re.WriteCurrentSegment()
	re.manifest.StopTime = time.Now()
	re.dumpManifest()
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
