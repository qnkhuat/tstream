import { FC, ReactElement, useState, useEffect } from "react";
import urljoin from "url-join";
import WSTerminal from "./WSTerminal";
import PubSub from "./../lib/pubsub";
import * as base64 from "../lib/base64";
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

function getUpTime(time: dayjs.Dayjs): string {
  const now = dayjs();
  let diff = now.diff(time, "second");

  let hours = Math.floor(diff / 3600);
  diff = diff - 3600 * hours;

  let minutes = Math.floor(diff / 60);
  diff = diff - minutes * 60;

  let seconds = diff;

  return `${hours}:${minutes > 9 ? minutes : `0${minutes}` }:${seconds> 9 ? seconds: `0${seconds}` }`;
}

const StreamPreview: FC<Props> = ({ title, wsUrl, streamerID, nViewers, startedTime, lastActiveTime }): ReactElement => {

  const [ upTime, setUpTime ] = useState(getUpTime(dayjs(startedTime)));
  const msgManager = new PubSub();

  useEffect(() => {
    const ws = new WebSocket(wsUrl);
    msgManager.sub("RequestMessage", () => {
      var payload = JSON.stringify({
      });

      var payload_byte = base64.toArrayBuffer(window.btoa(payload));
      var wrapper = JSON.stringify({
        Type: "RequestWinsize",
        Data: Array.from(payload_byte),
      });
      var test = JSON.parse(wrapper);
      const temp = base64.toArrayBuffer(window.btoa(wrapper))
      ws.send(temp);
      console.log("ws.send successfully");
    });


    setInterval(() => {
      setUpTime(getUpTime(dayjs(startedTime)));
    }, 1000);
  }, [])

  //<WSTerminal wsUrl={wsUrl} height={350} width={500}/>
  return (
    <div className="relative px-4 pt-4 bg-black rounded">

      <div className="p-1 bg-red-400 rounded absolute top-4 right-4">
        <p className="text-mdtext-whtie font-semibold">{upTime}</p>
      </div>

      <div className="absolute bottom-0 left-0 w-full rounded-b bg-gray-600 bg-opacity-90 p-4" >

        <p className="font-semibold">{title}</p>
        <div className="flex justify-between">
          <p className="text-md">@{streamerID}</p>
          <p className="text-md">{nViewers} Viewers</p>
        </div>
      </div>
    </div>
  )

}
export default StreamPreview;

