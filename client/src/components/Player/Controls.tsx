import React, { useEffect } from "react";

import PlayArrowIcon from '@mui/icons-material/PlayArrow';
import PauseIcon from '@mui/icons-material/Pause';
import IconButton from '@mui/material/IconButton';
import Slider from '@mui/material/Slider';

interface Props {
  duration: number;
  currentTime: number;
  onPlay: () => void;
  onPause: () => void;
  onChange?: (value: number) => void;
  playing: boolean;
  className?: string;
}

const Controls: React.FC<Props> = ({ onPlay, onPause, onChange, duration, currentTime, playing = false, className = "" }: Props) => {

  useEffect(() => {
    const togglePlay = (e: any) => {
      e.keyCode== 32 && playing ? onPause() : onPlay();
    }
    window.addEventListener("keyup", togglePlay)

    return () => {
      window.removeEventListener("keyup", togglePlay);
    }

  }, [onPlay, onPause]);

  const valueLabelFormat = (value: number) => {
    return Math.round(value /1000);
  }
  return <div className={`flex items-center ${className} px-10 mb-5`}>
    <IconButton 
      onClick={() => playing ? onPause() : onPlay()}
      className="mr-4">
      {playing ? <PauseIcon/> : <PlayArrowIcon/>}
    </IconButton>
    <Slider
      className="mr-4"
      size="small"
      defaultValue={currentTime}
      value={currentTime}
      step={0.5}
      max={Math.round(duration)}
      aria-label="Small"
      valueLabelDisplay="auto"
      onChange={ (_, value) => { if(onChange && typeof value === 'number') onChange(value) }}
      valueLabelFormat={valueLabelFormat}
      color="primary"
    />
    <p className="text-sm font-bold">{currentTime/1000}</p>
  </div>
}


export default Controls;
