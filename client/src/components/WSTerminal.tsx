import React, { useRef, useEffect, useState } from "react";
import Xterm from "./Xterm";
import * as constants from "../lib/constants";
import PubSub from "../lib/pubsub";

// TODO: add handle % and px
interface Props {
  msgManager: PubSub;
  className?: string;
  width?: number; // in pixel
  height?: number; // in pixel
}

interface Winsize {
  Rows: number;
  Cols: number;
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

const WSTerminal: React.FC<Props> = ({ msgManager, width= -1, height= -1, className=""}) => {
  const termRef = useRef<Xterm>(null);
  const divRef = useRef<HTMLDivElement>(null);
  const [ divSize, setDivSize ] = useState<number[]>([0, 0]); // store rendered size (width, height)

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
    msgManager?.sub(constants.MSG_TWRITE, (buffer: Uint8Array) => {
      termRef.current?.writeUtf8(buffer);
    })

    msgManager?.sub(constants.MSG_TWINSIZE, (winsize: Winsize) => {
      termRef.current?.resize(winsize.Cols, winsize.Rows)
      rescale();
    })

    msgManager?.pub(constants.MSG_TREQUEST_WINSIZE, null);
  }, []);


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
