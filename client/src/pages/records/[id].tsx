import React, { useState, useEffect } from 'react';
import { RouteComponentProps } from "react-router-dom";
import * as api from "../../api";
import * as buffer from "../../lib/buffer";
import pako from "pako";

import Terminal from "../../components/Terminal";

interface Params {
  Id: string;
  roomKey?: string;
}

interface Props extends RouteComponentProps<Params> {}

const Records: React.FC<Props> = (props: Props) => {
  useEffect(() => {
    //api.getRecordManifest(props.match.params.Id).then(console.log);
    api.getRecordSegment(props.match.params.Id, "2.gz").then((data) => {
      const msgArray = JSON.parse(pako.ungzip(data, { to: "string" }));
      console.log(msgArray[0]);
      //const msgArray = JSON.parse(pako.ungzip(data,{ to: 'string' })[0]);
    });
  }, []);
  
  return <>
  </>
}

export default Records;
