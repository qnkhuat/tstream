import React, { useState, useEffect, useRef } from "react";
import Xterm from "./Xterm";
import Terminal, { WriteManager } from "./Terminal";
import PubSub from "../lib/pubsub";
import * as constants from "../lib/constants";
import * as message from "../types/message";

// TODO: add handle % and px for size
interface Props {
  msgManager: PubSub;
  width: number; // in pixel
  height: number; // in pixel
  delay?: number;
  className?: string;
}

const PubSubTerminal: React.FC<Props> = ({ msgManager, width = -1, height = -1, delay = 0, className = "" }: Props) => {
  const termRef = useRef<Xterm>(null);
  const [termSize, setTermSize] = useState<message.TermSize>({Rows: 0, Cols: 0});

  // handle message to from msgmanager
  useEffect(() => {
    const writeCB = (bufferData: Uint8Array) => {
      termRef.current?.writeUtf8(bufferData);
    };

    let winsizeCB = (ws: message.TermSize) => {
      termRef.current?.resize(ws.Cols, ws.Rows);
    }

    const writeManager = new WriteManager(writeCB, winsizeCB, delay);

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
      msgManager.unsub(constants.MSG_TWRITEBLOCK);
      msgManager.unsub(constants.MSG_TWINSIZE);
    }

  }, [msgManager]);

  return (
    <Terminal
      width={width}
      height={height}
      rows={termSize.Rows}
      cols={termSize.Cols}
      ref={termRef}
    />
  )
}

export default PubSubTerminal;
