import React, { useState, useEffect, useRef } from "react";
import Terminal from "./Terminal";
import PubSub from "../lib/pubsub";
import * as constants from "../lib/constants";
import * as message from "../types/message";

interface Props {
  msgManager: PubSub;
  width: number; // in pixel
  height: number; // in pixel
  delay?: number;
  className?: string;
}

const PubSubTerminal: React.FC<Props> = ({ msgManager, width = -1, height = -1, delay = 0, className = "" }: Props) => {
  const termRef = useRef<Terminal>(null);
  const [termSize, setTermSize] = useState<message.TermSize>({ Rows: 0, Cols: 0 });

  // handle message to from msgmanager
  useEffect(() => {
    
    msgManager.pub("request", constants.MSG_TREQUEST_CACHE_CONTENT);
    msgManager.pub("request", constants.MSG_TREQUEST_WINSIZE);

    msgManager.sub(constants.MSG_TWRITEBLOCK, (block: message.TermWriteBlock) => {
      termRef.current?.addBlock(block);
    });

    msgManager.sub(constants.MSG_TWINSIZE, (ws: message.TermSize) => {
      termRef.current?.resize(ws);
      setTermSize(ws);
    })

    return () => {
      msgManager.unsub(constants.MSG_TWRITEBLOCK);
      msgManager.unsub(constants.MSG_TWINSIZE);
    }

  }, [msgManager, termRef.current]);

  return (
    <Terminal
      delay={delay}
      className={className}
      width={width}
      height={height}
      rows={termSize.Rows}
      cols={termSize.Cols}
      ref={termRef}
    />
  )
}

export default PubSubTerminal;
