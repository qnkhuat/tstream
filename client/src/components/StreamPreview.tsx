import { FC, ReactElement, useState, useEffect } from "react";
import WSTerminal from "./WSTerminal";
import PubSub from "./../lib/pubsub";
import * as base64 from "../lib/base64";
import * as util from "../lib/util";
import * as constants from "../lib/constants";
import * as message from "../lib/message";

import dayjs from "dayjs";
import customParseFormat from 'dayjs/plugin/customParseFormat';

import PersonIcon from '@material-ui/icons/Person';
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

  const [ upTime, setUpTime ] = useState(util.formatDuration(dayjs(), dayjs(startedTime)));
  const [ msgManager, setMsgManager ] = useState<PubSub>();

  useEffect(() => {
    const ws = new WebSocket(wsUrl);

    // Send client info for server to verify
    let payload = JSON.stringify({
      Type: "ClientInfo",
      Data: {Role: "Viewer"}
    });
    util.sendWhenConnected(ws, payload);

    const tempMsg = new PubSub();
    ws.onmessage = (ev: MessageEvent) => {
      let msg = JSON.parse(ev.data);

      switch (msg.Type) {
        case constants.MSG_TWRITEBLOCK:
          let blockMsg: message.TermWriteBlock = JSON.parse(window.atob(msg.Data));
          tempMsg.pub(msg.Type, blockMsg);
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
      util.sendWhenConnected(ws, payload);
    })

    setMsgManager(tempMsg);

    setInterval(() => {
      setUpTime(util.formatDuration(dayjs(), dayjs(startedTime)));
    }, 1000);

    // preview doesn't need to be live
    setTimeout(() => {
      ws.close();
    }, 2000);

  }, [])


  return (
    <div className="relative bg-black rounded-lg w-full my-4">
      {msgManager &&
      <WSTerminal msgManager={msgManager} height={window.innerWidth > 600 ? 350 : 250} width={window.innerWidth > 600 ? 500 : window.innerWidth - 35 }/>
      }
      <div className="p-1 bg-red-400 rounded-lg absolute top-4 right-4">
        <p className="text-mdtext-whtie font-semibold">{upTime}</p>
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

