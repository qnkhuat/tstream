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

interface State {
  msgList: ChatMsg[];
  inputContent: string;
  name: string;
  color: string;
  isWaitingUsername: boolean,
  tempMsg: string,
}

const ChatSection: React.FC<ChatMsg> = ({ Name, Content, Color }) => {
  return (
    <>
      <div className="w-full flex p-2 hover:bg-gray-800 rounded-lg">
        {Name !== '' && <span style={{color: Color}} className="font-black">{Name}</span>}        
        {Name !== '' && <span className="text-green-600 pl-2 pr-2">{">"}</span>}
        <div style={{wordWrap: 'break-word', overflow: 'hidden'}}>{Content}</div>
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
      tempMsg: '',
    }
  }
  
  addNewMsg(chatMsg: ChatMsg) {
    let newMsgList = this.state.msgList as ChatMsg[];
    newMsgList.push(chatMsg);
    this.setState({
      msgList: newMsgList,
    })
  }

  componentDidMount() {
    this.props.msgManager?.sub(constants.MSG_TCHAT, (chatMsg: ChatMsg) => {
      this.addNewMsg(chatMsg);
    });
    
    const payload = localStorage.getItem('tstreamUser');
    if (payload !== null) {
      const tstreamUser : TstreamUser = JSON.parse(payload);
      this.setState({
        name: tstreamUser.name,
        color: tstreamUser.color,
      });
    } 
    
    // disable enter default behavior of textarea 
    document.getElementById("textarea").addEventListener('keydown', (e) => {
      var code = e.keyCode || e.which;
      if (code === 13) {
        e.preventDefault();
        let textarea = document.getElementById("textarea");
        textarea.rows = 1;
        this.onSendMsg(this.state.inputContent);
      }
    });
    document.getElementById("chatbox").style.height = `calc(100vh - ${document.getElementById("textarea").clientHeight}px - 57px)`;
  }

  onSendMsg(content: string) {
    let tempMsg : string = content.trim();
    let name : string = '';
    let color : string = '';

    // Don't find the user data in the browser
    if (this.state.name === '' || this.state.color === '') {
      let notification: string = '';

      // ask for first time
      if (!this.state.isWaitingUsername) {
        notification = "Please enter your username(I.e: elonmusk)"; 
        this.setState({
          tempMsg: tempMsg,
          isWaitingUsername: true,
        });
      } 
      else {
        // invalid username
        if (tempMsg === '' || tempMsg.length > 10) {
          notification = 'Invalid Username';
        }
        // valid username
        else {
          name = tempMsg;
          color = constants.COLOR_LIST[Math.floor(Math.random() * (constants.COLOR_LIST.length))];
          this.setState({
            name: name,
            color: color,
            isWaitingUsername: false,
          });

          let tstreamUser : TstreamUser = {
            name: tempMsg,
            color: color,
          }
          const payload = localStorage.setItem('tstreamUser', JSON.stringify(tstreamUser));
          tempMsg = this.state.tempMsg;

          // if the first message is empty, just ignore it
          if (tempMsg === "") {
            this.setState({
              inputContent: "",
            });
            return ;
          }
        }
      }

      // send notification
      if (notification !== '') {
        let data = {
          Name: '', 
          Content: notification,
          Color: '', 
        };

        this.setState({
          inputContent: "",
        });

        this.addNewMsg(data);
        return ;
      }
    }

    if (tempMsg === '') {
      return;
    }

    if (name === '') {
      name = this.state.name;
    }
    if (color === '') {
      color = this.state.color;
    }

    let data = {
      Name: name,
      Content: tempMsg,
      Color: color,
    };

    this.setState({
      inputContent: "",
    });

    this.addNewMsg(data);
    this.props.msgManager?.pub(constants.MSG_TREQUEST_CHAT, data);
  }

  render() {
    return (
      <div className="w-full flex flex-col border-l border-gray-500 relative" style={{width: "400px", fontFamily: "'Ubuntu Mono', monospace"}}>
        <div className="bg-black overflow-y-auto overflow-x-none p-2 flex flex-col-reverse" id="chatbox">
          {
            this.state.msgList.slice(0).reverse().map(
              (item, index) => <ChatSection Name={item.Name} Content={item.Content} Color={item.Color} key={index}/>
            )
          }
        </div>
        <div className="absolute bottom-0 transform w-full">
          <div className="border-b border-gray-500 flex-shrink-0 flex items-center justify-between">
            <textarea
              className="text-white px-3 py-3 flex-grow bg-gray-600 border-4 border-gray-500 focus:bg-black focus:border-purple-600 rounded-lg"
              placeholder={"Chat with everyone..."}
              value={this.state.inputContent}
              onChange={(e) => {
                this.setState({
                  inputContent: e.target.value,
                });
                let textarea = document.getElementById("textarea");
                textarea.rows = Math.floor(e.target.value.length / 45) + 1;
                document.getElementById("chatbox").style.height = `calc(100vh - ${document.getElementById("textarea").clientHeight}px - 57px)`;
              }}
              onKeyPress={(e) => {
              }}
              rows={1}
              id="textarea"
            />
          </div>
        </div>
      </div>
    )
  }
}

export default Chat;
