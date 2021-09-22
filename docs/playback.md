# Resources
[player](https://github.com/asciinema/asciinema-player)
[server](https://github.com/asciinema/asciinema-server)
[html spec](https://html.spec.whatwg.org/#event-media-progress)


What we want:
- Record stdout in chunks (5-10 mins a file)
- When a viewer join a room they will download playback file in batch.
  - Why not stream it? 
    - If we stream then for each playback session we have to keep a process to maange what to send
    - And twitch + youtube work the same way, viewer just download a file and playback with it
  - A problem occur: when do we switch between record file?

- But asciinema use a different terminal renderer than what we are curerntly using. should we switch? xterm.js is a mature project
=> Nah we will re-implement the playback ourown. Take the asciinema-player as the baseline.
Why?
- The asciinema-player development are not active. the last change is 2 years ago.
- It's written in clojure, I love clojure but it's not the stack we are currently using for the front-end
- Help me better understand about how to render it

- Drawback: this format doesn't handle size changing. BUll shit

Action plan:
- Re-write tstream to compatible with [asciicast](https://github.com/asciinema/asciinema/blob/develop/doc/asciicast-v2.md) format
- Server save by interval with this format.
- Able to playback this format on asciinema.org
- Implement asciinema player in tsx using xterm.js as renderer
- Integrate it to our front-end

This is hard-core shit




After consideration Here is what I think:
What we want:
- We want this to be as fast as possible, if this is slower than traditional stremaing => it's worthless

That's said:
- Viewer will be at max 1.5 seconds lag with streamer
- Streaming will be serve using websocket
- Server will record those changes and save it every x seconds
- We will follow something like asciicast for format
- Playback will use http to download files
- The download part will be handled by the player
- Let's have 2 seperate versions for watching the stream and playback for now


How the playback is gonna work:
- The playback will feed data into the Write manager
- Write manager will accumulate into a queue
- Write manager keep an inner clock to where things went
- Write manager will have a scan rate => it'll scan and set time out for message during the next time span. this should be small ( ~ 200ms );
- Write manager will have a play/pause state





























