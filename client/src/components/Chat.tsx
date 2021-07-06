import React, { useRef, useEffect, useState, useCallback } from "react";
import * as constants from "../lib/constants";
import PubSub from "../lib/pubsub";
import * as base64 from "../lib/base64";

interface TstreamUser {
  name: string,
  color: string,
}

interface Props {
  msgManager: PubSub;
}

interface ChatMsg {
  Name: string;
  Content: string;
  Color: string;
}

interface ChatInfo {
  Msg: ChatMsg;
  isMe: boolean;
}

interface State {
  msgList: ChatInfo[];
  inputContent: string;
  name: string;
  color: string;
  isWaitingUsername: boolean,
}

const ChatSection: React.FC<ChatInfo> = ({ Msg, isMe }) => {

  return (
    <>
      <div className={`${isMe ? 'justify-end ml-auto mr-0' : ''} w-3/4 flex p-2`}>
        {!isMe && <span style={{color: Msg.Color}}>{Msg.Name}: </span>}
        <div style={{wordWrap: 'break-word', overflow: 'hidden'}}>{Msg.Content}</div>
      </div>
    </>
  )
}

class Chat extends React.Component<Props, State> {

  constructor(props: Props) {
    super(props);
    this.state = {
      msgList: [],
      inputContent: '',
      name: '',
      color: '',
      isWaitingUsername: false,
    }
  }
  
  addNewMsg(chatInfo: ChatInfo) {
    let newMsgList = this.state.msgList as ChatInfo[];
    newMsgList.push(chatInfo);
    this.setState({
      msgList: newMsgList,
    })
  }

  componentDidMount() {
    this.props.msgManager?.sub(constants.MSG_TCHAT, (chatMsg: ChatMsg) => {
      var chatInfo : ChatInfo = {
        Msg: chatMsg,
        isMe: false,
      }
      this.addNewMsg(chatInfo);
    });
    
    const payload = localStorage.getItem('tstreamUser');
    if (payload !== null) {
      const tstreamUser : TstreamUser = JSON.parse(payload);
      this.setState({
        name: tstreamUser.name,
        color: tstreamUser.color,
      });
    } 
  }

  onSendMsg(content: string, clearInput: boolean) {
    let tempMsg = content.trim();
    if (this.state.name === '' || this.state.color === '') {
      let notification: string = '';
      if (!this.state.isWaitingUsername) {
        notification = "It seems that you're a new user here. Please enter your name in the input section below. Remember your name must not be empty and contain not more than 10 characters."; 
        
      }
      else {
        if (tempMsg === '' || tempMsg.length > 10) {
          notification = 'Invalid Username';
        }
        else {
          var color: string = constants.COLOR_LIST[Math.floor(Math.random() * (constants.COLOR_LIST.length))];
          this.setState({
            name: tempMsg,
            color: color,
          });

          let tstreamUser : TstreamUser = {
            name: tempMsg,
            color: color,
          }
          const payload = localStorage.setItem('tstreamUser', JSON.stringify(tstreamUser));
          notification = "Welcome " + tempMsg + " to the TStream !!!!!. Happy a nice day (=^･^=) ...."
        }
      }

      let data = {
        Name: '', 
        Content: notification,
        Color: '', 
      };

      this.setState({
        inputContent: "",
      });
      
      var chatInfo : ChatInfo = {
        Msg: data,
        isMe: true,
      }
      this.addNewMsg(chatInfo);
      this.setState({
        isWaitingUsername: true,
      })
      return ;
    }

    if (tempMsg === '') {
      return;
    }
    let data = {
      Name: this.state.name,
      Content: tempMsg,
      Color: this.state.color,
    };
    if (clearInput) {
      this.setState({
        inputContent: "",
      });
    }
    var chatInfo : ChatInfo = {
      Msg: data,
      isMe: true,
    }
    this.addNewMsg(chatInfo);
    this.props.msgManager?.pub(constants.MSG_TREQUEST_CHAT, data);
  }

  render() {
    return (
      <div className="w-full flex flex-col border-l border-gray-500 relative" style={{width: "400px"}}>
        <div className="bg-black overflow-y-auto overflow-x-none p-2 flex flex-col-reverse" style={{height: "calc(100vh - 10rem - 57px)"}}>
          {this.state.msgList.slice(0).reverse().map((item, index) => <ChatSection Msg={item.Msg} isMe={item.isMe} key={index}/>)}
        </div>
        <div className="absolute bottom-0 transform w-full">
          <div className="h-20 border-b border-gray-500 flex-shrink-0 flex items-center justify-between pr-2">
            <input
              className="text-white px-3 py-3 flex-grow mr-2"
              placeholder={"Chat with everyone..."}
              style={{backgroundColor: '#121212'}}
              value={this.state.inputContent}
              onChange={(e) => {
                this.setState({
                  inputContent: e.target.value,
                });
              }}
              onKeyPress={(e) => {
                var code = e.keyCode || e.which;
                if (code === 13) {
                  this.onSendMsg(this.state.inputContent, true);
                }
              }}
            />
            {/* <button className="text-3xl transform hover:scale-125 duration-100" onClick={() => this.onSendMsg('&#128540;', false)}>&#128540;</button> */}
          </div>
          <div className="h-20 flex-shrink-0 flex flex-row-reverse items-center justify-between px-5 py-3">
            {/* <div className="flex-grow">
              <button className="text-red-600 text-4xl transform hover:scale-125 duration-100">&#9829;</button>
            </div> */}
            <button
              className="px-10 py-2 bg-red-600 text-white rounded flex-shrink-0"
              onClick = {() => {
                this.onSendMsg(this.state.inputContent, true)
              }}
            >
              Send
            </button>
          </div>
        </div>
      </div>
    )
  }
}

export default Chat;
