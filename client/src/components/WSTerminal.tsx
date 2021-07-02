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

function base64ToArrayBuffer(input:string): Uint8Array {
  var binary_string =  window.atob(input);
  var len = binary_string.length;
  var bytes = new Uint8Array( len );
  for (var i = 0; i < len; i++)        {
    bytes[i] = binary_string.charCodeAt(i);
  }
  return bytes;
}

const WSTerminal: React.FC<Props> = ({  msgManager, width=-1, height=-1, className=""}) => {
  const termRef = useRef<Xterm>(null);
  const divRef = useRef<HTMLDivElement>(null);
  const [ divSize, setDivSize ]= useState<number[]>([0, 0]); // store rendered size
  const [ transform, setTransform ] = useState<string>("");

  function resize() {
    if (divRef.current && (width > 0 || height > 0)) {
      //divRef.current.style.transform = ``; // reset to normal scale before compute initial size
      const xtermScreens = divRef.current.getElementsByClassName("xterm-screen");
      if (xtermScreens.length > 0) {

        const xtermScreen = xtermScreens[0] as HTMLDivElement;
        const initialWidth = xtermScreen.offsetWidth,
          initialHeight = xtermScreen.offsetHeight,
          widthRatio = initialWidth / width,
          heightRatio = initialHeight / height;

        let scale: number = 0;
        scale = widthRatio > heightRatio ? width / initialWidth : height / initialHeight;
        //divRef.current.style.transform = `scale(${scale}) translate(-50%, -50%)`;
        divRef.current.style.transform = `scale(${scale})`;
        setDivSize([scale * initialWidth, scale*initialHeight])
      }
    }
  }

  msgManager?.sub("Write", (buffer: Uint8Array) => {
    termRef.current?.writeUtf8(buffer);
  })

  //useEffect(() => {
  //  ws.onmessage = (ev: MessageEvent) => {
  //    let msg = JSON.parse(ev.data);

  //    if (msg.Type === "Write") {
  //      var buffer = base64ToArrayBuffer(msg.Data)
  //      termRef.current?.writeUtf8(buffer);
  //    } else if (msg.Type === "Winsize") {
  //      let winSizeMsg = JSON.parse(window.atob(msg.Data))
  //      termRef.current?.resize(winSizeMsg.Cols, winSizeMsg.Rows)

  //      // resize to fit desired size
  //      resize()
  //    }
  //  }

  //  window.addEventListener('resize', () => {resize()});

  //}, [])

  return (
    <div className={`relative ${className}`} style={{
      width: width + "px",
        height: height + "px",
      }}>
      <div ref={divRef}
        //className="absolute top-1/2 left-1/2 origin-top-left">
        className="absolute origin-top-left">
        <Xterm
          options={{rightClickSelectsWord: false}}
          ref={termRef}/>
      </div>
    </div>
  )
}

export default WSTerminal;
