import React, { useEffect, useState } from 'react';
import { Terminal } from 'xterm';
import { FitAddon } from 'xterm-addon-fit';
import 'xterm/css/xterm.css';
import { useParams } from "react-router-dom";
import PubSub from "./../../lib/pubsub";

import WSTerminal from "../../components/WSTerminal";

function base64ToArrayBuffer(input:string) {

  var binary_string =  window.atob(input);
  var len = binary_string.length;
  var bytes = new Uint8Array( len );
  for (var i = 0; i < len; i++)        {
    bytes[i] = binary_string.charCodeAt(i);
  }
  return bytes;
}


interface Params {
  username: string;
}

function Room() {
  const params: Params = useParams();
  const [ inputValue, setInputValue ] = useState("");
  const sessionID = params.username;
  const wsUrl = `ws://0.0.0.0:3000/ws/${sessionID}/viewer`;
  const ws = new WebSocket(wsUrl);
  const msgManager = new PubSub();

  useEffect(() => {
    ws.onmessage = (ev: MessageEvent) => {
      console.log("got message; ", ev.data);
      let msg = JSON.parse(ev.data);

      if (msg.Type === "Write") {
        var buffer = base64ToArrayBuffer(msg.Data)
        msgManager.pub(msg.Type, buffer);
      } else if (msg.Type === "Winsize") {
        msgManager.pub(msg.Type, msg);
      }
    }


  }, [])

  return (
    <div id="room">
      <WSTerminal msgManager={msgManager} width={400} height={300}/>
      <WSTerminal msgManager={msgManager} width={400} height={300}/>
    </div>
  );

}

export default Room;
