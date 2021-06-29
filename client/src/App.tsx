import React, { useEffect, useState } from 'react';
import './App.css';
import { Terminal } from 'xterm';
import { FitAddon } from 'xterm-addon-fit';
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

  const [inputValue, setInputValue] = useState("");
  useEffect(() => {
    console.log("yuo");
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

    const wrapperDiv = document.getElementById("terminal");
    if (wrapperDiv != null) {
      wrapperDiv.innerHTML = "";
      const termDiv = document.createElement("div");
      wrapperDiv.appendChild(termDiv)
      term.open(termDiv);
    }

  }, [])

  return (
    <div className="App">
      <h1>Yooooooooooooooo</h1>
      <div id="terminal"></div>
      <input id="message" onChange={e => setInputValue(e.target.value)}></input>
      <button onClick={e => {
        console.log(inputValue);
        websocket.send(inputValue);
      }}>Send message</button>
      <button onClick={e => {
        websocket.close();
      }}>Close connection</button>

    </div>
  );
}

export default App;
