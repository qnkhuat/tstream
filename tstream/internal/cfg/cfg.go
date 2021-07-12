package cfg

const (
	// Room
	ROOM_BUFFER_SIZE = 20 // number of recent broadcast message to buffer

	// Streamer
	STREAMER_READ_BUFFER_SIZE    = 1024 // streamer websocket read buffer size
	STREAMER_WRITE_BBUFFER_SIZE  = 1024 // streamer websocket write buffer size
	STREAMER_REFRESH_INTERVAL    = 30   // Interval to refresh streamer pty. Unit in seconds
	STREAMER_ENVKEY_SESSIONID    = "TSTREAM_SESSIONID"
	STREAMER_RETRY_CONNECT_AFTER = 10      // retry connect with server if websocket is broke
	STREAMER_VERSION             = "1.0.0" // retry connect with server if websocket is broke

	// Server. All units are in seconds
	SERVER_READ_BUFFER_SIZE        = 1024    // server websocket read buffer size
	SERVER_WRITE_BBUFFER_SIZE      = 1024    // server websocket write buffer size
	SERVER_CLEAN_INTERVAL          = 60      // Scan for idle room interval. Unit in seconds
	SERVER_CLEAN_THRESHOLD         = 60 * 10 // Threshold to be classified as idle room. Unit in seconds
	SERVER_PING_INTERVAL           = 10      // Interval to ping streamer to check status
	SERVER_DISCONNECTED_THRESHHOLD = 60      // Threshold of inactive time to classify streamer as disconnected
	SERVER_SYNCDB_INTERVAL         = 60      // Sync server state with DB interval
	SERVER_STREAMER_VERSION        = "1.0.0" // Used to verify compatible verion of streamer
)
