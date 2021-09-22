import React, { useState, useEffect, useRef } from "react";
import pako from "pako";

import Terminal from "./Terminal";
import Xterm from "./Xterm";

import PubSub from "../lib/pubsub";
import * as constants from "../lib/constants";
import * as buffer from "../lib/buffer";
import * as message from "../types/message";

interface Props {
  msgManager: PubSub;
  width: number; // in pixel
  height: number; // in pixel
  delay?: number;
  className?: string;
}

const PubSubTerminal: React.FC<Props> = ({ msgManager, width = -1, height = -1, delay = 0, className = "" }: Props) => {
  const termRef = useRef<Xterm>(null);
  const [termSize, setTermSize] = useState<message.TermSize>({ Rows: 0, Cols: 0 });

  // handle message to from msgmanager
  useEffect(() => {
    if( termRef.current ) {
      const writeManager = new WriteManager(termRef.current, { delay });

      msgManager.pub("request", constants.MSG_TREQUEST_CACHE_CONTENT);
      msgManager.pub("request", constants.MSG_TREQUEST_WINSIZE);

      msgManager.sub(constants.MSG_TWRITEBLOCK, (block: message.TermWriteBlock) => {
        writeManager!.addBlock(block);
      });

      msgManager.sub(constants.MSG_TWINSIZE, (ws: message.TermSize) => {
        termRef.current?.resize(ws.Cols, ws.Rows);
        setTermSize(ws);
      })

      return () => {
        writeManager.detach();
        msgManager.unsub(constants.MSG_TWRITEBLOCK);
        msgManager.unsub(constants.MSG_TWINSIZE);
      }
    }

  }, [termRef, msgManager]);

  return (
    <Terminal
      delay={1500}
      className={className}
      width={width}
      height={height}
      rows={termSize.Rows}
      cols={termSize.Cols}
      ref={termRef}
    />
  )
}

interface WriteMangerOptions {
  delay?: number;
}

class WriteManager {

  termRef: Xterm | null;
  queue: message.Wrapper[] = [];
  delay: number; // in milliseconds

  constructor(termRef: Xterm, 
    { delay = 0 }: WriteMangerOptions) {
    this.termRef = termRef;
    this.delay = delay;
  }

  flushQueue() {
    this.queue = [];
  }

  detach() {
    this.termRef = null;
  }

  consume() {
    if (this.queue.length == 0 || ! this.termRef) return
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
              // write the previous buffers 
              if ( bufferArray.length > 0) this.termRef.writeUtf8(buffer.concatab(bufferArray));
              // reset it
              bufferArray = [];
              // resize
              this.termRef.resize(msg.Data.cols, msg.Data.rows);
              break;

            default:
              console.error("Unhandled message type: ", msg.Type);
          }

          this.queue.shift();
        } else break;
      }
      if ( bufferArray.length > 0) this.termRef.writeUtf8(buffer.concatab(bufferArray));

      // schedule to render upcomming messages
      // TODO: are there any ways we don't have to create many settimeout liek this?
      // tried sequentially call to settimeout but the performance is much worse
      this.queue.forEach((msg) => {
        switch (msg.Type) {

          case constants.MSG_TWRITE:
            let bufferData = buffer.str2ab(msg.Data);
            setTimeout(() => this.termRef!.writeUtf8(bufferData), msg.Delay);
            break;

          case constants.MSG_TWINSIZE:
            setTimeout(() => this.termRef!.resize(msg.Data.Cols, msg.Data.Rows), msg.Delay);
            break;

          default:
            console.error("Unhandled message type: ", msg.Type);
        }

      })
      this.flushQueue();
    }
  }

  addBlock(block: message.TermWriteBlock) {
    // when viewers receive this block
    // it only contains the actual start-time
    // we need to be able to re-compute the render time based on 
    // - when does this block being created
    // - now time
    // - the delay factor. In case of play back the delay = now - stream sesion start time
    const blockDelayTime =  (new Date(block.StartTime)).getTime() - (new Date()).getTime() + (this.delay - block.Duration);

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

    this.queue.push(...msgArray); // Concatnate
    this.consume();

  }
}


export default PubSubTerminal;
