import React, { useState, useRef, useEffect, useReducer } from 'react';
import pako from "pako";

import Terminal from "../Terminal";
import Xterm from "../Xterm";
import Controls from "./Controls";

import { playerActions } from "./store";

import * as api from "../../api";
import * as message from "../../types/message";
import * as store from "./store";

interface Props {
  id: string;
  width: number;
  height: number;
}

const Player: React.FC<Props> = ({id, width, height}: Props) => {
  const [ state, dispatch ] = useReducer(store.playerReducer, store.initialState);

  useEffect(() => {
    
    api.getRecordManifest(id.toString()).
      then((manifest) => dispatch(playerActions.setManifest(manifest))).
      catch(err => console.error(err));

    api.getRecordSegment(id.toString(), "3.gz").then((data) => {
      const msgArray = JSON.parse(pako.ungzip(data, {to : "string"}));
      const alo = msgArray[0];
      let blockMsg: message.TermWriteBlock = JSON.parse(window.atob(alo.Data));
      //tempWriteManager.addBlock(blockMsg);
    }).catch(console.error);

    //setWriteManager(tempWriteManager);
    //tempWriteManager.play();
  }, []);


  return <div className="relative">
    <Terminal 
      //ref={termRef}
      width={width}
      height={height} />
  </div>
}

export default Player;
