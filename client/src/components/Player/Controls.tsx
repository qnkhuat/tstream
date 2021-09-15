import React from "react";

import PlayArrowIcon from '@mui/icons-material/PlayArrow';
import PauseIcon from '@mui/icons-material/Pause';
import IconButton from '@mui/material/IconButton';
import Slider from '@mui/material/Slider';

interface Props {
  playing?: boolean;
  className?: string;
  onPlay: () => void;
  onPause: () => void;
}

const Controls: React.FC<Props> = ({ onPlay, onPause, playing=false, className = "" }: Props) => {

  //const duration = useMemo(() => {
  //  if (state.recordDuration) return utils.formatDuration(state.recordDuration * 1000);
  //  else return "00:00";
  //}, [recordDuration]);
  const duration = "00:00";

  return <div className={`flex items-center ${className} px-10 mb-5`}>
    <IconButton 
      onClick={() => playing ? onPause() : onPlay()}
      className="mr-4">
      {playing ? <PauseIcon/> : <PlayArrowIcon/>}
    </IconButton>
    <Slider
      className="mr-4"
      size="small"
      defaultValue={0}
      aria-label="Small"
      valueLabelDisplay="auto"
      color="primary"
    />
    <p className="text-sm font-bold">{duration}</p>
  </div>
}


export default Controls;
