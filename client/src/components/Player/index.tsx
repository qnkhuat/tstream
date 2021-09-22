import React, { useRef, useEffect } from 'react';
import pako from "pako";

import Terminal from "../Terminal";
import Xterm from "../Xterm";

import * as api from "../../api";
import { accurateInterval } from "../../utils";
import * as constants from "../../lib/constants";
import * as buffer from "../../lib/buffer";
import * as message from "../../types/message";


interface Props {
  id: string;
  width: number;
  height: number;
}

const Player: React.FC<Props> = ({id, width, height}: Props) => {
  const termRef = useRef<Xterm>(null);

  useEffect(() => {
    if (!termRef.current) return;
    const writeManager = new WriteManager(termRef.current, {playing: true, refreshInterval: 2000, mode: "streaming"});
    //api.getRecordManifest(id.toString()).then(console.log).catch(console.error);

    api.getRecordSegment(id.toString(), "3.gz").then((data) => {
      const msgArray = JSON.parse(pako.ungzip(data, {to : "string"}));
      msgArray.forEach((msg: message.TermWriteBlock) => writeManager.addBlock(JSON.parse(window.atob(msg.Data))) );
    }).catch(console.error);

    return () => {
      writeManager.detach();
    }
  }, [termRef]);

  return <Terminal 
    ref={termRef}
    width={width}
    height={height} 
  />
}

interface WriteMangerOptions {
  delay?: number;
  playing?: boolean;
  refreshInterval?: number;
  mode?: "streaming" | "playback";
}


class WriteManager {

  termRef: Xterm | null;
  queue: message.Wrapper[] = [];
  delay: number; // in milliseconds
  startTime: number | null = null;
  currentTime: number | null = null;
  playing: boolean = false;
  refreshInterval: number;
  clearConsumeInterval?: () => void | undefined;
  mode: "streaming" | "playback";

  constructor(termRef: Xterm, 
    { delay = 0, 
      playing = false, 
      refreshInterval = 200,
      mode = "streaming",
    }: WriteMangerOptions) {
    this.mode = mode;
    this.refreshInterval = refreshInterval;
    this.termRef = termRef;
    this.delay = delay;
    if( playing ) this.play();
  }

  flushQueue() {
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

  printStatus() {
    console.log("------------");
    console.log("Startime: ", this.startTime);
    console.log("CurrentTime: ", this.currentTime);
    console.log("Queue len: ", this.queue.length);
    console.log("First queue: ", this.queue[0]?.Delay);
  }

  consume() {
    this.printStatus()
        
    if(!this.playing || !this.termRef) return;

    const returnCallback = () => {
      if(this.currentTime) this.currentTime = this.currentTime + this.refreshInterval;
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
    console.log("New block: ", (new Date(block.StartTime)).getTime());
    // the starttime of stream or records will be the the starttime of the first block received
    if (!this.startTime || !this.currentTime) {
      const blockStartTime = (new Date(block.StartTime)).getTime();
      this.startTime = blockStartTime;
      // the delay is how much the block message sent to server is delayed to its start time
      // it means that the delay will be include the block duration time and a bit of buffer
      // For example : delay = 1.5s and block duration 1s. we have .5 second for the delay of network
      this.currentTime = (new Date()).getTime() - blockStartTime - (this.delay! - block.Duration);
      console.log("set Current titme: ", this.currentTime);
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
      //console.log("msg Delay: ", msg.Delay);
    })

    this.addQueue(msgArray);

  }
}


export default Player;
