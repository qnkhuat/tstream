import React, { ReactElement, useRef, useEffect, useState, CSSProperties } from "react";
import Xterm from "./Xterm";
import * as constants from "../lib/constants";
import * as base64 from "../lib/base64";

// TODO: add handle % and px
interface Props {
  wsUrl: string;
  className?: string;
  width?: number; // in pixel
  height?: number; // in pixel
}

interface Winsize {
  Rows: number;
  Cols: number;
}

function base64ToArrayBuffer(input:string): Uint8Array {
  var binary_string =  window.atob(input);
  var len = binary_string.length;
  var bytes = new Uint8Array( len );
  for (var i = 0; i < len; i++)        {
    bytes[i] = binary_string.charCodeAt(i);
  }
  return bytes;
}

function proposeScale(boundWidth: number, boundHeight: number, realWidth: number, realHeight: number): number {
  const widthRatio = realWidth / boundWidth,
    heightRatio = realHeight / boundHeight;
  if (boundWidth > 0 && boundHeight > 0 ) {
    return  widthRatio > heightRatio ? boundWidth / realWidth : boundHeight / realHeight;
  } else {
    return boundWidth > 0 ? boundWidth / realWidth :  boundHeight / realHeight;
  }
}

const WSTerminal: React.FC<Props> = ({ wsUrl, width= -1, height= -1, className=""}) => {
  console.log("Got: ", width, height);
  const termRef = useRef<Xterm>(null);
  const divRef = useRef<HTMLDivElement>(null);
  const [ divSize, setDivSize ] = useState<number[]>([0, 0]); // store rendered size (width, height)
  const [ transform, setTransform ] = useState<string>("");
  const [ termFontsize, setTermFontsize] = useState<number>(0);

  const [ws, setWebsocket] = useState<WebSocket>();

  function rescale() {

    if (divRef.current && (width > 0 || height > 0)) {

      const xtermScreens = divRef.current.getElementsByClassName("xterm-screen");
      if (xtermScreens.length > 0) {

        const xtermScreen = xtermScreens[0] as HTMLDivElement;
        const initialWidth = xtermScreen.offsetWidth,
          initialHeight = xtermScreen.offsetHeight;

        let scale = proposeScale(width, height, initialWidth, initialHeight);

        divRef.current.style.transform = `scale(${scale}) translate(-50%, -50%)`;
        setDivSize([scale * initialWidth, scale * initialHeight])
      }
    } else {
      console.error("Parent div not found", );
    }
  }

  useEffect(() => {
    const ws = new WebSocket(wsUrl);
    setWebsocket(ws);
  }, []);

  useEffect(() => {
    if (ws) {
      ws.onmessage = (ev: MessageEvent) => {
        let msg = JSON.parse(ev.data);

        if (msg.Type === constants.MSG_TWRITE) {
          var buffer = base64.toArrayBuffer(msg.Data)
          termRef.current?.writeUtf8(buffer);
        } else if (msg.Type === constants.MSG_TWINSIZE) {
          let winsize = JSON.parse(window.atob(msg.Data));
          termRef.current?.resize(winsize.Cols, winsize.Rows)
          rescale();
        }
      }
    }
    window.addEventListener("resize", () => rescale());
  }, [height, width, ws]);

  const s = {
    width: width > 0 ? width + "px" : divSize[0] + "px",
    height: height > 0 ? height + "px" : divSize[1] + "px",
  }

  return (
    <div className={`relative ${className}`} style={s}>
      <div ref={divRef}
        className="divref absolute top-1/2 left-1/2 origin-top-left"
      >
        <Xterm
          options={{
            rightClickSelectsWord: false,
          }}
          ref={termRef}/>
      </div>
    </div>
  )
}

export default WSTerminal;
