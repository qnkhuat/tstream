import React, { useEffect, useState, useRef } from 'react';
import { RouteComponentProps, withRouter, Link } from "react-router-dom";

import * as base64 from "../../lib/base64";
import * as util from "../../lib/util";
import * as constants from "../../lib/constants";
import PubSub from "../../lib/pubsub";

import WSTerminal from "../../components/WSTerminal";
import Chat from "../../components/Chat";
import Navbar from "../../components/Navbar";

import dayjs from "dayjs";

import CssBaseline from '@material-ui/core/CssBaseline';
import { createTheme, ThemeProvider } from '@material-ui/core/styles';

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

    this.state = {
      termSize: null,
      msgManager: msgManager,
      roomInfo: null,
      ws: ws,
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
      <div id="room">
        <ThemeProvider theme={darkTheme}>
          <CssBaseline />
          <div
            ref={this.navbarRef}
          >
            <Navbar />
          </div>
          {this.state.msgManager && this.state.termSize &&
          <div id="terminal-view" className="relative flex">
            <div className="p-1 bg-red-400 rounded absolute top-4 right-4">
              {this.state.roomInfo &&
              <p className="text-mdtext-whtie font-semibold">{util.getUpTime(dayjs(this.state.roomInfo.StartedTime))}</p>}
            </div>

            <WSTerminal
              className="bg-black"
              msgManager={this.state.msgManager}
              width={this.state.termSize?.Width ? this.state.termSize.Width : -1}
              height={this.state.termSize?.Height ? this.state.termSize.Height : -1}
            />
            <Chat
              msgManager={this.state.msgManager}
            />
          </div>
          }
        </ThemeProvider>
      </div>
    )
  }
}

//function Room() {
//  const params: Params = useParams();
//  const chatWinsize = 400; // px
//
//  const navbarRef = useRef<HTMLDivElement>(null);
//  const [ termSize, setTermSize ] = useState<Winsize>();
//  const [ msgManager, setMsgManager ] = useState<PubSub>();
//  const [ updateState, setUpdate ] = useState(0);
//  const [ roomInfo, setRoomInfo ] = useState();
//  const [ upTime, setUptime ] = useState<string>("");
//
//  function resize() {
//    if (navbarRef.current) {
//      setTermSize({
//        Width: window.innerWidth - chatWinsize,
//        Height: window.innerHeight - navbarRef.current.offsetHeight,
//      });
//    }
//
//  }
//
//  useEffect(() => {
//    const wsUrl = util.getWsUrl(params.username);
//    const ws = new WebSocket(wsUrl);
//
//    const tempMsg = new PubSub();
//
//    ws.onmessage = (ev: MessageEvent) => {
//      let msg = JSON.parse(ev.data);
//
//      if (msg.Type === constants.MSG_TWRITE) {
//
//        var buffer = base64.toArrayBuffer(msg.Data)
//        tempMsg.pub(msg.Type, buffer);
//
//      } else if (msg.Type === constants.MSG_TWINSIZE) {
//
//        let winSizeMsg = JSON.parse(window.atob(msg.Data));
//        tempMsg.pub(msg.Type, winSizeMsg);
//
//      } else if (msg.Type === constants.MSG_ROOM_INFO) {
//        let roomInfo = JSON.parse(window.atob(msg.Data));
//        setRoomInfo(roomInfo);
//        console.log("Got room info: ", roomInfo);
//      }
//    }
//
//    tempMsg.sub("request", (msgType: string) => {
//      var payload_byte = base64.toArrayBuffer(window.btoa(""));
//      var wrapper = JSON.stringify({
//        Type: msgType,
//        Data: Array.from(payload_byte),
//      });
//      const payload = base64.toArrayBuffer(window.btoa(wrapper));
//      util.sendWhenConnected(ws, payload);
//    })
//
//
//    tempMsg.pub("request", constants.MSG_TREQUEST_ROOM_INFO);
//
//
//    setMsgManager(tempMsg);
//    window.addEventListener("resize", () => resize());
//    resize();
//  }, [])
//
//  return (

//  );
//
//}
//
export default withRouter(Room);
