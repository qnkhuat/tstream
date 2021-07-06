# tstream
Stream your terminal

Think Twitch, but only work on terminal.

The upside is it's very simple to start your stream. One command line and you good to go


# How to run
Open 3 terminals
1. `go run tstream/cmd/server/main.go` to start server
2. `go run tstream/cmd/streamer/main.go` to start a streaming session => This shell will be streamed
3. `cd client && npm install && npm run start` then go to localhost:3001. Your terminal will be streamed here

# RoadMap
- [x] One command to stream terminal session to web => just like tty-share
- [ ] Add Chat feature
- [ ] Add voice
- [ ] Browsing and admin system

