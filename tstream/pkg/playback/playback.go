package playback

/***
File size note:
7 Min parrot: 8.8 Mbs

***/
import (
	"encoding/json"
	"github.com/qnkhuat/tstream/pkg/message"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

type Playback struct {
	id    uint64
	root  string
	queue []message.TermWrite
	In    chan message.Wrapper
}

func New(id uint64, root string) *Playback {
	in := make(chan message.Wrapper, 256)
	return &Playback{
		id:   id,
		root: root,
		In:   in,
	}
}

func (p *Playback) Start() {
	// Open a json writer file
	// fileroot is root/id.json
	path := filepath.Join(p.root, strconv.FormatUint(p.id, 10)+".json")

	f, err := os.Create(path)
	if err != nil {
		log.Printf("Failed to create playback file")
		return
	}

	log.Printf("Playback are saving to %s", path)

	defer f.Close()
	// Start the playback
	for {
		msg, ok := <-p.In
		if !ok {
			log.Printf("Failed to read from channel")
			continue
		}

		// Write the message to the file
		if err := json.NewEncoder(f).Encode(msg); err != nil {
			log.Printf("Failed ot save message: %s", err)
		}
	}
}
