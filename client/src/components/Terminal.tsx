import React from "react";
import pako from "pako";
import Xterm from "./Xterm";
import * as constants from "../lib/constants";
import * as message from "../types/message";
import * as buffer from "../lib/buffer";

interface Props {
  width: number; // in pixel
  height: number; // in pixel
  rows?: number;
  cols?: number;
  delay?: number;
  refreshInterval?: number; // the writemanager will scan and schedule to write after each refreshInterval. Unit in milliseconds
  className?: string;
}

interface State {
  playing: boolean;

  //delay: number; // in milliseconds
  startTime?: number | undefined; // the absolute start time of the stream. used to calculate the duration of stream or records
  currentTime?: number | undefined; // currenttime of the player relative to the starttime
}

class Terminal extends React.Component<Props, State> {

  termRef = React.createRef<Xterm>();
  queue: message.Wrapper[] = [];

  static defaultProps = { 
    width: -1, 
    height: -1, 
    rows : 0, 
    cols : 0, 
    className:"",
    delay: 0,
    refreshInterval: 200,
  };

  state: State = {
    playing: true,
  }

  constructor(props:Props) {
    super(props);
    this.play();
  }

  rescale() {
    if (!this.props) return;
    const { width, height, rows, cols } = this.props;
    if (this.termRef.current && (width! > 0 || height! > 0)) {
      const core = (this.termRef.current?.terminal as any)._core,
        cellWidth = core._renderService.dimensions.actualCellWidth,
        cellHeight = core._renderService.dimensions.actualCellHeight,
        currentFontSize = this.termRef.current.terminal.getOption('fontSize'),
        termRows = rows! > 0 ? rows : this.termRef.current.terminal.rows,
        termCols = cols! > 0 ? cols : this.termRef.current.terminal.cols,
        hFontSizeMultiplier = height / (cellHeight * termRows!),
        wFontSizeMultiplier = width / (cellWidth * termCols!),
        // method doesn't ensure termianl will 100% fit the required size since fontsize are discrete
        // Another method is to transform scale to fit the window
        // But I haven't figured out why the scaled version sometimes make terminal deformed 
        // after multiple times of apply scale transformation
        newFontSize = Math.floor(
          hFontSizeMultiplier > wFontSizeMultiplier 
          ? currentFontSize * wFontSizeMultiplier 
          : currentFontSize * hFontSizeMultiplier);

      this.termRef.current.terminal.setOption('fontSize', newFontSize);
      this.termRef.current.terminal.refresh(0, termRows! - 1); // force xterm to re-render everything
    }
  }

  play() {
    this.setState({ playing: true });
    this.consume();
  }

  pause() {
    this.setState({ playing: false });
  }

  resetQueue() {
    this.queue = [];
  }

  addQueue(q: message.Wrapper[]) {
    // TODO : carefully insert the queue so that the queue is ensure to be incremented in delaytime
    this.queue.push(...q);
  }

  consume(): any {
    const returnCallback = () => {
      if (this.state.currentTime) this.setState({ currentTime: this.state.currentTime + this.props.refreshInterval! });

      setTimeout(() => {
        this.consume();
      }, this.props.refreshInterval);
    }


    if (!this.state.currentTime || ! this.state.playing || this.queue.length == 0) return returnCallback();

    const currentTime = this.state.currentTime!;
    const endTime = currentTime + this.props.refreshInterval!;

    while (this.queue.length > 0 && this.queue[0].Delay < endTime) {
      if (this.queue[0].Delay > endTime) break;

      const msg: message.Wrapper = this.queue.shift()!;
      const msgTimeout = msg.Delay - currentTime;

      switch (msg.Type) {

        case constants.MSG_TWRITE:
          let bufferData = buffer.str2ab(msg.Data);
          setTimeout(() => this.termRef.current?.writeUtf8(bufferData), msgTimeout);
          break;

        case constants.MSG_TWINSIZE:
            setTimeout(() => this.termRef.current?.resize(msg.Data.Cols, msg.Data.Rows), msgTimeout);
          break;

        default:
            console.error("Unhandled message type: ", msg.Type);
      }
    }

    return returnCallback();
  }

  resize(ws: message.TermSize) {
    this.termRef.current?.resize(ws.Cols, ws.Rows);
  }

  addBlock(block: message.TermWriteBlock) {
    // the starttime of stream or records will be the the starttime of the first block received
    let startTime = this.state.startTime;
    let currentTime = this.state.currentTime;
    if (!startTime || !currentTime) {
      const blockStartTime = (new Date(block.StartTime)).getTime();
      startTime = blockStartTime;
      // the delay is how much the block message sent to server is delayed to its start time
      // it means that the delay will be include the block duration time and a bit of buffer
      // For example : delay = 1.5s and block duration 1s. we have .5 second for the delay of network
      currentTime = (new Date()).getTime() - blockStartTime - (this.props.delay! - block.Duration);
      this.setState({
        startTime: startTime,
        currentTime: currentTime,
      });
    }

    const blockDelayTime = (new Date(block.StartTime)).getTime() - startTime;

    // this is a big chunk of encoding/decoding
    // Since we have to : reduce message size by usign gzip and also
    // every single termwrite have to be decoded, or else the rendering will screw up
    // the whole block often took 9-20 milliseconds to decode a 3 seconds block of message
    let data = pako.ungzip(buffer.str2ab(block.Data), { to: "string" });
    let msgArrayString: string[] = JSON.parse(data);

    let msgArray: message.Wrapper[] = [];
    msgArrayString.forEach((msgString: string) => {
      // re-compute the offset of this message with respect to the render time
      let msg: message.Wrapper = JSON.parse(window.atob(msgString));
      msg.Delay = blockDelayTime + msg.Delay;
      msgArray.push(msg);
      //console.log("msg Delay: ", msg.Delay);
    })

    this.addQueue(msgArray);

  }
  componentDidMount() {
    window.addEventListener("resize", this.rescale);
  }

  componentWillUnmount() {
    window.removeEventListener("resize", this.rescale);
  }

  render() {
    return <div className={`relative overflow-hidden bg-black ${this.props.className}`}
      style={{width: this.props.width!, height: this.props.height!}}>
      <div className="overlay bg-transparent absolute top-0 left-0 z-10 w-full h-full"></div>
      <div className="divref absolute top-1/2 left-1/2 origin-top-left transform -translate-x-1/2 -translate-y-1/2 overflow-hidden">
        <Xterm 
          ref={this.termRef} 
          options={{
            rightClickSelectsWord: false,
              disableStdin: true,
          }}/>
      </div>
    </div>
  }

}

export default Terminal;
