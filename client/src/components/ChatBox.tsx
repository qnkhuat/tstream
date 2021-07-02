import React, { useEffect, useState } from 'react';

function base64ToArrayBuffer(input:string) {

  var binary_string =  window.atob(input);
  var len = binary_string.length;
  var bytes = new Uint8Array( len );
  for (var i = 0; i < len; i++)        {
    bytes[i] = binary_string.charCodeAt(i);
  }
  return bytes;
}

function Chat(props: any) { 
  const [ inputUsername, setInputUsername ] = useState<String>('');
  const [ inputValue, setInputValue ] = useState<String>('');
  const [ ws, setWs ] = useState<WebSocket | null>(null);
  const [ messengeData, setMessengeData ] = useState<Array<any>>([]);

  useEffect(() => {
    const conn = new WebSocket(props.wsUrl as string);
    conn.onmessage = (ev: MessageEvent) => {
      let msg = JSON.parse(ev.data);
      //process the messenge received
      if (msg.Type === "Chat") {
        let messenge = JSON.parse(window.atob(msg.Data))
        var tempMessengeData = messengeData as any;
        tempMessengeData.push({
          "name": messenge.name,
          "content": messenge.content,
          "time": messenge.time,
        });
        setMessengeData(tempMessengeData);
      }
    }
    setWs(conn);
  }, [])

  function chatSection() {
    var data = messengeData as any;
    let jsxMessenges: Array<any> = [];
    for (var i = data.length - 1; i >= 0; i--) {
        jsxMessenges.push(
          <div className="chat" key={i}>
            <p className="username">{data[i].name}</p>
            <p>{data[i].time}</p> 
            <p className="chat-content">{data[i].content}</p>
          </div>
        );
    }
    return <div id="chat-section">{jsxMessenges}</div>
  }

  return (
    <>
      <div style={{display: 'flex', justifyContent: 'center'}}>
        <input onChange={e => setInputUsername(e.target.value)} placeholder="name" style={{borderWidth: '2px'}}></input>
        <input onChange={e => setInputValue(e.target.value)} placeholder="content" style={{borderWidth: '2px'}}></input>
        <button style={{borderWidth: '2px', fontWeight: 'bold'}} onClick={e => {
          console.log(inputValue);
          var name = inputUsername;
          var content = inputValue;
          var time = new Date().toISOString();
          var payload = JSON.stringify({
            "name": name,
            "content": content,
            "time": time, 
          }); 
          if (ws !== null) {
            console.log("Messenge Sent");
            var tempMessengeData = messengeData as any;
            tempMessengeData.push({
              "name": name,
              "content": content,
              "time": time,
            });
            setMessengeData(tempMessengeData);
            ws.send(payload);
          }
        }}>Send message</button>
      </div>
     {chatSection()}
    </>
  );
}

export default Chat;
