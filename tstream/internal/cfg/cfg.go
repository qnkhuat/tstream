package cfg

const (
	// Room
	ROOM_BUFFER_SIZE = 20 // number of recent broadcast message to buffer

	// streamer
	STREAMER_READ_BUFFER_SIZE   = 1024 // streamer websocket read buffer size
	STREAMER_WRITE_BBUFFER_SIZE = 1024 // streamer websocket write buffer size
	STREAMER_REFRESH_INTERVAL   = 60   // Interval to refresh streamer pty. Unit in seconds

	// Server
	SERVER_CLEAN_INTERVAL     = 60      // Scan for idle room interval. Unit in seconds
	SERVER_CLEAN_THRESHOLD    = 60 * 10 // Threshold to be classified as idle room. Unit in seconds
	SERVER_READ_BUFFER_SIZE   = 1024    // server websocket read buffer size
	SERVER_WRITE_BBUFFER_SIZE = 1024    // server websocket write buffer size

)
