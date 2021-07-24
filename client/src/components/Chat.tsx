import React from "react";
import * as constants from "../lib/constants";
import PubSub from "../lib/pubsub";
import TextField from '@material-ui/core/TextField';
import KeyboardArrowRightRoundedIcon from '@material-ui/icons/KeyboardArrowRightRounded';

// key in local storage
const USER_CONFIG_KEY = "tstreamUser";

interface TstreamUser {
  name: string,
    color: string,
}

interface Props {
  msgManager: PubSub;
  className?: string;
}

export interface ChatMsg {
  Name: string;
  Content: string;
  Color: string;
  Time: string;
}

interface State {
  msgList: ChatMsg[];
  inputContent: string;
  userConfig: TstreamUser | null;
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

  addNewMsg(chatMsg: ChatMsg) {
    let newMsgList = this.state.msgList as ChatMsg[];
    newMsgList.push(chatMsg);
    this.setState({
      msgList: newMsgList,
      inputContent: "",
    })
  }

  componentDidMount() {    
    this.props.msgManager?.sub(constants.MSG_TCHAT_IN, (cacheChat: Array<ChatMsg>) => {
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

    // disable enter default behavior of textarea 
    document.getElementById("chat-input")!.addEventListener('keydown', (e) => {
      if (e.key === 'Enter') {
        e.preventDefault();

        let content = this.state.inputContent;
        if (content.length > 0 && content[0] == "/") {
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
        if (args.length == 2) {
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
        <div className="bottom-0 transform w-full" id="chat-input">
          <div className="border-b border-gray-500 flex-shrink-0 flex items-center justify-between">
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
