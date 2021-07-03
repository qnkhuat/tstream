import React, { useEffect, useState, useRef } from 'react';
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

  const parentDiv = useRef<HTMLDivElement>(null);
  const layoutRef = useRef<HTMLDivElement>(null);

  function resize() {
    if (parentDiv.current) {
      setTermSize({
        Width: parentDiv.current.offsetWidth - chatWinsize,
        Height: parentDiv.current.offsetHeight,
      });
    }
  }

  const wsUrl = util.getWsUrl(params.username);
  useEffect(() => {
    window.addEventListener("resize", () => resize());
    resize();
  }, [parentDiv])

  return (
    <div id="room"
      className ="h-screen w-screen"
      ref={parentDiv}>
      {termSize &&
      <WSTerminal
        className="bg-red-500"
        wsUrl={wsUrl}
        width={termSize?.Width ? termSize.Width : -1}
        height={termSize?.Height ? termSize.Height : -1}
      />
      }
    </div>
  );

}

export default Room;
