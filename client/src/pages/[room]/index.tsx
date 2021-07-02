import React, { useEffect, useState } from 'react';
import { Terminal } from 'xterm';
import { FitAddon } from 'xterm-addon-fit';
import 'xterm/css/xterm.css';
import { useParams } from "react-router-dom";

import WSTerminal from "../../components/WSTerminal";
import Chat from "../../components/ChatBox";

function base64ToArrayBuffer(input:string) {

  var binary_string =  window.atob(input);
  var len = binary_string.length;
  var bytes = new Uint8Array( len );
  for (var i = 0; i < len; i++)        {
    bytes[i] = binary_string.charCodeAt(i);
  }
  return bytes;
}


interface Params {
  username: string;
}

function Room() {
  const params: Params = useParams();
  const [ inputValue, setInputValue ] = useState("");
  const sessionID = params.username;
  const wsUrl = `ws://0.0.0.0:3000/ws/${sessionID}/viewer`;

  return (
    <div className="App">
      <WSTerminal wsUrl={wsUrl} width={400} height={300}/>
      <Chat wsUrl={wsUrl}/>
    </div>
  );

}

export default Room;
