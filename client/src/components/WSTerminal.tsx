import React from "react";
import Xterm from "./Xterm";
import * as constants from "../lib/constants";
import PubSub from "../lib/pubsub";

// TODO: add handle % and px
interface Props {
  msgManager: PubSub;
  width: number; // in pixel
  height: number; // in pixel
  className?: string;
}

interface Winsize {
  Rows: number;
  Cols: number;
}


class WSTerminal extends React.Component<Props, {}> {

  static defaultProps = {
    width: -1,
    height: -1,
    className: "",
  }

  termRef : React.RefObject<Xterm>;
  divRef: React.RefObject<HTMLDivElement>;

  constructor(props: Props) {
    super(props)
    this.termRef = React.createRef<Xterm>();
    this.divRef = React.createRef<HTMLDivElement>();
  }


  componentDidMount() {
    this.props.msgManager.sub(constants.MSG_TWRITE, (buffer: Uint8Array) => {
      this.termRef.current?.writeUtf8(buffer);
    })

    this.props.msgManager.sub(constants.MSG_TWINSIZE, (winsize: Winsize) => {
      this.termRef.current?.resize(winsize.Cols, winsize.Rows)
      this.rescale();
    })

    this.props.msgManager.pub("request", constants.MSG_TREQUEST_WINSIZE);
    this.props.msgManager.pub("request", constants.MSG_TREQUEST_CACHE_CONTENT);

    window.addEventListener("resize", () => this.rescale());
    this.rescale();

  }

  componentDidUpdate(prevProps: Props) {
    if (this.props.width != prevProps.width || this.props.height != prevProps.height) {
      this.rescale();
    }
  }

  componentWillUnmount() {
    this.props.msgManager.unsub(constants.MSG_TWRITE);
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
            }}
            ref={this.termRef}/>
        </div>
      </div>
    )
  }

}

export default WSTerminal;
