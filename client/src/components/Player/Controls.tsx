import React from "react";
import PlayArrowIcon from '@mui/icons-material/PlayArrow';
import PauseIcon from '@mui/icons-material/Pause';
import IconButton from '@mui/material/IconButton';
import Slider from '@mui/material/Slider';

interface Props {
  play: boolean;
  rate: number;
  currentTime: number;
  handlePlay?: () => void;
  handlePause?: () => void;
  handleJumpTo?: (value: number) => void;
  className?: string;
}

const Controls: React.FC<Props> = ({ play = false, rate, handlePlay, handlePause, handleJumpTo, className = "" }: Props) => {
  return <div className={`flex items-center ${className} px-10 mb-5`}>
    <IconButton onClick={play ? handlePause : handlePlay}
      className="mr-4"
    >
      {play && <PauseIcon/>}
      {!play && <PlayArrowIcon/>}
    </IconButton>
    <Slider
      size="small"
      defaultValue={0}
      aria-label="Small"
      valueLabelDisplay="auto"
      color="primary"
      onChange={(e, value) => {if(handleJumpTo) handleJumpTo(value[0])}}
    />
  </div>
}


export default Controls;
