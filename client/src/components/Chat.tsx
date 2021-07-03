import React, { useRef, useEffect, useState } from "react";
import * as constants from "../lib/constants";
import PubSub from "../lib/pubsub";
import * as base64 from "../lib/base64";

interface Props {
  msgManager: PubSub;
}

interface ChatMsg {
  Name: string;
  Content: string;
  Color: string;
}

interface ChatSec {
  Msg: ChatMsg;
  isMe: boolean;
}

const ChatSection: React.FC<ChatSec> = ({ Msg, isMe }) => {
  
  return (
    <>
      <div className={`${isMe ? 'text-right ml-auto mr-0' : ''} w-3/4 flex p-2`}>
        {!isMe && <div style={{color: Msg.Color}}>{Msg.Name}: </div>}
        <p>{Msg.Content}</p>
      </div>
      <div style={{clear: 'both'}}></div>
    </>
  )
}

const Chat: React.FC<Props> = ({ msgManager }) => {
  const [ msgList, setMsgList ] = useState<Array<ChatMsg>>([
    {
      Name: "manhcd",
      Content: "Oi gioi oi tao dep trai nhat trai dat nay aaaaaaaaaa aaaaaaaaaa aaaaaaaaaaaa aaaaaaaaaaaaa aaaaaaaaaaa aaaaaa",
      Color: "#FA8072",
    }, 
    {
      Name: "trangttt",
      Content: "Oi gioi oi tao xau trai nhat trai dat nay aaaaaaaaaa aaaaaaaaaa aaaaaaaaaaaa aaaaaaaaaaaaa aaaaaaaaaaa aaaaaa",
      Color: "#DC143C",
    },
  ]);
  const [ inputContent, setInputContent ] = useState<string>('');
  var tempList : ChatMsg[] = [];

  console.log("msgList is", msgList);
  useEffect(() => {
    msgManager?.sub(constants.MSG_TCHAT, (chatMsg: ChatMsg) => {
      console.log("msgList in sub is ", msgList);
      let newMsgList = msgList as Array<ChatMsg>;
      newMsgList.push(chatMsg);
      tempList.push(chatMsg);
      setMsgList(newMsgList);
    })
  }, []);
  
  function onSendMsg(content: string, clearInput: boolean) {
    let tempMsg = content.trim();
    if (tempMsg === '') {
      return;
    }     
    let data = {
      Name: "User",
      Content: tempMsg,
      Color: "manhcd",
    };     
    if (clearInput) setInputContent('');
    msgManager?.pub(constants.MSG_TREQUEST_CHAT, data);
  }

  console.log("before return: ", msgList.length);
  console.log("tempList is ", tempList);
  return (
    <div className="w-full flex flex-col border-l border-gray-500">
      <div className="h-full bg-black overflow-y-auto overflow-x-none p-2 flex flex-col-reverse">
        <p>{msgList.length}</p>
        {tempList.map((item, index) => <ChatSection Msg={item} isMe={(index % 2 == 0) ? true : false}/>)}
      </div>
      <div className="h-20 border-b border-gray-500 flex-shrink-0 flex items-center justify-between pr-2">
        <input className="text-white px-3 py-3 flex-grow" placeholder={"Chat with everyone..."} style={{backgroundColor: '#121212'}} value={inputContent} onChange={(e) => setInputContent(e.target.value)} /> 
        <button className="text-3xl transform hover:scale-125 duration-100" onClick={() => onSendMsg('&#128540;', false)}>&#128540;</button>
      </div>
      <div className="h-20 flex-shrink-0 flex items-center justify-between px-5 py-3">
        <div className="flex-grow">
          <button className="text-red-600 text-4xl transform hover:scale-125 duration-100">&#9829;</button>
        </div>
        <button 
          className="px-10 py-2 bg-red-600 text-white rounded flex-shrink-0"
          onClick = {() => onSendMsg(inputContent, true)}
        >
          Send
        </button>
      </div>
    </div>
  )
}

export default Chat;
