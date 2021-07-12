import { FC, ReactElement, useState, useEffect, useRef } from "react";
import WSTerminal from "./WSTerminal";
import PubSub from "./../lib/pubsub";
import * as base64 from "../lib/base64";
import * as util from "../lib/util";
import * as constants from "../lib/constants";

import dayjs from "dayjs";
import customParseFormat from 'dayjs/plugin/customParseFormat';

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

    setInterval(() => {
      setUpTime(util.formatDuration(dayjs(), dayjs(startedTime)));
    }, 1000);

    // preview doesn't need to be live
    setTimeout(() => {
      ws.close();
    }, 1000);

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
          <p className="text-md hidden">{nViewers} Viewers</p>
        </div>
      </div>
    </div>
  )

}
export default StreamPreview;

