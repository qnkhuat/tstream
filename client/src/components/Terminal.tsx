// A Wrapper of XTerm where it resize itself based on the provided with and height
import React, { useState, useEffect, useRef } from "react";
import Xterm from "./Xterm";
import * as constants from "../lib/constants";
import * as message from "../types/message";
import * as buffer from "../lib/buffer";
import pako from "pako";
import { accurateInterval } from "../utils";

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

interface WriteMangerOptions {
  delay?: number;
  playing?: boolean;
  refreshInterval?: number;
}

export class WriteManager {

  termRef: Xterm | null;
  queue: message.Wrapper[] = [];
  delay: number; // in milliseconds
  startTime: number | null = null;
  currentTime: number | null = null;
  playing: boolean = false;
  refreshInterval: number;
  clearConsumeInterval?: () => void | undefined;

  constructor(termRef: Xterm, { delay = 0, playing = false, refreshInterval = 200 }: WriteMangerOptions) {
    this.refreshInterval = refreshInterval;
    this.termRef = termRef;
    this.delay = delay;
    if( playing ) this.play();
  }

  resetQueue() {
    this.queue = [];
  }

  addQueue(q: message.Wrapper[]) {
    this.queue.push(...q); // Concatnate
  }

  detach() {
    this.termRef = null;
    this.pause();
  }

  play() {
    if (!this.clearConsumeInterval) {
      this.clearConsumeInterval= accurateInterval(() => {
        this.playing = true;
        this.consume();
      }, this.refreshInterval, {immediate: true});
    }
  }

  pause() {
    if(this.clearConsumeInterval) this.clearConsumeInterval();
    this.playing = false;
  }

  consume() {
    console.log("------------");
    console.log("Startime: ", this.startTime);
    console.log("CurrentTime: ", this.currentTime);
    console.log("Queue len: ", this.queue.length);
    console.log("First queue: ", this.queue[0]?.Delay);
    
    if(!this.playing || !this.termRef) return;

    const startTime = Date.now();
    const returnCallback = () => {
      console.log("Processing time: ", (Date.now() - startTime));
      if(this.currentTime) this.currentTime = this.currentTime + (Date.now() - startTime) + this.refreshInterval;
    }

    if (!this.currentTime || this.queue.length == 0) return returnCallback();

    const currentTime = this.currentTime;
    const endTime = currentTime + this.refreshInterval;

    while (this.queue.length > 0 && this.queue[0].Delay < endTime) {
      if (this.queue[0].Delay > endTime) break;

      const msg: message.Wrapper = this.queue.shift()!;
      const msgTimeout = msg.Delay - currentTime;

      switch (msg.Type) {

        case constants.MSG_TWRITE:
          let bufferData = buffer.str2ab(msg.Data);
          setTimeout(() => this.termRef!.writeUtf8(bufferData), msgTimeout);
          break;

        case constants.MSG_TWINSIZE:
            setTimeout(() => this.termRef!.resize(msg.Data.Cols, msg.Data.Rows), msgTimeout);
          break;

        default:
            console.error("Unhandled message type: ", msg.Type);
      }
    }

    return returnCallback();
  }

  addBlock(block: message.TermWriteBlock) {
    //console.log('incoming data');
    // the starttime of stream or records will be the the starttime of the first block received
    if (!this.startTime || !this.currentTime) {
      const blockStartTime = (new Date(block.StartTime)).getTime();
      this.startTime = blockStartTime;
      // the delay is how much the block message sent to server is delayed to its start time
      // it means that the delay will be include the block duration time and a bit of buffer
      // For example : delay = 1.5s and block duration 1s. we have .5 second for the delay of network
      this.currentTime = (new Date()).getTime() - blockStartTime - (this.delay! - block.Duration);
    }

    const blockDelayTime = (new Date(block.StartTime)).getTime() - this.startTime;
    console.log("Block delay time: ", blockDelayTime);

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
      //console.log("msg Delay: ", msg.Delay);
    })

    this.addQueue(msgArray);

  }
}

export default Terminal;
