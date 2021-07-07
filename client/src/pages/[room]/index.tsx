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
import Loading from "../../components/Loading";

import dayjs from "dayjs";

import PersonIcon from '@material-ui/icons/Person';

interface Params {
  username: string;
}

interface RectSize {
  width: number;
  height: number;

}

enum RoomStatus {
  Disconnected = "Disconnected",
    Stopped = "Stopped",
    Streaming = "Streaming",
}

interface RoomInfo {
  StreamerID: string;
  StartedTime:string;
  NViewers: number;
  Title: string;
  RoomStatus: RoomStatus;
}

interface Params {
  username: string;
}

interface Props extends RouteComponentProps<Params> {
}

interface State {
  termSize: RectSize | null;
  roomInfo: RoomInfo | null;
  mouseMove: boolean;
  connectStatus: RoomStatus;
}

class Room extends React.Component<Props, State> {

  navbarRef: React.RefObject<HTMLDivElement>;
  chatWinsize: number = 400; // TODO: implement dynamic chat win size
  mouseMovetimeout: ReturnType<typeof setTimeout> | null;
  ws: WebSocket | null;
  msgManager: PubSub | null;

  constructor(props: Props) {
    super(props);

    this.navbarRef = React.createRef<HTMLDivElement>();
    this.mouseMovetimeout = null;
    this.ws = null;
    this.msgManager = null;


    this.state = {
      termSize: null,
      roomInfo: null,
      mouseMove:false,
      connectStatus: RoomStatus.Streaming,
    };

  }

  flashTitle() {
    this.setState({ mouseMove: true });

    (() => {
      if (this.mouseMovetimeout) clearTimeout(this.mouseMovetimeout);
      this.mouseMovetimeout = setTimeout(() => {
        this.setState({mouseMove:false})
      }, 500);
    })();
  }

  resizeTerminal() {
    if (this.navbarRef.current) {
      this.setState({termSize: {
        width: window.innerWidth - this.chatWinsize,
        height: window.innerHeight - this.navbarRef.current.offsetHeight,
      }});
    }
  }

  componentWillUnmount() {
    this.ws?.close();
  }


  componentDidMount() {
    const wsUrl = util.getWsUrl(this.props.match.params.username);
    const msgManager = new PubSub();

    const ws =  new WebSocket(wsUrl);
    ws.onclose = (ev: CloseEvent) => {
      let roomInfo = this.state.roomInfo;
      if (roomInfo) {
        roomInfo.RoomStatus = RoomStatus.Stopped;
        this.setState({roomInfo: roomInfo});
      }
    }

    ws.onmessage = (ev: MessageEvent) => {
      let msg = JSON.parse(ev.data);

      console.log("Receive message: ", msg.Type);
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
    setInterval(() => {
      msgManager.pub("request", constants.MSG_TREQUEST_ROOM_INFO);
    }, 5000);

    this.msgManager = msgManager;
    this.ws = ws;

    this.resizeTerminal();
    window.addEventListener('resize', () => {
      this.resizeTerminal();
    })
  }


  render() {
    const isStreamStopped = this.state.roomInfo?.RoomStatus == RoomStatus.Stopped;
    const terminalSize: RectSize =  {
      width: this.state.termSize?.width ? this.state.termSize.width : -1,
      height: this.state.termSize?.height ? this.state.termSize.height : -1,
    }
    return (
      <>
        <div id="navbar" ref={this.navbarRef}>
          <Navbar />
        </div>

        <div id="room" className="flex relative">
          {this.msgManager &&
          <>
            <div id="terminal-view" className="relative"
              onMouseMove={() => this.flashTitle()}>
              {this.state.roomInfo && !isStreamStopped &&
              <div id="info">
                <div
                  className={`top-0 left-0 w-full absolute z-10 px-4 py-2 bg-opacity-80 bg-gray-400
                ${this.state.mouseMove ? "visible" : "hidden"}`}>
                  <p className="text-2xl">{this.state.roomInfo.Title}</p>
                  <p className="text-md">@{this.state.roomInfo.StreamerID}</p>
                </div>
                <div id="info-uptime" className="p-1 bg-red-400 rounded absolute top-4 right-4 z-10">
                  <Uptime className="text-md text-white font-semibold" startTime={new Date(this.state.roomInfo.StartedTime)} />
                </div>

                <div id="info-nviewers" className="p-1 bg-gray-400 rounded absolute bottom-4 right-4 z-10">
                  <p className="text-mdtext-whtie font-semibold"><PersonIcon/> {this.state.roomInfo.NViewers}</p>
                </div>
              </div>}

              <div id="terminal-window">
                {!isStreamStopped &&
                <WSTerminal
                  className="bg-black"
                  msgManager={this.msgManager}
                  width={terminalSize.width}
                  height={terminalSize.height}
                />}
                {isStreamStopped &&
                  <div id="closed"
                    style={terminalSize}
                    className="w-full h-full bg-black flex justify-center items-center">
                    <p className="text-2xl font-bold">The stream has stopped</p>
                  </div>}
              </div>
            </div>

            <Chat msgManager={this.msgManager}/>
          </>}

        </div>
      </>
    )
  }
}

export default withRouter(Room);
