// A Wrapper of XTerm where it resize itself based on the provided with and height
import React, { useEffect } from "react";
import Xterm from "./Xterm";
import * as constants from "../lib/constants";
import * as message from "../types/message";
import * as buffer from "../lib/buffer";
import pako from "pako";
import { accurateInterval } from "../utils";

interface Props {
  width: number; // in pixel
  height: number; // in pixel
  rows?: number;
  cols?: number;
  delay?: number;
  className?: string;
}

const Terminal = React.forwardRef<Xterm, Props>(({ width=-1, height=-1, rows=0, cols = 0, className = "" }, ref) => {
  useEffect(() => {
    const rescale = () => {
      if (ref && typeof ref != "function" && ref.current && (width! > 0 || height! > 0)) {
        const core = (ref.current?.terminal as any)._core,
          cellWidth = core._renderService.dimensions.actualCellWidth,
          cellHeight = core._renderService.dimensions.actualCellHeight,
          currentFontSize = ref.current.terminal.getOption('fontSize'),
          termRows = rows > 0 ? rows : ref.current.terminal.rows,
          termCols = cols > 0 ? cols : ref.current.terminal.cols,
          hFontSizeMultiplier = height / (cellHeight * termRows),
          wFontSizeMultiplier = width / (cellWidth * termCols),
          // method doesn't ensure termianl will 100% fit the required size since fontsize are discrete
          // Another method is to transform scale to fit the window
          // But I haven't figured out why the scaled version sometimes make terminal deformed 
          // after multiple times of apply scale transformation
          newFontSize = Math.floor(
            hFontSizeMultiplier > wFontSizeMultiplier 
            ? currentFontSize * wFontSizeMultiplier 
            : currentFontSize * hFontSizeMultiplier);

        ref.current.terminal.setOption('fontSize', newFontSize);
        ref.current.terminal.refresh(0, termRows - 1); // force xterm to re-render everything
      }
    }

    const handleResize = () => { 
      rescale(); 
    };

    window.addEventListener("resize", handleResize);

    // first refresh
    handleResize();
    return () => {
      window.removeEventListener("resize", handleResize);
    }
  }, [width, height, ref, rows, cols]);


  return <div className={`relative overflow-hidden bg-black ${className}`}
    style={{width: width!, height: height!}}>
    <div className="overlay bg-transparent absolute top-0 left-0 z-10 w-full h-full"></div>
    <div className="divref absolute top-1/2 left-1/2 origin-top-left transform -translate-x-1/2 -translate-y-1/2 overflow-hidden">
      <Xterm 
        ref={ref} 
        options={{
          rightClickSelectsWord: false,
            disableStdin: true,
        }}/>
    </div>
  </div>
});



export default Terminal;
