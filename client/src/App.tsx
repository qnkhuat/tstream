import React, { useEffect } from 'react';
import './App.css';
import { Terminal } from 'xterm';
//import { FitAddon } from 'xterm-addon-fit';
import { AttachAddon } from 'xterm-addon-attach';
import 'xterm/css/xterm.css';
import base64 from './base64';

function base64ToArrayBuffer(input:string) {

  var binary_string =  window.atob(input);
  var len = binary_string.length;
  var bytes = new Uint8Array( len );
  for (var i = 0; i < len; i++)        {
    bytes[i] = binary_string.charCodeAt(i);
  }
  return bytes;
}


////const websocket = new WebSocket('ws://localhost:8002/s/local/ws');
const websocket = new WebSocket('ws://0.0.0.0:3000/r/qnkhuat/wsv');

function App() {

  useEffect(() => {
    console.log(websocket)
    var term = new Terminal({
      cursorBlink: true,
      macOptionIsMeta: true,
      scrollback: 1000,
      fontSize: 12,
      letterSpacing: 0,
      fontFamily: 'SauceCodePro MonoWindows, courier-new, monospace',
    });

    websocket.onopen = (e: Event) => {
      console.log("Socket open");
    }

    websocket.onmessage = (ev: MessageEvent) => {
      console.log("GOt message", ev.data);
      let msg = JSON.parse(ev.data);
      if (msg.Type === "Write") {
        var buffer = base64ToArrayBuffer(msg.Data)
        term.writeUtf8(buffer);
      }
      else if (msg.Type === "Winsize") {
        let winSizeMsg = JSON.parse(window.atob(msg.Data))
        term.resize(winSizeMsg.Cols, winSizeMsg.Rows)
      }
    }

    const terminalDiv = document.getElementById("terminal");
    if (terminalDiv) {
      term.open(terminalDiv);
    }

  })

  return (
    <div className="App">
      <h1>Yooooooooooooooo</h1>
      <div id="terminal"></div>
    </div>
  );
}

export default App;
