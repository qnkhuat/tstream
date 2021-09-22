import { FC, ReactElement, useState, useEffect } from "react";
import PubSubTerminal from "./PubSubTerminal";
import PubSub from "./../lib/pubsub";
import * as utils from "../utils";
import * as constants from "../lib/constants";
import * as message from "../types/message";

import Uptime from "./Uptime";
import dayjs from "dayjs";
import customParseFormat from 'dayjs/plugin/customParseFormat';

import PersonIcon from '@mui/icons-material/Person';
dayjs.extend(customParseFormat);

interface Props {
  title: string;
  startedTime: string;
  wsUrl: string;
  nViewers: number;
  streamerID: string;
  lastActiveTime?: string;
  width?: number; // in pixel
  height?: number; // in pixel
}

const StreamPreview: FC<Props> = ({ title, wsUrl, streamerID, nViewers, startedTime, lastActiveTime }): ReactElement => {

  const [ msgManager, setMsgManager ] = useState<PubSub>();

  useEffect(() => {
    const ws = new WebSocket(wsUrl);

    // Send client info for server to verify
    let payload = JSON.stringify({
      Type: "ClientInfo",
      Data: {Role: "Viewer"}
    });
    utils.sendWhenConnected(ws, payload);

    const tempMsg = new PubSub();

    ws.onmessage = (ev: MessageEvent) => {
      let msg = JSON.parse(ev.data);

      switch (msg.Type) {
        case constants.MSG_TWRITEBLOCK:
          tempMsg.pub(msg.Type, msg.Data);
          break;

        case constants.MSG_TWINSIZE:
          let winSizeMsg = msg.Data;
          tempMsg.pub(msg.Type, winSizeMsg);
          break;

        default:
          console.log("Unhandled message type: ", msg.Type)
          break;

      }
    }

    tempMsg.sub("request", (msgType: string) => {
      var payload = JSON.stringify({
        Type: msgType,
        Data: "",
      });
      setTimeout(() => {
        utils.sendWhenConnected(ws, payload);
      }, 100);
    })

    setMsgManager(tempMsg);

    // preview doesn't need to be live
    setTimeout(() => {
      ws.close();
    }, 2000);

  }, [])

  return (
    <div className="relative bg-black rounded-lg w-full my-4">
      {msgManager &&
      <PubSubTerminal
        className="bg-black"
        msgManager={msgManager} 
        height={window.innerWidth > 600 ? 350 : 250} 
        width={window.innerWidth > 600 ? 500 : window.innerWidth - 35 }/>
      }
      <div className="p-1 bg-red-400 rounded-lg absolute top-4 right-4">
        <Uptime className="text-md text-white font-semibold" startTime={new Date(startedTime)} />
      </div>

      <div className="absolute bottom-0 left-0 w-full rounded-b-lg bg-gray-600 bg-opacity-90 p-4" >

        <p className="font-semibold">{title}</p>
        <div className="flex justify-between">
          <p className="text-md">@{streamerID}</p>
          <p className="text-md font-bold"> <PersonIcon/> {nViewers}</p>
        </div>
      </div>
    </div>
  )

}
export default StreamPreview;

