import { FC, ReactElement } from "react";
import urljoin from "url-join";
interface Props {
  title: string;
  wsUrl?: string;
  streamerID?: string;
  startedTime?: string;
  lastActiveTime?: string;
}

const StreamerPreview: FC<Props> = ({ title, wsUrl, streamerID, startedTime, lastActiveTime }): ReactElement => {
  return (
    <div className="border-2 border-red-400">
      <h3>{title}</h3>
      <h3>{urljoin(process.env.REACT_APP_API_URL as string, "ws", streamerID, "viewer")}</h3>
      <h3>{streamerID}</h3>
      <h3>{wsUrl}</h3>
      <h3>{startedTime}</h3>
      <h3>{lastActiveTime}</h3>
    </div>
  )

}
export default StreamerPreview;

