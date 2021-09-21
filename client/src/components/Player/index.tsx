import React, { useRef, useEffect } from 'react';
import pako from "pako";

import Terminal, { WriteManager } from "../Terminal";
import XTerm from "../Xterm";

import * as api from "../../api";
import * as message from "../../types/message";

interface Props {
  id: string;
  width: number;
  height: number;
}

const Player: React.FC<Props> = ({id, width, height}: Props) => {
  const termRef = useRef<XTerm>(null);

  useEffect(() => {
    if (!termRef.current) return;
    const writeManager = new WriteManager(termRef.current, {playing: true, refreshInterval: 2000});

    api.getRecordSegment(id.toString(), "3.gz").then((data) => {
      const msgArray = JSON.parse(pako.ungzip(data, {to : "string"}));
      msgArray.forEach((msg: message.TermWriteBlock) => writeManager.addBlock(JSON.parse(window.atob(msg.Data))) );
    }).catch(console.error);

    return () => {
      writeManager.detach();
    }
  }, [termRef.current]);

  return <Terminal 
    ref={termRef}
    width={width}
    height={height} 
  />
}

export default Player;
