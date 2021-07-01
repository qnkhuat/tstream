import { FC, ReactElement } from "react";
import urljoin from "url-join";
import WSTerminal from "./WSTerminal";

interface Props {
  title: string;
  wsUrl?: string;
  streamerID?: string;
  startedTime?: string;
  lastActiveTime?: string;
  width?: number; // in pixel
  height?: number; // in pixel
}

const StreamerPreview: FC<Props> = ({ title, wsUrl, streamerID, startedTime, lastActiveTime }): ReactElement => {
  return (
    <div className="border-2 border-red-400 w-full h-full">
      <h3>{title}</h3>
      <h3>{urljoin(process.env.REACT_APP_API_URL as string, "ws", streamerID as string, "viewer")}</h3>
      <h3>{streamerID}</h3>
      <h3>{wsUrl}</h3>
      <h3>{startedTime}</h3>
      <h3>{lastActiveTime}</h3>
      <WSTerminal wsUrl={wsUrl as string}
      />
    </div>
  )

}
export default StreamerPreview;

