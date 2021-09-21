import React, { useRef, useEffect } from 'react';
import pako from "pako";

import Terminal from "../Terminal";
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

    //api.getRecordManifest(id.toString()).
    //  then((manifest) => console.log(manifest)).
    //  catch(err => console.error(err));

    api.getRecordSegment(id.toString(), "3.gz").then((data) => {
      const msgArray = JSON.parse(pako.ungzip(data, {to : "string"}));
      const alo = msgArray[0];
      let blockMsg: message.TermWriteBlock = JSON.parse(window.atob(alo.Data));
      //termRef.current!.addBlock(blockMsg);
    }).catch(console.error);

  }, [termRef.current]);


  return <Terminal 
      ref={termRef}
      //controls={true}
      //mode="playback"
      width={width}
      height={height} 
    />
}

export default Player;
