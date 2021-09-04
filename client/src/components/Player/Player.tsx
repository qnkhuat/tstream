import React, { useRef, useEffect, useReducer } from 'react';
import pako from "pako";
import Terminal, { WriteManager } from "../Terminal";
import Xterm from "../Xterm";
import * as api from "../../api";
import * as message from "../../types/message";

interface Props {
  id: number;
  width: number;
  height: number;
}

interface State {
  play: boolean;
  rate: number;
  currentTime: number; // in milliseconds
}

export enum PlayerAction { 
  SeekTo,
    Play,
    Pause,
    UpdateRate

}


//const reducer = (state: State, action: Aciton) => 

const Player: React.FC<Props> = ({id, width, height}: Props) => {
  const termRef = useRef<Xterm>(null);
  useEffect(() => {
    const writeCB = (bufferData: Uint8Array) => {
      termRef.current?.writeUtf8(bufferData);
    };

    let winsizeCB = (ws: message.TermSize) => {
      termRef.current?.resize(ws.Cols, ws.Rows);
    }

    const writeManager = new WriteManager(writeCB, winsizeCB, 0);

    api.getRecordManifest(id.toString()).then((data) => {
      console.log(data);
    });

    api.getRecordSegment(id.toString(), "3.gz").then((data) => {
      //const msgArray = JSON.parse(buffer.ab2str(data));
      const msgArray = JSON.parse(pako.ungzip(data, {to : "string"}));
      //console.log(msgArray)
      //const manual = JSON.parse(buffer.ab2str(pako.ungzip(data)));
      const alo = msgArray[0];
      let blockMsg: message.TermWriteBlock = JSON.parse(window.atob(alo.Data));
      writeManager.addBlock(blockMsg);
      //const msgArray = JSON.parse(pako.ungzip(data,{ to: 'string' })[0]);
    }).catch(console.error);
  }, []);


  return <>
    <Terminal 
      ref={termRef}
      width={width}
      height={height}
    />
  </>
}

export default Player;
