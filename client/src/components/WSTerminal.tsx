import React, { ReactElement, useRef, useEffect, useState, CSSProperties } from "react";
import Xterm from "./Xterm";
import PubSub from "./../lib/pubsub";

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

const WSTerminal: React.FC<Props> = ({  msgManager, width= -1, height= -1, className=""}) => {
  console.log("Got: ", width, height);
  const termRef = useRef<Xterm>(null);
  const divRef = useRef<HTMLDivElement>(null);
  const [ divSize, setDivSize ]= useState<number[]>([0, 0]); // store rendered size (width, height)
  const [ transform, setTransform ] = useState<string>("");

  function rescale() {
    console.log("YOOOOOO");

    if (divRef.current && (width > 0 || height > 0)) {

      const xtermScreens = divRef.current.getElementsByClassName("xterm-screen");
      console.log("Do you find me?");
      if (xtermScreens.length > 0) {

        const xtermScreen = xtermScreens[0] as HTMLDivElement;
        const initialWidth = xtermScreen.offsetWidth,
          initialHeight = xtermScreen.offsetHeight;


        let scale = proposeScale(width, height, initialWidth, initialHeight);
        console.log("New scale: ", scale);
        //divRef.current.style.transform = `scale(${scale}) translate(-50%, -50%)`;
        divRef.current.style.transform = `scale(${scale})`;
        setDivSize([scale * initialWidth, scale*initialHeight])
      } else {
        console.log("Fuck no");
      }
    } else {
      console.log("Fuck no ooooooooooooooo: ", width, height);
    }
  }


  useEffect(() => {

    msgManager?.sub("Write", (buffer: Uint8Array) => {
      termRef.current?.writeUtf8(buffer);
    })

    msgManager?.sub("Winsize", (winsize: Winsize) => {
      termRef.current?.resize(winsize.Cols, winsize.Rows)
      rescale();
    })

    rescale();
    window.addEventListener('resize', () => rescale());
  }, [])


  return (
    <div className={`relative ${className}`} style={{
      width: width > 0 ? width + "px" : divSize[0] + "px",
        height: height > 0 ? height + "px" : divSize[1] + "px",
      }}>
      <div ref={divRef}
        //className="divref absolute top-1/2 left-1/2 origin-top-left">
        className="divref absolute origin-top-left">
        <Xterm
          options={{rightClickSelectsWord: false}}
          ref={termRef}/>
      </div>
    </div>
  )
}

export default WSTerminal;
