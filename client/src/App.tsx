import React, { useEffect } from 'react';
import './App.css';
import { Terminal } from 'xterm';
//import { FitAddon } from 'xterm-addon-fit';
import { AttachAddon } from 'xterm-addon-attach';
import 'xterm/css/xterm.css';
import base64 from "./base64";

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
      console.log("Got event: ", ev);
      console.log(ev.data);
      ev.data.text().then((msg: string) => {
        console.log(msg.length);
        let msgObject = JSON.parse(msg)
        if (msgObject.Type === "Write") {
          var buffer = base64.base64ToArrayBuffer(msgObject.Data)
          term.writeUtf8(buffer);
        }
      })
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
