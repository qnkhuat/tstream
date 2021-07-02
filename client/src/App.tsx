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

function Chat(props: any) { 
  function chatSection(messengeData: any) {
    var data = messengeData as any;
    let jsxMessenges: Array<any> = [];
    console.log(data);
    for (var i = data.length - 1; i >= 0; i--) {
        jsxMessenges.push(
          <div className="chat" key={i}>
            <p className="username">{data[i].name}</p>
            <p className="chat-content">{data[i].content}</p>
          </div>
        );
    }
    return <div id="chat-section">{jsxMessenges}</div>
  }

  return (
    <>
      {chatSection(props.messengeData)}
    </>
  );
}

function App() {

  const [ messengeData, setMessengeData ] = useState([]);
  const [ inputUsername, setInputUsername] = useState("");
  const [ inputValue, setInputValue ] = useState("");
  const url = window.location.href;
  const url_splits = url.split("/")
  const sessionID = url_splits[url_splits.length - 1]
  const ws = new WebSocket(`ws://0.0.0.0:3000/${sessionID}/wsv`);
  const [ websocket, setWebSocket ] = useState(ws);
  console.log(ws);


  useEffect(() => {

    var term = new Terminal({
      cursorBlink: true,
      macOptionIsMeta: true,
      scrollback: 1000,
      fontSize: 12,
      letterSpacing: 0,
      fontFamily: 'SauceCodePro MonoWindows, courier-new, monospace',
    });

    ws.onopen = (e: Event) => {
      console.log("Socket open");
    }

    ws.onmessage = (ev: MessageEvent) => {
      console.log("Got message", ev.data);
      let msg = JSON.parse(ev.data);
      if (msg.Type === "Chat") {
        let messenge = JSON.parse(window.atob(msg.Data))
        var tempMessengeData = messengeData as any;
        tempMessengeData.push({
          "name": messenge.name,
          "content": messenge.content,
        });
        setMessengeData(tempMessengeData);
      }
      else if (msg.Type === "Write") {
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
    <> 
    <div className="App">
      <h1>Yooooooooooooooo</h1>
      <div id="terminal"></div>
      <input onChange={e => setInputUsername(e.target.value)}></input>
      <input onChange={e => setInputValue(e.target.value)}></input>
      <button onClick={e => {
        console.log(inputValue);
        var payload = JSON.stringify({
          "name": inputUsername,
          "content": inputValue,
        }); 
        websocket.send(payload);
      }}>Send message</button>
      <button onClick={e => {
        websocket.close();
      }}>Close connection</button>
    </div>
     <Chat messengeData={messengeData}/>
    </>
   );
}

export default App;
