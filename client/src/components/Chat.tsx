import React from "react";
import * as constants from "../lib/constants";
import PubSub from "../lib/pubsub";
import TextField from '@material-ui/core/TextField';
import KeyboardArrowRightRoundedIcon from '@material-ui/icons/KeyboardArrowRightRounded';

interface TstreamUser {
  name: string,
    color: string,
}

interface Props {
  msgManager: PubSub;
  className?: string;
}

interface ChatMsg {
  Name: string;
  Content: string;
  Color: string;
  Time: string;
}

interface State {
  msgList: ChatMsg[];
  inputContent: string;
  name: string;
  color: string;
  isWaitingUsername: boolean,
  tempMsg: string,
}

const ChatSection: React.FC<ChatMsg> = ({ Name, Content, Color, Time}) => {
  return (
    <>
      <div className="w-full flex p-2 hover:bg-gray-900 rounded-lg">
        <div className="break-all">
          {
            Name === '' ? 
            <div className="font-bold">
              {
                Content === "Invalid Username" ? 
                  <img src="./warning.png" alt="warning" height="30" width="30" className="inline-block m-2"/> :
                  <img src="./hand-wave.png" alt="hand-wave" />
              }
                  {Content}
            </div> : 
            <>
              <span style={{color: Color}} className="font-black">{Name}</span>
              <span className="text-green-600 py-1"><KeyboardArrowRightRoundedIcon /></span>
              {Content}
            </>
          }
        </div>
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
      inputContent: "",
    })
  }

  componentDidMount() {    
    this.props.msgManager?.sub(constants.MSG_TCHAT, (cacheChat: Array<ChatMsg>) => {
      if (cacheChat === null) {
        return;
      }
      let newMsgList = this.state.msgList as ChatMsg[];
      for (let i = 0; i < cacheChat.length; i++) {
        newMsgList.push(cacheChat[i]);
      }
      this.setState({
        msgList: newMsgList,
      });
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
    document.getElementById("textarea")!.addEventListener('keydown', (e) => {
      var code = e.keyCode || e.which;
      if (code === 13) {
        e.preventDefault();
        this.onSendMsg(this.state.inputContent);
      }
    });
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
        notification = "Please enter your username (I.e: elonmusk)"; 
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

          localStorage.setItem('tstreamUser', JSON.stringify(tstreamUser));

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
          Time: new Date().toISOString(),
        };

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
      Time: new Date().toISOString(),
    };

    this.addNewMsg(data);
    this.props.msgManager?.pub(constants.MSG_TREQUEST_CHAT, data);
  }

  render() {
    return (
      <div 
        id="chat-wrapper"
        className={`flex flex-col w-full h-full flex flex-col border-l border-gray-500 relative pt-12 ${this.props.className}`} 
        style={{width: '400px', fontFamily: "'Ubuntu Mono', monospace"}}
      >
        <div style={{height: '0px'}} className="bg-black overflow-y-auto overflow-x-none p-2 flex flex-col-reverse flex-grow" id="chatbox">
          {
            this.state.msgList.slice(0).reverse().map(
              (item, index) => <ChatSection Name={item.Name} Content={item.Content} Color={item.Color} Time={item.Time} key={index}/>
            )
          }
        </div>
        <div className="bottom-0 transform w-full" id="textarea">
          <div 
            className="border-b border-gray-500 flex-shrink-0 flex items-center justify-between"
          >
             <TextField
              InputProps={{
                style: {
                  flexGrow: 1, 
                  borderRadius: ".5rem",
                  backgroundColor: "rgba(75,85,99,1)",
                  fontFamily: "'Ubuntu Mono', monospace",
               }
              }}
              placeholder={(this.state.isWaitingUsername) ? "Please enter your name..." : "Chat with everyone..."}
              fullWidth
              multiline
              value={this.state.inputContent}
              onChange={(e) => {
                this.setState({
                  inputContent: e.target.value,
                });
              }}
            />
          </div>
        </div>
      </div>
    )
  }
}

export default Chat;
