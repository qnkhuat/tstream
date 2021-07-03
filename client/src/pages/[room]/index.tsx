import React, { useEffect, useState } from 'react';
import { useParams } from "react-router-dom";

import * as base64 from "../../lib/base64";
import * as util from "../../lib/util";
import WSTerminal from "../../components/WSTerminal";

import StreamPreview from "../../components/StreamPreview";

import * as constants from "../../lib/constants";
interface Params {
  username: string;
}

interface Winsize {
  Width: number;
  Height: number;

}

function Room() {
  const params: Params = useParams();
  const chatWinsize = 400; // px

  const [ inputValue, setInputValue ] = useState("");
  const [ termSize, setTermSize ] = useState<Winsize>();

  function resize() {
    setTermSize({
      Width: window.innerWidth - chatWinsize,
      Height: window.innerHeight,
    });
  }

  const wsUrl = util.getWsUrl(params.username);
  useEffect(() => {
    window.addEventListener("resize", () => resize());
    resize();
  }, [])

  return (
    <div id="room">
      <>
        {termSize &&
        <>
          <WSTerminal
            className="bg-black-900"
            wsUrl={wsUrl}
            width={termSize?.Width ? termSize.Width : -1}
            height={termSize?.Height ? termSize.Height : -1}
          />
        </>
        }
      </>
    </div>
  );

}

export default Room;
