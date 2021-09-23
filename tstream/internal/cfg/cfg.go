package cfg

import (
	"time"
)

const (
	MANIFEST_VERSION                 = 0
	SERVER_VERSION                   = "1.3.2" // Used to verify compatible verion of streamer
	STREAMER_VERSION                 = "1.3.2" // retry connect with server if websocket is broke
	SERVER_STREAMER_REQUIRED_VERSION = "1.3.2" // Used to verify compatible verion of streamer

	// Room
	ROOM_BUFFER_SIZE      = 3                // number of recent broadcast message to buffer
	ROOM_CACHE_MSG_SIZE   = 25               // number of recent chat messages to buffer
	ROOM_SEGMENT_DURATION = 30 * time.Second // Duration of a segment for recording

	// Streamer
	STREAMER_ENVKEY_SESSIONID    = "TSTREAM_SESSIONID"
	STREAMER_READ_BUFFER_SIZE    = 1024             // streamer websocket read buffer size
	STREAMER_WRITE_BBUFFER_SIZE  = 1024             // streamer websocket write buffer size
	STREAMER_REFRESH_INTERVAL    = 30 * time.Second // Interval to refresh streamer pty
	STREAMER_RETRY_CONNECT_AFTER = 10 * time.Second // retry connect with server if websocket is broke

	// Server
	SERVER_READ_BUFFER_SIZE        = 1024                  // server websocket read buffer size
	SERVER_WRITE_BBUFFER_SIZE      = 1024                  // server websocket write buffer size
	SERVER_CLEAN_INTERVAL          = 60 * time.Second      // Scan for idle room interval
	SERVER_CLEAN_THRESHOLD         = 60 * 10 * time.Second // Threshold to be classified as idle room
	SERVER_PING_INTERVAL           = 10 * time.Second      // Interval to ping streamer to check status
	SERVER_DISCONNECTED_THRESHHOLD = 60 * time.Second      // Threshold of inactive time to classify streamer as disconnected
	SERVER_SYNCDB_INTERVAL         = 60 * time.Second      // Sync server state with DB interval
)
