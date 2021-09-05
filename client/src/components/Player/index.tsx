import React, { useRef, useEffect, useReducer } from 'react';
import pako from "pako";
import Terminal, { WriteManager } from "../Terminal";
import Xterm from "../Xterm";
import * as api from "../../api";
import * as message from "../../types/message";
import * as store from "./store";
import Controls from "./Controls";

interface Props {
  id: string;
  width: number;
  height: number;
}

const Player: React.FC<Props> = ({id, width, height}: Props) => {
  const [ state, dispatch ] = useReducer(store.playerReducer, store.initialState);

  const termRef = useRef<Xterm>(null);

  useEffect(() => {
    const writeCB = (bufferData: Uint8Array) => {
      termRef.current?.writeUtf8(bufferData);
    };

    let winsizeCB = (ws: message.TermSize) => {
      termRef.current?.resize(ws.Cols, ws.Rows);
    }

    const writeManager = new WriteManager(writeCB, winsizeCB, 0);

    api.getRecordSegment(id.toString(), "3.gz").then((data) => {
      const msgArray = JSON.parse(pako.ungzip(data, {to : "string"}));
      const alo = msgArray[0];
      let blockMsg: message.TermWriteBlock = JSON.parse(window.atob(alo.Data));
      writeManager.addBlock(blockMsg);
    }).catch(console.error);

  }, []);

  return <>
    <div className="relative">
      <Terminal 
        ref={termRef}
        width={width}
        height={height}
      />
      <Controls 
        state={state}
        dispatch={dispatch}
        //handlePlay={() => dispatch({type: store.PlayerActionType.Play})}
        //handlePause={() => dispatch({type: store.PlayerActionType.Pause})}
        //handleJumpTo={(value) => dispatch({type: store.PlayerActionType.JumpTo, payload: {to: value}})}
        className="absolute bottom-0 w-full z-30 left-0"
      />
    </div>
  </>
}

export default Player;
