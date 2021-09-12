import React, { useMemo } from "react";

import PlayArrowIcon from '@mui/icons-material/PlayArrow';
import PauseIcon from '@mui/icons-material/Pause';
import IconButton from '@mui/material/IconButton';
import Slider from '@mui/material/Slider';

import * as utils from  "../../utils";
import { PlayerState, PlayerAction, playerActions } from "./store";

interface Props {
  className?: string;
  state: PlayerState;
  dispatch: React.Dispatch<PlayerAction>;
}

const Controls: React.FC<Props> = ({ state, dispatch, className = "" }: Props) => {

  const duration = useMemo(() => {
    if (state.recordDuration) return utils.formatDuration(state.recordDuration * 1000);
    else return "00:00";
  }, [state.recordDuration]);

  return <div className={`flex items-center ${className} px-10 mb-5`}>
    <IconButton 
      onClick={() => state.play ? dispatch(playerActions.pause()) : dispatch(playerActions.play()) }
      className="mr-4">
      {state.play && <PauseIcon/>}
      {!state.play && <PlayArrowIcon/>}
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
