import React, { useState, useEffect, useRef } from 'react';
import { RouteComponentProps } from "react-router-dom";
import pako from "pako";

import Terminal, { WriteManager } from "../../components/Terminal";
import Xterm from "../../components/Xterm";
import * as message from "../../types/message";
import * as api from "../../api";

interface Params {
  Id: string;
  roomKey?: string;
}

interface Props extends RouteComponentProps<Params> {}

const Records: React.FC<Props> = (props: Props) => {
  const termRef = useRef<Xterm>(null);

  useEffect(() => {
    const writeCB = (bufferData: Uint8Array) => {
      termRef.current?.writeUtf8(bufferData);
    };

    let winsizeCB = (ws: message.TermSize) => {
      termRef.current?.resize(ws.Cols, ws.Rows);
    }

    const writeManager = new WriteManager(writeCB, winsizeCB, 0);

    api.getRecordManifest(props.match.params.Id).then((data) => {
      console.log(data);
    });
    api.getRecordSegment(props.match.params.Id, "3.gz").then((data) => {
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
      width={800}
      height={400}
    />
  </>
}

export default Records;
