import React, { useEffect, useState } from 'react';
import { Terminal } from 'xterm';
import { FitAddon } from 'xterm-addon-fit';
import 'xterm/css/xterm.css';
import { useParams } from "react-router-dom";
import PubSub from "./../../lib/pubsub";
import * as base64 from "./../../lib/base64";
import WSTerminal from "../../components/WSTerminal";



interface Params {
  username: string;
}

function Room() {
  const params: Params = useParams();
  const sessionID = params.username;

  const wsUrl = `ws://0.0.0.0:3000/ws/${sessionID}/viewer`;
  const ws = new WebSocket(wsUrl);

  const [ inputValue, setInputValue ] = useState("");
  const msgManager = new PubSub();

  useEffect(() => {
    ws.onmessage = (ev: MessageEvent) => {
      let msg = JSON.parse(ev.data);
      if (msg.Type === "Write") {
        var buffer = base64.toArrayBuffer(msg.Data)
        msgManager.pub(msg.Type, buffer);
      } else if (msg.Type === "Winsize") {
        let winSizeMsg = JSON.parse(window.atob(msg.Data));
        msgManager.pub(msg.Type, winSizeMsg);
      }
    }
  }, [])

  return (
    <div id="room">
      <WSTerminal msgManager={msgManager} width={1000} height={300} />
    </div>
  );

}

export default Room;
