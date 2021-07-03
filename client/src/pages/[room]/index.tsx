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

interface RoomInfo {
  StreamerID: string;
  StartedTime:string;
  NViewers: number;
  Title: string;
}



function Room() {
  const params: Params = useParams();
  const chatWinsize = 400; // px

  const [ termSize, setTermSize ] = useState<Winsize>();
  const [ msgManager, setMsgManager ] = useState<PubSub>();
  const navbarRef = useRef<HTMLDivElement>(null);
  const [ updateState, setUpdate ] = useState(0);
  const [ roomInfo, setRoomInfo ] = useState();

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
      console.log("Got message: ", msg.Type);

      if (msg.Type === constants.MSG_TWRITE) {

        var buffer = base64.toArrayBuffer(msg.Data)
        tempMsg.pub(msg.Type, buffer);

      } else if (msg.Type === constants.MSG_TWINSIZE) {

        let winSizeMsg = JSON.parse(window.atob(msg.Data));
        tempMsg.pub(msg.Type, winSizeMsg);

      } else if (msg.Type === constants.MSG_ROOM_INFO) {
        let roomInfo = JSON.parse(window.atob(msg.Data));
        setRoomInfo(roomInfo);
      }
    }

    tempMsg.sub("request", (msgType: string) => {
      var payload_byte = base64.toArrayBuffer(window.btoa(""));
      var wrapper = JSON.stringify({
        Type: msgType,
        Data: Array.from(payload_byte),
      });
      const payload = base64.toArrayBuffer(window.btoa(wrapper))
      util.sendWhenConnected(ws, payload);
    })


    setMsgManager(tempMsg);
    window.addEventListener("resize", () => resize());
    resize();
  }, [navbarRef])
  console.log("set update: ", updateState);

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
