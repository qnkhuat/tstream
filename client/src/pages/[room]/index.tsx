import React from 'react';
import { RouteComponentProps, withRouter } from "react-router-dom";

import * as utils from "../../utils";
import * as constants from "../../lib/constants";
import * as message from "../../types/message";
import * as pako from "pako";
import PubSub from "../../lib/pubsub";

import Chat from "../../components/Chat";
import Navbar from "../../components/Navbar";
import WSTerminal from "../../components/WSTerminal";
import Uptime from "../../components/Uptime";
import Loading from "../../components/Loading";
import AudioRTC from "../../components/AudioRTC";

import IconButton from '@material-ui/core/IconButton';
import PersonIcon from '@material-ui/icons/Person';
import FullscreenIcon from '@material-ui/icons/Fullscreen';
import FullscreenExitIcon from '@material-ui/icons/FullscreenExit';

// In horizontal mode chat window will have minimum width
const ChatWindowMinWidth: number = 400; // px

// In verticle mode term will have in height
const TermWindowMinHeightRatio: number = 0.6 // %


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

enum Orientation {
  Vertical = "Verticlal",
    Horizontal = "Horizontal",
}

interface RoomInfo {
  StreamerID: string;
  StartedTime:string;
  NViewers: number;
  Title: string;
  Status: RoomStatus;
  Delay: number;
}

interface Params {
  roomID: string;
}

interface Props extends RouteComponentProps<Params> {
}

interface State {
  termSize: RectSize;
  chatSize: RectSize;
  roomInfo: RoomInfo | null;
  mouseMove: boolean;
  connectStatus: RoomStatus;
  fullScreen: boolean | null;
  orientation: Orientation | null;
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
  mouseMovetimeout: ReturnType<typeof setTimeout> | null = null;
  ws: WebSocket | null = null;
  msgManager: PubSub | null = null;

  constructor(props: Props) {
    super(props);

    this.navbarRef = React.createRef<HTMLDivElement>();

    this.state = {
      termSize: {width: 0, height: 0},
      chatSize: {width: 0, height: 0},
      roomInfo: null,
      mouseMove:false,
      connectStatus: RoomStatus.Streaming,
      fullScreen: null,
      orientation: null,
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
    let fullScreenState = this.state.fullScreen === null ? true : !this.state.fullScreen;
    this.setState({fullScreen: fullScreenState});
    this.arrangeTermChat(fullScreenState);
  }

  arrangeTermChat(fullScreen: boolean | null) {
    if (this.navbarRef.current) {

      // Decide fullscreen if hasn't been set
      if (fullScreen === null && window.innerWidth < ChatWindowMinWidth * 2) fullScreen = true;
      else if (fullScreen === null) fullScreen = false;

      // Decide orientation if hasn't been set
      let orientation = this.state.orientation;
      const windowRatio = window.innerWidth / window.innerHeight;
      if (windowRatio < .8) orientation = Orientation.Vertical;
      if (windowRatio >= .8) orientation = Orientation.Horizontal;

      const availableHeight = window.innerHeight - this.navbarRef.current.offsetHeight;

      // first assume we are not in full screen mode
      let termSize: RectSize = {
        width: orientation === Orientation.Horizontal ? window.innerWidth - ChatWindowMinWidth : window.innerWidth,
        height: orientation === Orientation.Horizontal ? availableHeight : availableHeight * TermWindowMinHeightRatio,
      }
      let chatSize: RectSize = {
        width: orientation === Orientation.Horizontal ? window.innerWidth - termSize.width : window.innerWidth,
        height: orientation === Orientation.Horizontal ? availableHeight : availableHeight - termSize.height,
      }

      if (fullScreen) {
        // in case full screen termsize will fill in the space of chat size
        if (orientation === Orientation.Vertical ) termSize.height += chatSize.height;
        if (orientation === Orientation.Horizontal ) termSize.width += chatSize.width;

        chatSize = {
          width:0,
          height:0,
        }
      }

      this.setState({
        termSize: termSize,
        chatSize: chatSize,
        orientation : orientation,
        fullScreen: fullScreen,
      });
    }
  }

  componentWillUnmount() {
    this.ws?.close();
  }


  componentDidMount() {
    const wsUrl = utils.getWsUrl(this.props.match.params.roomID);
    const msgManager = new PubSub();

    // set up websocket connection
    const ws =  new WebSocket(wsUrl);

    // Send client info for server to verify
    let payload = JSON.stringify({
      Type: "ClientInfo",
      Data: {Role: "Viewer"}
    })

    utils.sendWhenConnected(ws, payload);

    ws.onclose = (ev: CloseEvent) => {
      let roomInfo = this.state.roomInfo;
      if (roomInfo && roomInfo.Status !== RoomStatus.NotExisted) {
        roomInfo.Status = RoomStatus.Stopped;
        this.setState({roomInfo: roomInfo});
      }
    }

    // TODO : reconnect on error
    // https://github.com/joewalnes/reconnecting-websocket
    ws.onerror = (ev: Event) => {
      let roomInfo = {} as RoomInfo;
      roomInfo.Status = RoomStatus.NotExisted;
      this.setState({roomInfo: roomInfo});
    }

    ws.onmessage = (ev: MessageEvent) => {
      let msg = JSON.parse(ev.data);

      switch (msg.Type) {

        case constants.MSG_TWRITEBLOCK:
          let blockMsg: message.TermWriteBlock = JSON.parse(window.atob(msg.Data));
          msgManager.pub(msg.Type, blockMsg);
          break;

         case constants.MSG_TWINSIZE:

          msgManager.pub(msg.Type, msg.Data);
          break;

        case constants.MSG_TCHAT:

          msgManager.pub(constants.MSG_TCHAT_IN, msg.Data);
          break;

        case constants.MSG_TROOM_INFO:

          this.setState({roomInfo: msg.Data});
          break;

        default:

          console.error("Unhandled message: ", msg.Type)

      }
    }

    // set up msg manager to manage all in and out request of websocket
    msgManager.sub("request", (msgType: string) => {

      let payload = JSON.stringify({
        Type: msgType,
        Data: "",
      });

      utils.sendWhenConnected(ws, payload);

    })

    msgManager.sub(constants.MSG_TCHAT_OUT, (chat:message.ChatMsg) => {
      let chatList: message.ChatMsg[] = [chat];

      let payload = JSON.stringify({
        Type: constants.MSG_TCHAT,
        Data: chatList,
      });

      utils.sendWhenConnected(ws, payload);
    })

    msgManager.pub("request", constants.MSG_TREQUEST_ROOM_INFO);

    // periodically update roominfo to get number of viewers
    setInterval(() => {
      msgManager.pub("request", constants.MSG_TREQUEST_ROOM_INFO);
    }, 5000);

    this.msgManager = msgManager;
    this.ws = ws;

    this.arrangeTermChat(this.state.fullScreen);
    window.addEventListener('resize', () => {
      this.arrangeTermChat(this.state.fullScreen);
    })

  }

  render() {
    document.title = getSiteTitle(this.props.match.params.roomID, this.state.roomInfo?.Title as string);
    const isConnected = this.state.roomInfo != null;
    const isStreamStopped = this.state.roomInfo?.Status === RoomStatus.Stopped;
    const isRoomExisted = this.state.roomInfo?.Status !== RoomStatus.NotExisted;
    const terminalSize: RectSize = this.state.termSize;
    return (
      <>
        <div id="navbar" ref={this.navbarRef}>
          <Navbar />
        </div>

        {!isConnected && <Loading />}

        {isConnected && this.msgManager &&
          <div id="room" className={`flex relative ${this.state.orientation === Orientation.Horizontal ? "flex-row" : "flex-col"}`}>
            <div id="terminal-view" className="relative"
              onMouseMove={() => this.flashTitle()}>
              {isConnected && !isStreamStopped && isRoomExisted &&
              <div id="info" className={`relative w-full ${this.state.mouseMove ? "visible" : "hidden"}`}>

                <div className={`top-0 left-0 w-full absolute z-10 px-4 py-2 bg-opacity-80 bg-gray-800 `}>
                  <p className="text-2xl">{this.state.roomInfo!.Title}</p>
                  <p className="text-md">@{this.state.roomInfo!.StreamerID}</p>
                </div>

                <div id="info-uptime" className={`p-1 bg-red-400 rounded absolute top-4 right-4 z-10`}>
                  <Uptime className="text-md text-white font-semibold" startTime={new Date(this.state.roomInfo!.StartedTime)} />
                </div>

                <div id="info-nviewers" className="p-1 bg-gray-400 rounded absolute top-4 right-24 z-10">
                  <p className="text-md text-whtie font-semibold"><PersonIcon/> {this.state.roomInfo!.NViewers}</p>
                </div>

              </div>}

              <div id="terminal-window">

                {!isStreamStopped && isRoomExisted && this.state.roomInfo && 
                <WSTerminal
                  className="bg-black"
                  msgManager={this.msgManager}
                  width={terminalSize.width}
                  height={terminalSize.height}
                  delay={this.state.roomInfo.Delay}
                />}

                {isStreamStopped && isRoomExisted &&
                  <div
                    style={terminalSize}
                    className="bg-black flex justify-center items-center">
                    <p className="text-2xl font-bold">The stream has stopped</p>
                  </div>}

                {!isRoomExisted &&
                  <div
                    style={terminalSize}
                    className="bg-black flex justify-center items-center">
                    <p className="text-2xl font-bold">Stream not existed</p>
                  </div>}

                <IconButton className={`absolute bottom-3 right-4 z-10 rounded-full bg-gray-700 p-1`} onClick={this.toggleChatWindow.bind(this)}>
                  {this.state.fullScreen && <FullscreenExitIcon/>}
                  {!this.state.fullScreen && <FullscreenIcon/>}
                </IconButton>

                <AudioRTC className="absolute bottom-4 left-4 z-10" roomID={this.props.match.params.roomID} />
              </div>


            </div>

            <Chat msgManager={this.msgManager} height={this.state.chatSize.height} width={this.state.chatSize.width} 
              className={`${this.state.fullScreen ? "hidden" : "visible"} 
            ${this.state.orientation === Orientation.Horizontal ? "border-l" : "border-t"} border-gray-500 rounded-lg`}/>
          </div>
        }
      </>
    )
  }
}

export default withRouter(Room);
