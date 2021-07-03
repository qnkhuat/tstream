import React, { useEffect, useState } from 'react';
import { Terminal } from 'xterm';
import { FitAddon } from 'xterm-addon-fit';
import 'xterm/css/xterm.css';
import { useParams } from "react-router-dom";

import * as base64 from "./../../lib/base64";
import * as util from "./../../lib/util";
import PubSub from "./../../lib/pubsub";
import WSTerminal from "../../components/WSTerminal";

import StreamPreview from "../../components/StreamPreview";

import * as constants from "../../constants";
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
  const msgManager = new PubSub(true);

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
    console.log("Called")
    const wsUrl = util.getWsUrl(params.username);
    const ws = new WebSocket(wsUrl);

    ws.onmessage = (ev: MessageEvent) => {
      let msg = JSON.parse(ev.data);
      console.log("Got message: ", msg.Type);

      if (msg.Type === constants.MSG_TWRITE) {
        var buffer = base64.toArrayBuffer(msg.Data)
        msgManager.pub(msg.Type, buffer);
      } else if (msg.Type === constants.MSG_TWINSIZE) {
        let winSizeMsg = JSON.parse(window.atob(msg.Data));
        msgManager.pub(msg.Type, winSizeMsg);
      } else {
        console.error("Unknown msg type");
      }
    }

    window.addEventListener("resize", () => resize());
    resize();
  }, [])


  return (
    <div id="room">
      {termSize &&
      <WSTerminal msgManager={msgManager}
        width={termSize?.Width ? termSize.Width : -1} height={termSize?.Height ? termSize.Height : -1}
      />}
    </div>
  );

}

export default Room;
