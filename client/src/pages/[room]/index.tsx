import React, { useEffect, useState } from 'react';
import { useParams } from "react-router-dom";

import * as base64 from "../../lib/base64";
import * as util from "../../lib/util";
import WSTerminal from "../../components/WSTerminal";

import StreamPreview from "../../components/StreamPreview";

import * as constants from "../../lib/constants";
import PubSub from "../../lib/pubsub";
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
  const [ msgManager, setMsgManager ] = useState<PubSub>();


  function resize() {
    setTermSize({
      Width: window.innerWidth - chatWinsize,
      Height: window.innerHeight,
    });
  }

  useEffect(() => {
    const wsUrl = util.getWsUrl(params.username);
    const ws = new WebSocket(wsUrl);

    const tempMsg = new PubSub();
    ws.onmessage = (ev: MessageEvent) => {
      let msg = JSON.parse(ev.data);

      if (msg.Type === constants.MSG_TWRITE) {

        var buffer = base64.toArrayBuffer(msg.Data)
        tempMsg.pub(msg.Type, buffer);

      } else if (msg.Type === constants.MSG_TWINSIZE) {

        let winSizeMsg = JSON.parse(window.atob(msg.Data));
        tempMsg.pub(msg.Type, winSizeMsg);

      }
    }

    tempMsg.sub(constants.MSG_TREQUEST_WINSIZE, () => {

      var payload_byte = base64.toArrayBuffer(window.btoa(""));
      var wrapper = JSON.stringify({
        Type: constants.MSG_TREQUEST_WINSIZE,
        Data: Array.from(payload_byte),
      });
      const payload = base64.toArrayBuffer(window.btoa(wrapper))
      util.sendWhenConnected(ws, payload);
    })

    setMsgManager(tempMsg);
    window.addEventListener("resize", () => resize());
    resize();
  }, [])

  return (
    <div id="room">
      <>
        <>
          {msgManager && termSize &&
          <WSTerminal
            className="bg-black"
            msgManager={msgManager}
            width={termSize?.Width ? termSize.Width : -1}
            height={termSize?.Height ? termSize.Height : -1}
          />}
        </>
      </>
    </div>
  );

}

export default Room;
