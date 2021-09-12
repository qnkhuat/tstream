import React, { useEffect } from "react";
import Xterm from "./Xterm";
import * as constants from "../lib/constants";
import * as message from "../types/message";
import * as buffer from "../lib/buffer";
import pako from "pako";

interface Props {
  width: number; // in pixel
  height: number; // in pixel
  rows?: number;
  cols?: number;
  delay?: number;
  className?: string;
}

const Terminal = React.forwardRef<Xterm, Props>(({ width=-1, height=-1, rows=0, cols = 0, className = "" }, ref) => {

  useEffect(() => {
    const rescale = () => {
      if (ref && typeof ref != "function" && ref.current && (width! > 0 || height! > 0)) {
        const core = (ref.current?.terminal as any)._core,
          cellWidth = core._renderService.dimensions.actualCellWidth,
          cellHeight = core._renderService.dimensions.actualCellHeight,
          currentFontSize = ref.current.terminal.getOption('fontSize'),
          termRows = rows > 0 ? rows : ref.current.terminal.rows,
          termCols = cols > 0 ? cols : ref.current.terminal.cols,
          hFontSizeMultiplier = height / (cellHeight * termRows),
          wFontSizeMultiplier = width / (cellWidth * termCols),
          // method doesn't ensure termianl will 100% fit the required size since fontsize are discrete
          // Another method is to transform scale to fit the window
          // But I haven't figured out why the scaled version sometimes make terminal deformed 
          // after multiple times of apply scale transformation
          newFontSize = Math.floor(
            hFontSizeMultiplier > wFontSizeMultiplier 
            ? currentFontSize * wFontSizeMultiplier 
            : currentFontSize * hFontSizeMultiplier);

        ref.current.terminal.setOption('fontSize', newFontSize);
        ref.current.terminal.refresh(0, termRows - 1); // force xterm to re-render everything
      }
    }

    const handleResize = () => { 
      rescale(); 
    };

    window.addEventListener("resize", handleResize);

    // first refresh
    handleResize();
    return () => {
      window.removeEventListener("resize", handleResize);
    }
  }, [width, height, ref, rows, cols]);


  return <div className={`relative overflow-hidden bg-black ${className}`}
    style={{width: width!, height: height!}}>
    <div className="overlay bg-transparent absolute top-0 left-0 z-10 w-full h-full"></div>
    <div className="divref absolute top-1/2 left-1/2 origin-top-left transform -translate-x-1/2 -translate-y-1/2 overflow-hidden">
      <Xterm 
        ref={ref} 
        options={{
          rightClickSelectsWord: false,
            disableStdin: true,
        }}/>
    </div>
  </div>
});

export class WriteManager {

  queue: message.Wrapper[] = [];
  writeCB: (arr:Uint8Array) => void;
  winsizeCB: (ws:message.TermSize) => void;
  delay: number; // in milliseconds
  startTime: number | undefined; // the start time of the stream. used to calculate the duration of stream or records
  currentTime: number | undefined; // currenttime of the player
  refreshInterval: number; // the writemanager will scan and schedule to write after each refreshInterval. Unit in milliseconds

  constructor(writeCB: (arr: Uint8Array) => void, winsizeCB: (ws: message.TermSize) => void, delay: number = 0) {
    this.writeCB = writeCB;
    this.winsizeCB = winsizeCB;
    this.delay = delay;
    this.refreshInterval = 200; // TODO: move this to optional arguments
    this.queue = [];
    this.consume();
  }

  play() {
    //this.play
  }

  pause() {

  }

  jumpTo() {

  }


  resetQueue() {
    this.queue = [];
  }

  addQueue(q: message.Wrapper[]) {
    this.queue.push(...q);
  }

  consume(): any {
    if(!this.currentTime) return setTimeout(() => this.consume(), this.refreshInterval);
    const startTime = this.currentTime!;
    const endTime = startTime + this.refreshInterval;

    const returnCallback = () => {
      this.currentTime = endTime;
      setTimeout(() => {
        this.currentTime = startTime + this.refreshInterval;
        this.consume();
      }, this.refreshInterval);
    }

    if (!this.play || this.queue.length == 0) {
      return returnCallback();
    }


    while (this.queue.length > 0 && this.queue[0].Delay < endTime) {
      if (this.queue[0].Delay > endTime) break;

      const msg: message.Wrapper = this.queue.shift()!;
      const msgTimeout = msg.Delay - startTime;

      switch (msg.Type) {

        case constants.MSG_TWRITE:
          let bufferData = buffer.str2ab(msg.Data);
          setTimeout(() => this.writeCB(bufferData), msgTimeout);
          break;

        case constants.MSG_TWINSIZE:
          setTimeout(() => this.winsizeCB(msg.Data), msgTimeout);
          break;

        default:
          console.error("Unhandled message type: ", msg.Type);
      }
    }

    return returnCallback();
  }

  addBlock(block: message.TermWriteBlock) {
    // the starttime of stream or records will be the the starttime of the first block received
    if (!this.startTime) {
      this.startTime = (new Date(block.StartTime)).getTime();
      this.currentTime = (new Date()).getTime() - this.startTime - (this.delay - block.Duration);
    }

    const blockDelayTime = (new Date(block.StartTime)).getTime() - this.startTime;

    // this is a big chunk of encoding/decoding
    // Since we have to : reduce message size by usign gzip and also
    // every single termwrite have to be decoded, or else the rendering will screw up
    // the whole block often took 9-20 milliseconds to decode a 3 seconds block of message
    let data = pako.ungzip(buffer.str2ab(block.Data), { to: "string" });
    let msgArrayString: string[] = JSON.parse(data);

    let msgArray: message.Wrapper[] = [];
    msgArrayString.forEach((msgString: string) => {
      // re-compute the offset of this message with respect to the render time
      let msg: message.Wrapper = JSON.parse(window.atob(msgString));
      msg.Delay = blockDelayTime + msg.Delay;
      msgArray.push(msg);
    })

    this.addQueue(msgArray);

  }
}


export default Terminal;
