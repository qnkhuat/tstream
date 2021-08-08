import React from "react";
import Xterm from "./Xterm";
import * as constants from "../lib/constants";
import * as message from "../lib/message";
import * as base64 from "../lib/base64";
import * as pako from "pako";
import PubSub from "../lib/pubsub";


interface Winsize {
  Rows: number;
  Cols: number;
}

// TODO: add handle % and px for size
interface Props {
  msgManager: PubSub;
  width: number; // in pixel
  height: number; // in pixel
  delay?: number;
  className?: string;
}

class WSTerminal extends React.Component<Props, {}> {

  static defaultProps = {
    width: -1,
    height: -1,
    delay: 0,
    className: "",
  }

  termRef : React.RefObject<Xterm>;
  divRef: React.RefObject<HTMLDivElement>;
  writeManager: WriteManager;


  constructor(props: Props) {
    super(props)
    this.termRef = React.createRef<Xterm>();
    this.divRef = React.createRef<HTMLDivElement>();
    let writeCB = (buffer: Uint8Array) => {
      this.termRef.current?.writeUtf8(buffer);
    };

    let winsizeCB = (ws: Winsize) => {
      this.termRef.current?.resize(ws.Cols, ws.Rows);
      this.rescale();
    }

    this.writeManager = new WriteManager(writeCB, winsizeCB, this.props.delay);
  }


  componentDidMount() {
    this.props.msgManager.sub(constants.MSG_TWRITEBLOCK, (block: message.TermWriteBlock) => {
      this.writeManager.addBlock(block);
    })

    this.props.msgManager.pub("request", constants.MSG_TREQUEST_WINSIZE);
    this.props.msgManager.pub("request", constants.MSG_TREQUEST_CACHE_CONTENT);

    window.addEventListener("resize", () => this.rescale());
    this.rescale();

  }

  componentDidUpdate(prevProps: Props) {
    if (this.props.width !== prevProps.width || this.props.height !== prevProps.height) {
      this.rescale();
    }
  }

  componentWillUnmount() {
    this.props.msgManager.unsub(constants.MSG_TWRITEBLOCK);
    this.props.msgManager.unsub(constants.MSG_TWINSIZE);
  }


  rescale() {
    if (this.termRef.current && this.divRef.current && (this.props.width! > 0 || this.props.height! > 0)) {
      const core = (this.termRef.current?.terminal as any)._core,
        cellWidth = core._renderService.dimensions.actualCellWidth,
        cellHeight = core._renderService.dimensions.actualCellHeight,
        currentFontSize = this.termRef.current.terminal.getOption('fontSize'),
        rows = this.termRef.current.terminal.rows,
        cols = this.termRef.current.terminal.cols,
        hFontSizeMultiplier = this.props.height! / (cellHeight * rows),
        wFontSizeMultiplier = this.props.width! / (cellWidth * cols),
        // This method doesn't ensure termianl will fully fit the required size since fontsize are discrete
        // Another method is to transform scale to fit the window
        // But I haven't figured out why the scaled version sometimes make terminal deformed after multiple times of apply scale transformation
        newFontSize = Math.floor(hFontSizeMultiplier > wFontSizeMultiplier ? currentFontSize * wFontSizeMultiplier : currentFontSize * hFontSizeMultiplier);

      this.termRef.current.terminal.setOption('fontSize', newFontSize);
      this.termRef.current.terminal.refresh(0, rows-1); // force xterm to re-render everything
    }
  }


  render() {
    return (
      <div className={`relative ${this.props.className} overflow-hidden`}
        style={{width: this.props.width!, height: this.props.height!}}>
        <div ref={this.divRef}
          className="divref absolute top-1/2 left-1/2 origin-top-left transform -translate-x-1/2 -translate-y-1/2 overflow-hidden">
          <Xterm
            options={{
              rightClickSelectsWord: false,
              disableStdin: true,
            }}
            ref={this.termRef}/>
        </div>
      </div>
    )
  }

}

class WriteManager {

  queue: message.Wrapper[] = [];
  writeCB: (arr:Uint8Array) => void;
  winsizeCB: (ws:Winsize) => void
  delay: number; // in milliseconds

  constructor(writeCB: (arr: Uint8Array) => void, winsizeCB: (ws: Winsize) => void, delay: number = 0) {
    this.writeCB = writeCB;
    this.winsizeCB = winsizeCB;
    this.delay = delay;
  }

  resetQueue() {
    this.queue = [];
  }
  

  addQueue(q: message.Wrapper[]) {
    this.queue.push(...q); // Concatnate
    this.consume();
  }

  consume() {
    if (this.queue.length == 0) {
      return
    } else {
      
      // any message has offset < 0 => messages from the past with respect to render time
      // concat all these messages into one buffer and render at once
      let bufferArray: Uint8Array[] = [];
      while (true && this.queue.length != 0) {
        let msg = this.queue[0];

        if (msg.Delay < 0) {
          switch (msg.Type) {
            case constants.MSG_TWRITE:
              let buffer = base64.str2ab(msg.Data)
              bufferArray.push(buffer);
              break;

            case constants.MSG_TWINSIZE:
              this.winsizeCB(msg.Data);
              break;

            default:
              console.error("Unhandled message type: ", msg.Type);
          }

          this.queue.shift();
        } else  break;
      }
      if ( bufferArray.length > 0) this.writeCB(base64.concatab(bufferArray));

      // schedule to render upcomming messages
      // TODO: are there any ways we don't have to create many settimeout liek this?
      // tried sequentially call to settimeout but the performance is much worse
      this.queue.forEach((msg) => {
        switch (msg.Type) {

          case constants.MSG_TWRITE:
            let buffer = base64.str2ab(msg.Data);
            setTimeout(() => {
              this.writeCB(buffer);
            }, msg.Delay);
              break;

          case constants.MSG_TWINSIZE:
            setTimeout(() => {
              this.winsizeCB(msg.Data);
            }, msg.Delay);
              break;

          default:
            console.error("Unhandled message type: ", msg.Type);
        }

      })
      this.resetQueue();
    }
  }

  addBlock(block: message.TermWriteBlock) {
    // when viewers receive this block
    // it only contains the actual start-time
    // we need to be able to re-compute the render time based on 
    // - now time
    // - when does this block being created
    // - the delay factor. In case of play back the delay = now - stream sesion start time
    let blockDelayTime = (new Date()).getTime() - (new Date(block.StartTime)).getTime() - this.delay;

    // this is a big chunk of encoding/decoding
    // Since we have to : reduce message size by usign gzip and also
    // every single termwrite have to be decoded, or else the rendering will screw up
    // the whole block often took 9-20 milliseconds to decode a 3 seconds block of message
    let data = pako.ungzip(base64.str2ab(block.Data));
    let msgArrayString: string[] = JSON.parse(base64.ab2str(data));

    let msgArray: message.Wrapper[] = [];
    msgArrayString.forEach((msgString: string, i) => {
      // re-compute the offset of this message with respect to the render time
      let msg: message.Wrapper = JSON.parse(window.atob(msgString));
      msg.Delay = msg.Delay - blockDelayTime;
      msgArray.push(msg);
    })

    this.addQueue(msgArray);
    
  }
}


export default WSTerminal;
