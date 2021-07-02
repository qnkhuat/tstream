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
  //<WSTerminal wsUrl={wsUrl as string}
  //      width={600}
  //      height={337}
  //      className="bg-black"
  //    />
  return (
    <div className="relative p-4"
    >

      <div className="absolute bottom-0 left-0 w-full z-10 bg-gray-400 bg-opacity-50 p-4" >
        <h3>{streamerID}</h3>
      </div>
    </div>
  )

}
export default StreamerPreview;

