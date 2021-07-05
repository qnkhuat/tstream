import React, { useEffect, useState, useRef } from 'react';
import { RouteComponentProps, withRouter, Link } from "react-router-dom";

import * as base64 from "../../lib/base64";
import * as util from "../../lib/util";
import * as constants from "../../lib/constants";
import PubSub from "../../lib/pubsub";

import Chat from "../../components/Chat";
import Navbar from "../../components/Navbar";
import WSTerminal from "../../components/WSTerminal";
import Uptime from "../../components/Uptime";

import dayjs from "dayjs";

import CssBaseline from '@material-ui/core/CssBaseline';
import { createTheme, ThemeProvider } from '@material-ui/core/styles';

import PersonIcon from '@material-ui/icons/Person';

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



interface Params {
  username: string;
}

interface Props extends RouteComponentProps<Params> {

}

interface State {
  termSize: Winsize | null;
  msgManager: PubSub | null;
  roomInfo: RoomInfo | null;
  ws: WebSocket | null;
  upTimeStr: string | null;
}

class Room extends React.Component<Props, State> {

  navbarRef: React.RefObject<HTMLDivElement>;
  chatWinsize: number = 400; // TODO: implement dynamic chat win size

  constructor(props: Props) {
    super(props);

    this.navbarRef = React.createRef<HTMLDivElement>();

    const wsUrl = util.getWsUrl(this.props.match.params.username);
    const ws: WebSocket = new WebSocket(wsUrl);

    const msgManager = new PubSub();

    ws.onmessage = (ev: MessageEvent) => {
      let msg = JSON.parse(ev.data);

      if (msg.Type === constants.MSG_TWRITE) {

        var buffer = base64.toArrayBuffer(msg.Data)
        msgManager.pub(msg.Type, buffer);

      } else if (msg.Type === constants.MSG_TWINSIZE) {

        let winSizeMsg = JSON.parse(window.atob(msg.Data));
        msgManager.pub(msg.Type, winSizeMsg);

      } else if (msg.Type === constants.MSG_TCHAT) {

        let encoded_string = "";
        for (let i = 0; i < msg.Data.length; i++) {
          encoded_string = encoded_string.concat(String.fromCharCode(msg.Data[i]));
        }

        let chatMsg = JSON.parse(encoded_string);
        msgManager.pub(msg.Type, chatMsg);
      } else if (msg.Type === constants.MSG_ROOM_INFO) {

        let roomInfo = JSON.parse(window.atob(msg.Data));
        this.setState({roomInfo: roomInfo});
        console.log("Got room info: ", roomInfo);


      }
    }

    msgManager.sub("request", (msgType: string) => {

      var payload_byte = base64.toArrayBuffer(window.btoa(""));
      var wrapper = JSON.stringify({
        Type: msgType,
        Data: Array.from(payload_byte),
      });
      const payload = base64.toArrayBuffer(window.btoa(wrapper));
      util.sendWhenConnected(ws, payload);

    })

    msgManager.sub(constants.MSG_TREQUEST_CHAT, (data) => {
      var payload = JSON.stringify(data);
      var payload_byte = base64.toArrayBuffer(window.btoa(payload));
      var wrapper = JSON.stringify({
        Type: constants.MSG_TCHAT,
        Data: Array.from(payload_byte),
      });
      var msg = base64.toArrayBuffer(window.btoa(wrapper));
      util.sendWhenConnected(ws, msg);
    })

    msgManager.pub("request", constants.MSG_TREQUEST_ROOM_INFO);
    // TODO: remove this manual request from viewer by actively sending from server
    // refresh roomInfo every seconds to update number of viewers
    setInterval(() => {
      msgManager.pub("request", constants.MSG_TREQUEST_ROOM_INFO);
    }, 10000);


    this.state = {
      termSize: null,
      msgManager: msgManager,
      roomInfo: null,
      ws: ws,
      upTimeStr: null,
    };


  }

  componentDidMount() {
    if (this.navbarRef.current) {
      this.setState({termSize: {
        Width: window.innerWidth - this.chatWinsize,
        Height: window.innerHeight - this.navbarRef.current.offsetHeight,
      }})
    }
  }

  render() {
    return (
      <ThemeProvider theme={darkTheme}>
        <CssBaseline />
        <div ref={this.navbarRef}>
          <Navbar />
        </div>
        {this.state.msgManager && this.state.termSize &&
        <div id="room" className="flex">
          <div id="terminal-view" className="relative">

            {this.state.roomInfo &&
            <>
              <div id="info-uptime" className="p-1 bg-red-400 rounded absolute top-4 right-4 z-10">
                <Uptime className="text-md text-white font-semibold" startTime={new Date(this.state.roomInfo.StartedTime)} />
              </div>

              <div id="info-nviewers" className="p-1 bg-gray-400 rounded absolute bottom-4 right-4 z-10">
                <p className="text-mdtext-whtie font-semibold"><PersonIcon/> {this.state.roomInfo.NViewers}</p>
              </div>

            </>
            }

            <WSTerminal
              className="bg-black"
              msgManager={this.state.msgManager}
              width={this.state.termSize?.Width ? this.state.termSize.Width : -1}
              height={this.state.termSize?.Height ? this.state.termSize.Height : -1}
            />

          </div>
          <Chat
            msgManager={this.state.msgManager}
          />
        </div>
        }
      </ThemeProvider>
    )
  }
}

export default withRouter(Room);
