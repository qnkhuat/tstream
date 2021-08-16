// Inspiration from https://github.com/asciinema/asciinema/blob/develop/doc/asciicast-v2.md
// With modification:
// Add message type: s : Window size
package message

type Header struct {
	version   int
	width     uint
	height    uint
	timestamp uint
	title     string
}

type AsciiCastEventType string

const (
	// read from stdin
	EIn AsciiCastEventType = "i"
	// write to stdout
	EOut AsciiCastEventType = "o"
	// winsize change
	ESize AsciiCastEventType = "s"
)

type Record struct {
	time      uint
	eventType AsciiCastEventType
	data      string
}
