import React from 'react';
import { RouteComponentProps, withRouter } from "react-router-dom";

import * as base64 from "../../lib/base64";
import * as util from "../../lib/util";
import * as constants from "../../lib/constants";
import PubSub from "../../lib/pubsub";

import Chat, { ChatMsg } from "../../components/Chat";
import Navbar from "../../components/Navbar";
import WSTerminal from "../../components/WSTerminal";
import Uptime from "../../components/Uptime";
import Loading from "../../components/Loading";
import AudioRTC from "../../components/AudioRTC";

import IconButton from '@material-ui/core/IconButton';
import PersonIcon from '@material-ui/icons/Person';
import DoubleArrowRoundedIcon from '@material-ui/icons/DoubleArrowRounded';

interface Params {
  roomID: string;
}

interface RectSize {
  width: number;
  height: number;
}

enum RoomStatus {
  NotExisted = "NotExisted",
    Disconnected = "Disconnected",
    Stopped = "Stopped",
    Streaming = "Streaming",
}

interface RoomInfo {
  StreamerID: string;
  StartedTime:string;
  NViewers: number;
  Title: string;
  Status: RoomStatus;
}

interface Params {
  roomID: string;
}

interface Props extends RouteComponentProps<Params> {
}

interface State {
  termSize: RectSize | null;
  roomInfo: RoomInfo | null;
  mouseMove: boolean;
  connectStatus: RoomStatus;
  hideChat: boolean | null;
}

function getSiteTitle(streamerId: string, title: string) {
  let siteTitle: string = streamerId;
  if (title) {
    siteTitle += ` - ${title}`;
  }
  return siteTitle;
}

class Room extends React.Component<Props, State> {

  navbarRef: React.RefObject<HTMLDivElement>;
  ChatWindowWidth: number = 400; // TODO: implement dynamic chat win size
  mouseMovetimeout: ReturnType<typeof setTimeout> | null = null;
  ws: WebSocket | null = null;
  msgManager: PubSub | null = null;

  constructor(props: Props) {
    super(props);

    this.navbarRef = React.createRef<HTMLDivElement>();

    this.state = {
      termSize: null,
      roomInfo: null,
      mouseMove:false,
      connectStatus: RoomStatus.Streaming,
      hideChat: null,
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

  toggleChatWindow() {
    let hideChatState = this.state.hideChat == null ? true : !this.state.hideChat;
    this.setState({hideChat: hideChatState});
    this.resizeTerminal(hideChatState);
  }

  resizeTerminal(hideChat: boolean | null) {
    if (this.navbarRef.current) {
      let termWidth = window.innerWidth;

      // at start, if the window is too small then we don't show chat
      if (window.innerWidth < this.ChatWindowWidth * 2 && hideChat == null) hideChat=true;

      if (!hideChat) {
        termWidth -= this.ChatWindowWidth;
      }

      this.setState({
        termSize: {
          width: termWidth,
          height: window.innerHeight - this.navbarRef.current.offsetHeight},
        hideChat: hideChat,
      });
    }
  }

  componentWillUnmount() {
    this.ws?.close();
  }


  componentDidMount() {
    const wsUrl = util.getWsUrl(this.props.match.params.roomID);
    const msgManager = new PubSub();

    // set up websocket connection
    const ws =  new WebSocket(wsUrl);

    // Send client info for server to verify
    let payload = JSON.stringify({
      Type: "ClientInfo",
      Data: {Role: "Viewer"}
    })

    util.sendWhenConnected(ws, payload);

    ws.onclose = (ev: CloseEvent) => {
      let roomInfo = this.state.roomInfo;
      if (roomInfo && roomInfo.Status != RoomStatus.NotExisted) {
        roomInfo.Status = RoomStatus.Stopped;
        this.setState({roomInfo: roomInfo});
      }
    }

    ws.onerror = () => {
      let roomInfo = {} as RoomInfo;
      roomInfo.Status = RoomStatus.NotExisted;
      this.setState({roomInfo: roomInfo});
    }

    ws.onmessage = (ev: MessageEvent) => {
      let msg = JSON.parse(ev.data);

      if (msg.Type === constants.MSG_TWRITE) {

        let buffer = base64.str2ab(JSON.parse(window.atob(msg.Data)).Data);
        msgManager.pub(msg.Type, buffer);

      } else if (msg.Type === constants.MSG_TWINSIZE) {

        msgManager.pub(msg.Type, msg.Data);

      } else if (msg.Type === constants.MSG_TCHAT) {

        msgManager.pub(constants.MSG_TCHAT_IN, msg.Data);

      } else if (msg.Type === constants.MSG_TROOM_INFO) {

        this.setState({roomInfo: msg.Data});

      }
    }

    // set up msg manager to manage all in and out request of websocket
    msgManager.sub("request", (msgType: string) => {

      let payload = JSON.stringify({
        Type: msgType,
        Data: "",
      });

      util.sendWhenConnected(ws, payload);

    })

    msgManager.sub(constants.MSG_TCHAT_OUT, (chat:ChatMsg) => {
      let chatList: ChatMsg[] = [chat];

      let payload = JSON.stringify({
        Type: constants.MSG_TCHAT,
        Data: chatList,
      });

      util.sendWhenConnected(ws, payload);
    })

    msgManager.pub("request", constants.MSG_TREQUEST_ROOM_INFO);

    // periodically update roominfo to get number of viewers
    setInterval(() => {
      msgManager.pub("request", constants.MSG_TREQUEST_ROOM_INFO);
    }, 5000);

    this.msgManager = msgManager;
    this.ws = ws;

    this.resizeTerminal(this.state.hideChat);
    window.addEventListener('resize', () => {
      this.resizeTerminal(this.state.hideChat);
    })

  }

  render() {
    document.title = getSiteTitle(this.props.match.params.roomID, this.state.roomInfo?.Title as string);
    const isConnected = this.state.roomInfo != null;
    const isStreamStopped = this.state.roomInfo?.Status === RoomStatus.Stopped;
    const isRoomExisted = this.state.roomInfo?.Status !== RoomStatus.NotExisted;
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
              {isConnected && !isStreamStopped && isRoomExisted &&
              <div id="info" >

                <div
                  className={`top-0 left-0 w-full absolute z-10 px-4 py-2 bg-opacity-80 bg-gray-800 ${this.state.mouseMove ? "visible" : "hidden"}`}>
                  <p className="text-2xl">{this.state.roomInfo!.Title}</p>
                  <p className="text-md">@{this.state.roomInfo!.StreamerID}</p>
                </div>
                <div id="info-uptime" className={`p-1 bg-red-400 rounded absolute top-4 ${this.state.hideChat ? "right-14" : "right-4"} z-10`}>
                  <Uptime className="text-md text-white font-semibold" startTime={new Date(this.state.roomInfo!.StartedTime)} />
                </div>

                <div id="info-nviewers" className="p-1 bg-gray-400 rounded absolute bottom-4 right-4 z-10">
                  <p className="text-md text-whtie font-semibold"><PersonIcon/> {this.state.roomInfo!.NViewers}</p>
                </div>

                <AudioRTC className="absolute bottom-4 left-4 z-10" roomID={this.props.match.params.roomID} />
              </div>}

              <div id="terminal-window">
                {!isConnected && <Loading />}

                {isConnected && !isStreamStopped && isRoomExisted &&
                  <WSTerminal
                    className="bg-black"
                    msgManager={this.msgManager}
                    width={terminalSize.width}
                    height={terminalSize.height}
                  />
                }

                {isConnected && isStreamStopped && isRoomExisted &&
                  <div
                    style={terminalSize}
                    className="bg-black flex justify-center items-center">
                    <p className="text-2xl font-bold">The stream has stopped</p>
                  </div>}

                {isConnected && !isRoomExisted &&
                  <div
                    style={terminalSize}
                    className="bg-black flex justify-center items-center">
                    <p className="text-2xl font-bold">Stream not existed</p>
                  </div>}


              </div>
            </div>

            <div className={`relative ${this.state.hideChat ? "w-0" : "w-full"}`}>
              <IconButton className={`absolute ${this.state.hideChat ? "right-2 transform rotate-180" : "left-2"} top-3 z-10 `} onClick={this.toggleChatWindow.bind(this)}>
                <DoubleArrowRoundedIcon />
              </IconButton>
              <Chat msgManager={this.msgManager} className={`${this.state.hideChat ? "hidden" : ""}`}/>
            </div>
          </>}

        </div>
      </>
    )
  }
}

export default withRouter(Room);
