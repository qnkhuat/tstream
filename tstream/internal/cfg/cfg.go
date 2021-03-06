package cfg

const (
	SERVER_VERSION                   = "1.3.3" // Version of tstream server
	STREAMER_VERSION                 = "1.3.3" // Version of tstream client
	SERVER_STREAMER_REQUIRED_VERSION = "1.3.2" // Streamer have to run this version or later to connect to server

	// Room
	ROOM_BUFFER_SIZE    = 3    // number of recent broadcast message to buffer
	ROOM_CACHE_MSG_SIZE = 25   // number of recent chat messages to buffer
	ROOM_DEFAULT_DELAY  = 3000 // act as both block size and delay time of streaming

	// Streamer
	STREAMER_READ_BUFFER_SIZE    = 1024 // streamer websocket read buffer size
	STREAMER_WRITE_BBUFFER_SIZE  = 1024 // streamer websocket write buffer size
	STREAMER_REFRESH_INTERVAL    = 30   // Interval to refresh streamer pty. Unit in seconds
	STREAMER_ENVKEY_SESSIONID    = "TSTREAM_SESSIONID"
	STREAMER_RETRY_CONNECT_AFTER = 10 // retry connect with server if websocket is broke

	// Server. All units are in seconds
	SERVER_READ_BUFFER_SIZE        = 1024    // server websocket read buffer size
	SERVER_WRITE_BBUFFER_SIZE      = 1024    // server websocket write buffer size
	SERVER_CLEAN_INTERVAL          = 60      // Scan for idle room interval. Unit in seconds
	SERVER_CLEAN_THRESHOLD         = 60 * 10 // Threshold to be classified as idle room. Unit in seconds
	SERVER_PING_INTERVAL           = 10      // Interval to ping streamer to check status
	SERVER_DISCONNECTED_THRESHHOLD = 60      // Threshold of inactive time to classify streamer as disconnected
	SERVER_SYNCDB_INTERVAL         = 60      // Sync server state with DB interval
)
