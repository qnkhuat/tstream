import React, { useEffect, useState, useRef } from 'react';
import { useParams, Link } from "react-router-dom";

import * as base64 from "../../lib/base64";
import * as util from "../../lib/util";
import WSTerminal from "../../components/WSTerminal";
import Navbar from "../../components/Navbar";

import * as constants from "../../lib/constants";
import PubSub from "../../lib/pubsub";

import { createTheme, ThemeProvider } from '@material-ui/core/styles';
import CssBaseline from '@material-ui/core/CssBaseline';

const darkTheme = createTheme({
  palette: {
    mode: "dark",
  },
});


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

  const [ termSize, setTermSize ] = useState<Winsize>();
  const [ msgManager, setMsgManager ] = useState<PubSub>();
  const navbarRef = useRef<HTMLDivElement>(null);

  function resize() {
    if (navbarRef.current) {
      setTermSize({
        Width: window.innerWidth - chatWinsize,
        Height: window.innerHeight - navbarRef.current.offsetHeight,
      });
    }

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

    tempMsg.sub(constants.MSG_TREQUEST_CACHE_MESSAGE, () => {

      var payload_byte = base64.toArrayBuffer(window.btoa(""));
      var wrapper = JSON.stringify({
        Type: constants.MSG_TREQUEST_CACHE_MESSAGE,
        Data: Array.from(payload_byte),
      });
      const payload = base64.toArrayBuffer(window.btoa(wrapper))
      util.sendWhenConnected(ws, payload);
    })


    setMsgManager(tempMsg);
    window.addEventListener("resize", () => resize());
    resize();
  }, [navbarRef])

  return (
    <div id="room">
      <ThemeProvider theme={darkTheme}>
        <CssBaseline />
        <div ref={navbarRef}>
          <Navbar />
        </div>
        {msgManager && termSize &&
        <WSTerminal
          className="bg-black"
          msgManager={msgManager}
          width={termSize?.Width ? termSize.Width : -1}
          height={termSize?.Height ? termSize.Height : -1}
        />}
      </ThemeProvider>
    </div>
  );

}

export default Room;
