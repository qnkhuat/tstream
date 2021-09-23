import React, { useRef, useState, useEffect, useReducer } from 'react';
import pako from "pako";

import Controls from "./Controls";
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

interface PlayerState {
  playing: boolean;
  duration: number;
  currentTime: number;
}

export enum PlayerActionType {
  Play,
    Pause,
    Change,
    SetDuration,
}

type PlayerAction = 
  | { type: PlayerActionType.Play }
  | { type: PlayerActionType.Pause }
  | { type: PlayerActionType.SetDuration, payload: { duration: number } }
  | { type: PlayerActionType.Change, payload: { currentTime: number } }
;

const playerReducer = (state: PlayerState, action: PlayerAction) => {
  switch(action.type) {
    case PlayerActionType.Play:
      return {
        ...state,
        playing: true,
      }
    case PlayerActionType.Pause:
      return {
        ...state,
        playing: false,
      }
    case PlayerActionType.SetDuration:
    case PlayerActionType.Change:
      return {
        ...state,
        ...action.payload,
      }
  }
}

const Player: React.FC<Props> = ({id, width, height}: Props) => {
  const termRef = useRef<Xterm>(null);
  const [ writeManager, setWriteManager ] = useState<WriteManager>();
  const [ playerState, dispatch] = useReducer(playerReducer, {
    playing: false,
    currentTime: 0,
    duration: 0,
  });

  useEffect(() => {
    if (!termRef.current) return;

    let writeManager: WriteManager;
    api.getRecordManifest(id.toString()).then((manifest: message.Manifest) => {
      writeManager = new WriteManager(termRef.current!, manifest, { 
        playing: true, 
        refreshInterval: 200,
        onPlay: () => dispatch({ type: PlayerActionType.Play }),
        onPause: () => dispatch({ type: PlayerActionType.Pause }),
        onChange: (currentTime) => dispatch({ type: PlayerActionType.Change, payload: { currentTime } }),
      });

      const recordDuration = (new Date(manifest.StopTime)).getTime() - (new Date(manifest.StartTime)).getTime();
      dispatch({ type: PlayerActionType.SetDuration, payload: { duration: recordDuration }});
      setWriteManager(writeManager);
    });


    return () => {
      if(writeManager) writeManager.detach();
    }
  }, [termRef]);

  return <div className="relative">
    <Controls 
      className="absolute bottom-0 left-0 w-full z-50"
      onPlay={() => writeManager?.play()}
      onPause={() => writeManager?.pause()}
      onChange={(value: number) => {writeManager?.jumpTo(value)}}
      playing={playerState.playing}
      duration={playerState.duration}
      currentTime={playerState.currentTime}
    />
    <Terminal 
      ref={termRef}
      width={width}
      height={height} 
    />
  </div>
}

interface WriteMangerOptions {
  delay?: number;
  playing?: boolean;
  refreshInterval?: number;
  bufferSize?: number;
  onPlay?: () => void;
  onPause?: () => void;
  onChange?: (currentTime: number) => void;
}


class WriteManager {

  termRef: Xterm | null;
  manifest: message.Manifest | undefined;
  queue: message.Wrapper[] = [];
  // the size of buffer to pre-load content. it should diviable for the segmenttime in manifest
  bufferSize: number; // in milliseconds
  bufferStopTime: number = 0;
  startTime: number;
  stopTime: number;
  currentTime: number = 0;
  playing: boolean = false;
  refreshInterval: number;
  clearConsumeInterval?: () => void | undefined;
  fetching: boolean = false;
  delay: number; // in milliseconds
  onPlay?: () => void;
  onPause?: () => void;
  onChange?: (currentTime: number) => void;

  constructor(termRef: Xterm, manifest: message.Manifest, 
    { 
      onPlay,
      onPause,
      onChange,
      delay = 0, 
      playing = false, 
      refreshInterval = 200,
      bufferSize = 1000 * 120,
    }: WriteMangerOptions) {

    this.onPlay = onPlay;
    this.onPause = onPause;
    this.onChange = onChange;

    this.refreshInterval = refreshInterval;
    this.termRef = termRef;
    this.delay = delay;
    this.bufferSize = bufferSize;

    // load the manifest
    this.manifest = manifest;
    this.startTime = (new Date(manifest.StartTime)).getTime();
    this.stopTime = (new Date(manifest.StopTime)).getTime()

    if( playing ) this.play();
  }

  flushQueue() {
    this.queue = [];
  }

  // add queue ensure the queue is always ordered in delay time
  // q: the incoming queue is ensured to be incremented in Delay time
  addQueue(q: message.Wrapper[]) {
    if( q.length == 0 ) return;

    if (this.queue.length == 0) {
      this.queue = q
      return
    }
    const startQ = q[0].Delay, 
      endQ = q[q.length - 1].Delay;

    // all msg in incoming queue is pushed to the back
    if (startQ > this.queue[this.queue.length - 1].Delay) {
      this.queue.push(...q);
      return
    }

    // all msg in incoming queue is at the front
    if (endQ < this.queue[0].Delay) {
      this.queue.unshift(...q);
      return;
    }

    // msg in coming queue is at the middle of current queue

    let i = 0;
    // find the index of msg inside queue where we could insert the incoming queue to
    // I.e this.queue = [1,2,3,6,7,8]
    // let q = [4,5]
    // What we want to di is insert q at the middle of the queue
    while ( i < this.queue.length && startQ >= this.queue[i].Delay) {
      i += 1;
    }

    // incase this.queue = [1,2,3,5,6,7,8]
    // q = [4, 5]
    // we will remove the rest and append q to the end to get
    // this.queue  = [1,2,3,4,5]
    if (i < this.queue.length && this.queue[i].Delay >= endQ) {
      this.queue = this.queue.slice(0, i);
      this.queue.push(...q);
      return;
    } else {
      // incase this.queue = [1,2,3,6,7,8]
      // q = [4, 5]
      // we insert the q into the middle
      // this.queue  = [1,2,3,4,5,6,7,8]
      this.queue = this.queue.splice(i, 0, ...q);
      return;
    }
  }

  detach() {
    this.termRef = null;
    this.pause();
  }

  fetchSegmentByIndex = (index:number, stopIndex: number) => {
    if (!this.fetching || !this.manifest || index > stopIndex || index > this.manifest.Segments.length - 1) {
      this.fetching = false;
      return;
    }

    api.getRecordSegment(this.manifest.Id, this.manifest.Segments[index].Path).
      then((data) => {
        const msgArray = JSON.parse(pako.ungzip(data, {to : "string"}));
        if (msgArray) msgArray.forEach((msg: message.Wrapper) => this.addBlock(msg.Data)) ;
        this.bufferStopTime += this.manifest!.SegmentDuration;
      }).then(() => this.fetchSegmentByIndex(index + 1, stopIndex));
  }

  plan() {
    if(!this.manifest || this.fetching) return;
    const expectedEndBufferTime = this.currentTime + this.bufferSize;
    if (expectedEndBufferTime < this.bufferStopTime) return;

    this.fetching = true;
    let currentSegmentIndex = 0;

    while (currentSegmentIndex + 1 <= this.manifest.Segments.length - 1
      && this.manifest.Segments[currentSegmentIndex+1].Offset < this.bufferStopTime) {
      currentSegmentIndex += 1;
    }
    //while (currentSegmentIndex < this.manifest.Segments.length - 1
    //  && this.manifest.Segments[currentSegmentIndex].Offset < this.bufferStopTime) {
    //  currentSegmentIndex += 1;
    //}


    // try to get the 
    currentSegmentIndex = Math.max(0, currentSegmentIndex -1, currentSegmentIndex -2);
    if (currentSegmentIndex > 1 ) currentSegmentIndex -=1;

    const numberOfSegmentsToFetch = Math.round((expectedEndBufferTime - this.bufferStopTime) / this.manifest.SegmentDuration);
    console.log('submit to fetch', currentSegmentIndex, currentSegmentIndex + numberOfSegmentsToFetch);
    this.fetchSegmentByIndex(currentSegmentIndex, currentSegmentIndex + numberOfSegmentsToFetch);
  }

  play() {
    if (!this.clearConsumeInterval) {
      if(this.onPlay) this.onPlay();
      this.clearConsumeInterval = accurateInterval(() => {
        this.playing = true;
        this.consume();
      }, this.refreshInterval, {immediate: true});
    }
  }


  pause() {
    if(this.clearConsumeInterval) {
      this.clearConsumeInterval()
      this.clearConsumeInterval = undefined;
    };
    if(this.playing && this.onPause) this.onPause();
    this.playing = false;
    this.fetching = false;
  }

  jumpTo(to: number){
    this.pause();
    this.flushQueue();
    this.currentTime = to;
    this.bufferStopTime= to;
    this.play();
  };

  printStatus() {
    console.log("------------");
    console.log("Startime: ", this.startTime);
    console.log("CurrentTime: ", this.currentTime);
    console.log("Fetching: ", this.fetching);
    console.log("Queue len: ", this.queue.length);
    console.log("First queue: ", this.queue[0]?.Delay);
    console.log("Last queue: ", this.queue[this.queue.length - 1]?.Delay);
    console.log("Buffer stop time: ", this.bufferStopTime);
  }

  consume() {
    this.plan();
    this.printStatus()

    if(!this.playing || !this.termRef) return;
    if(this.onChange) this.onChange(this.currentTime);

    const currentTime = this.currentTime;
    const endTime = currentTime + this.refreshInterval;

    while (this.queue.length > 0 && this.queue[0].Delay < endTime) {
      if (this.queue[0].Delay > endTime) break;

      const msg: message.Wrapper = this.queue.shift()!;
      const msgTimeout = msg.Delay - currentTime;

      switch (msg.Type) {

        case constants.MSG_TWRITE:
          let bufferData = buffer.str2ab(msg.Data);
          setTimeout(() => this.termRef?.writeUtf8(bufferData), msgTimeout);
          break;

        case constants.MSG_TWINSIZE:
            setTimeout(() => this.termRef?.resize(msg.Data.Cols, msg.Data.Rows), msgTimeout);
          break;

        default:
            console.error("Unhandled message type: ", msg.Type);
      }
    }

    this.currentTime = this.currentTime + this.refreshInterval;
    return;
  }

  addBlock(block: message.TermWriteBlock) {
    
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


export default Player;
