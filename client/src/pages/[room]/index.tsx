import React, { useEffect, useState } from 'react';
import { Terminal } from 'xterm';
import { FitAddon } from 'xterm-addon-fit';
import 'xterm/css/xterm.css';
import { useParams } from "react-router-dom";

import * as base64 from "./../../lib/base64";
import * as util from "./../../lib/util";
import PubSub from "./../../lib/pubsub";
import WSTerminal from "../../components/WSTerminal";

interface Params {
  username: string;
}

interface Winsize {
  Width: number;
  Height: number;

}

function Room() {
  const params: Params = useParams();
  const chatWinsize = 400; // px

  const [ inputValue, setInputValue ] = useState("");
  const [ termSize, setTermSize ] = useState<Winsize>();
  const msgManager = new PubSub();

  function resize() {
    setTermSize({
      Width: window.innerWidth - chatWinsize,
      Height: window.innerHeight,
    });
    console.log("resizing: ", {
      Width: window.innerWidth - chatWinsize,
      Height: window.innerHeight,
    });
  }

  useEffect(() => {
    const wsUrl = util.getWsUrl(params.username);
    const ws = new WebSocket(wsUrl);

    ws.onmessage = (ev: MessageEvent) => {
      let msg = JSON.parse(ev.data);

      if (msg.Type === "Write") {
        var buffer = base64.toArrayBuffer(msg.Data)
        msgManager.pub("Write", buffer);
      } else if (msg.Type === "Winsize") {
        let winSizeMsg = JSON.parse(window.atob(msg.Data));
        msgManager.pub(msg.Type, winSizeMsg);
      } else if (msg.Type == "Client") {
        let payload = JSON.parse(window.atob(msg.Data));
        console.log("payload is: ", payload);
      }
    }
    
    setTimeout(() => {
      ws.send(JSON.stringify({
        type: "chat",
        name: "manhcd",
        content: "Yoooooo",
        time: new Date().toISOString(),
      }));
      console.log("ws.send successfully");
    }, 2000);

    window.addEventListener("resize", () => resize());
    resize();

  }, [])


  return (
    <div id="room">
      {termSize &&
      <WSTerminal className="bg-red-400" msgManager={msgManager}
        width={termSize?.Width ? termSize.Width : -1} height={termSize?.Height ? termSize.Height : -1}/>}
    </div>
  );

}

export default Room;
