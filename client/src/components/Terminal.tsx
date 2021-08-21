import React, { useState, useEffect, useRef } from "react";
import Xterm from "./Xterm";
import PubSub from "../lib/pubsub";
import * as constants from "../lib/constants";
import * as message from "../types/message";
import * as buffer from "../lib/buffer";
import * as pako from "pako";


interface Winsize {
  Rows: number;
  Cols: number;
}

// TODO: add handle % and px for size
interface Props {
  msgManager: PubSub;
  width: number; // in pixel
  height: number; // in pixel
  delay?: number;
  className?: string;
}


const Terminal: React.FC<Props> = ({ msgManager, width = -1, height = -1, delay = 0, className = ""}: Props) => {
  const termRef = useRef<Xterm>(null);
  const divRef = useRef<HTMLDivElement>(null);

  const rescale = () => {
    if (termRef.current && divRef.current && (width! > 0 || height! > 0)) {
      const core = (termRef.current?.terminal as any)._core,
        cellWidth = core._renderService.dimensions.actualCellWidth,
        cellHeight = core._renderService.dimensions.actualCellHeight,
        currentFontSize = termRef.current.terminal.getOption('fontSize'),
        rows = termRef.current.terminal.rows,
        cols = termRef.current.terminal.cols,
        hFontSizeMultiplier = height / (cellHeight * rows),
        wFontSizeMultiplier = width / (cellWidth * cols),
        // method doesn't ensure termianl will 100% fit the required size since fontsize are discrete
        // Another method is to transform scale to fit the window
        // But I haven't figured out why the scaled version sometimes make terminal deformed 
        // after multiple times of apply scale transformation
        newFontSize = Math.floor(
          hFontSizeMultiplier > wFontSizeMultiplier 
          ? currentFontSize * wFontSizeMultiplier 
          : currentFontSize * hFontSizeMultiplier);

      termRef.current.terminal.setOption('fontSize', newFontSize);
      termRef.current.terminal.refresh(0, rows-1); // force xterm to re-render everything
    }
  }

  // handle message to from msgmanager
  useEffect(() => {
    const writeCB = (bufferData: Uint8Array) => {
      termRef.current?.writeUtf8(bufferData);
    };

    let winsizeCB = (ws: Winsize) => {
      termRef.current?.resize(ws.Cols, ws.Rows);
      rescale();
    }

    const writeManager = new WriteManager(writeCB, winsizeCB, delay);

    msgManager.pub("request", constants.MSG_TREQUEST_CACHE_CONTENT);
    msgManager.pub("request", constants.MSG_TREQUEST_WINSIZE);

    msgManager.sub(constants.MSG_TWRITEBLOCK, (block: message.TermWriteBlock) => {
      writeManager!.addBlock(block);
    });

    msgManager.sub(constants.MSG_TWINSIZE, (ws: Winsize) => {
      termRef.current?.resize(ws.Cols, ws.Rows);
      rescale();
    })

    return () => {
      msgManager.unsub(constants.MSG_TWRITEBLOCK);
      msgManager.unsub(constants.MSG_TWINSIZE);
    }

  }, [msgManager]);

  // handle when resize is requried
  useEffect(() => {
    const handleResize = () => { 
      rescale(); 
    };

    window.addEventListener("resize", handleResize);

    // first refresh
    handleResize();
    return () => {
      window.removeEventListener("resize", handleResize);
    }
  }, [width, height]);


  return (
    <div className={`relative ${className} overflow-hidden`}
      style={{width: width!, height: height!}}>
      <div ref={divRef}
        className="divref absolute top-1/2 left-1/2 origin-top-left transform -translate-x-1/2 -translate-y-1/2 overflow-hidden">
        <Xterm
          options={{
            rightClickSelectsWord: false,
              disableStdin: true,
          }}
          ref={termRef}/>
      </div>
    </div>
  )
}

class WriteManager {

  queue: message.Wrapper[] = [];
  writeCB: (arr:Uint8Array) => void;
  winsizeCB: (ws:Winsize) => void
    delay: number; // in milliseconds

  constructor(writeCB: (arr: Uint8Array) => void, winsizeCB: (ws: Winsize) => void, delay: number = 0) {
    this.writeCB = writeCB;
    this.winsizeCB = winsizeCB;
    this.delay = delay;
  }

  resetQueue() {
    this.queue = [];
  }

  addQueue(q: message.Wrapper[]) {
    this.queue.push(...q); // Concatnate
    this.consume();
  }

  consume() {
    if (this.queue.length == 0) return
    else {

      // any message has offset < 0 => messages from the past with respect to render time
      // concat all these messages into one buffer and render at once
      let bufferArray: Uint8Array[] = [];
      while (true && this.queue.length != 0) {
        let msg = this.queue[0];

        if (msg.Delay < 0) {
          switch (msg.Type) {
            case constants.MSG_TWRITE:
              let bufferData = buffer.str2ab(msg.Data)
              bufferArray.push(bufferData);
              break;

            case constants.MSG_TWINSIZE:
              this.winsizeCB(msg.Data);
              break;

            default:
              console.error("Unhandled message type: ", msg.Type);
          }

          this.queue.shift();
        } else break;
      }
      if ( bufferArray.length > 0) this.writeCB(buffer.concatab(bufferArray));

      // schedule to render upcomming messages
      // TODO: are there any ways we don't have to create many settimeout liek this?
      // tried sequentially call to settimeout but the performance is much worse
      this.queue.forEach((msg) => {
        switch (msg.Type) {

          case constants.MSG_TWRITE:
            let bufferData = buffer.str2ab(msg.Data);
            setTimeout(() => this.writeCB(bufferData), msg.Delay);
            break;

          case constants.MSG_TWINSIZE:
            setTimeout(() => this.winsizeCB(msg.Data), msg.Delay);
            break;

          default:
            console.error("Unhandled message type: ", msg.Type);
        }

      })
      this.resetQueue();
    }
  }

  addBlock(block: message.TermWriteBlock) {
    // when viewers receive this block
    // it only contains the actual start-time
    // we need to be able to re-compute the render time based on 
    // - now time
    // - when does this block being created
    // - the delay factor. In case of play back the delay = now - stream sesion start time
    let blockDelayTime = (new Date()).getTime() - (new Date(block.StartTime)).getTime() - this.delay;

    // this is a big chunk of encoding/decoding
    // Since we have to : reduce message size by usign gzip and also
    // every single termwrite have to be decoded, or else the rendering will screw up
    // the whole block often took 9-20 milliseconds to decode a 3 seconds block of message
    let data = pako.ungzip(buffer.str2ab(block.Data));
    let msgArrayString: string[] = JSON.parse(buffer.ab2str(data));

    let msgArray: message.Wrapper[] = [];
    msgArrayString.forEach((msgString: string, i) => {
      // re-compute the offset of this message with respect to the render time
      let msg: message.Wrapper = JSON.parse(window.atob(msgString));
      msg.Delay = msg.Delay - blockDelayTime;
      msgArray.push(msg);
    })

    this.addQueue(msgArray);

  }
}


export default Terminal;
