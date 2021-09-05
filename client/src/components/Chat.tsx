import React from "react";
import * as constants from "../lib/constants";
import * as message from "../types/message";
import PubSub from "../lib/pubsub";
import TextField from '@mui/material/TextField';
import KeyboardArrowRightRoundedIcon from '@mui/icons-material/KeyboardArrowRightRounded';

// key in local storage
const USER_CONFIG_KEY = "tstreamUser";

interface TstreamUser {
  name: string,
  color: string,
}

interface Props {
  msgManager: PubSub;
  height: number;
  width: number;
  className?: string;
}

interface State {
  msgList: message.ChatMsg[];
  inputContent: string;
  userConfig: TstreamUser | null;
  isWaitingUsername: boolean,
  tempMsg: string,
}

const ChatSection: React.FC<message.ChatMsg> = ({ Name, Content, Color, Time}) => {
  return (
    <>
      <div className="w-full flex p-2 hover:bg-gray-900 rounded-lg">
        <div className="break-all text-sm sm:text-base">
          {
            Name === '' ? 
              <div className="font-bold">
                <p>{Content}</p>
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

    let userConfig = this.getUserConfig()
    this.state = {
      msgList: [],
      inputContent: '',
      userConfig: userConfig,
      isWaitingUsername: false,
      tempMsg: '',
    }
  }

  addNewMsg(chatMsg: message.ChatMsg) {
    let newMsgList = this.state.msgList as message.ChatMsg[];
    newMsgList.push(chatMsg);
    this.setState({
      msgList: newMsgList,
      inputContent: "",
    })
  }

  componentDidMount() {    
    this.props.msgManager?.sub(constants.MSG_TCHAT_IN, (cacheChat: Array<message.ChatMsg>) => {
      if (cacheChat === null) {
        return;
      }
      let newMsgList = this.state.msgList as message.ChatMsg[];
      for (let i = 0; i < cacheChat.length; i++) {
        newMsgList.push(cacheChat[i]);
      }
      this.setState({
        msgList: newMsgList,
      });
    });

    // disable enter default behavior of textarea 
    document.getElementById("chat-input")!.addEventListener('keydown', (e) => {
      if (e.key === 'Enter') {
        e.preventDefault();

        let content = this.state.inputContent;
        if (content.length > 0 && content[0] === "/") {
          this.handleCommand(content.slice(1));
        } else {
          this.handleSendMsg(content);
        }
        this.setState({
          inputContent: "",
        });
      }
    });

    this.props.msgManager.pub("request", constants.MSG_TREQUEST_CACHE_CHAT);
  }

  // command doesn't include the first '/'
  handleCommand(command: string) {
    let args = command.split(' ');
    switch (args[0]) {
      case "help":
        this.addNotiMessage(`TStream - Streaming from terminal`);
        this.addNotiMessage(`/name (name) - to set username`);
        break;

      case "name":
        if (args.length === 2) {
          let userConfig = this.getUserConfig()
          if (userConfig == null) {
            let color = constants.COLOR_LIST[Math.floor(Math.random() * (constants.COLOR_LIST.length))];
            userConfig = {
              name: args[1],
              color: color,
            }
          } else {
            userConfig.name = args[1]
          }

          this.setUserConfig(userConfig);
          this.setState({userConfig: userConfig});
          this.addNotiMessage(`Set name successfully to ${userConfig.name}`);

        } else {
          this.addNotiMessage("Invalid command");
        }
        break;
      default: 
        this.addNotiMessage("Invalid command. Type /help to see available commands");
    }

  }

  // display a notify for viewer only 
  addNotiMessage(messsage: string) { 
    let data = {
      Name: '', 
      Content: messsage,
      Color: '', 
      Time: new Date().toISOString(),
    };

    this.addNewMsg(data);

  }

  getUserConfig(): TstreamUser | null {
    const payload = localStorage.getItem(USER_CONFIG_KEY);
    if (payload !== null) {
      const tstreamUser : TstreamUser = JSON.parse(payload);
      return tstreamUser
    } else {
      return null
    }
  }

  setUserConfig(config: TstreamUser) {
    localStorage.setItem(USER_CONFIG_KEY, JSON.stringify(config));
  }

  handleSendMsg(content: string) {
    let tempMsg : string = content.trim();

    // Don't find the user data in the browser
    if (! this.state.userConfig) {

      // ask for first time
      if (!this.state.isWaitingUsername) {
        this.setState({
          tempMsg: tempMsg,
          isWaitingUsername: true,
        });
        this.addNotiMessage("Please enter your username (I.e: elonmusk)");
        return ;

      } else {
        // invalid username
        if (tempMsg.includes(" ") || tempMsg === '') {
          this.addNotiMessage('Username must contain only lower case letters and number');
          return ;

        } else {
          // user just set username

          this.addNotiMessage("You can change name again with command /name (newname)");
          // valid username
          let userConfig : TstreamUser = {
            name: tempMsg,
            color:constants.COLOR_LIST[Math.floor(Math.random() * (constants.COLOR_LIST.length))],
          }

          this.setState({
            userConfig: userConfig,
            isWaitingUsername: false,
          });

          this.setUserConfig(userConfig);

          let data = {
            Name: userConfig.name,
            Content: this.state.tempMsg,
            Color: userConfig.color,
            Time: new Date().toISOString(),
          };

          this.addNewMsg(data);
          this.props.msgManager?.pub(constants.MSG_TCHAT_OUT, data);
        }
      }

    } else {
      if (tempMsg === "") return;
      let data = {
        Name: this.state.userConfig.name,
        Content: tempMsg,
        Color: this.state.userConfig.color,
        Time: new Date().toISOString(),
      };

      this.addNewMsg(data);
      this.props.msgManager?.pub(constants.MSG_TCHAT_OUT, data);
    }
  }

  render() {
    return (
      <div className={`flex flex-col relative scroll-bar-inline ${this.props.className}`} 
        style={{height: this.props.height, width: this.props.width, fontFamily: "'Ubuntu Mono', monospace"}}
      >
        <div id ="chatbox" className="bg-black overflow-y-scroll overflow-x-none p-2 flex flex-col-reverse flex-grow scroll-bar-inline">
          {this.state.msgList.slice(0).reverse().map(
            (item, index) => <ChatSection Name={item.Name} Content={item.Content} Color={item.Color} Time={item.Time} key={index}/>)}
        </div>
        <div id="chat-input" className="w-full flex-shrink-0">
          <TextField
            InputProps={{
              style: {
                  backgroundColor: "rgba(75,85,99,1)",
                  fontFamily: "'Ubuntu Mono', monospace",}
            }}
            placeholder={(this.state.isWaitingUsername) ? "Please enter your name" : "Send a message"}
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
    )
  }
}

export default Chat;
