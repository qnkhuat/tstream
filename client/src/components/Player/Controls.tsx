import React from "react";
import PlayArrowIcon from '@mui/icons-material/PlayArrow';
import PauseIcon from '@mui/icons-material/Pause';
import IconButton from '@mui/material/IconButton';
import Slider from '@mui/material/Slider';
import { PlayerState, PlayerAction, playerActions, PlayerActionType} from "./store";

interface Props {
  className?: string;
  state: PlayerState;
  dispatch: React.Dispatch<PlayerAction>;
}

const Controls: React.FC<Props> = ({ state, dispatch, className = "" }: Props) => {
  return <div className={`flex items-center ${className} px-10 mb-5`}>
    <IconButton onClick={() => state.play ? dispatch(playerActions.pause()) : dispatch(playerActions.play()) }
      className="mr-4"
    >
      {state.play && <PauseIcon/>}
      {!state.play && <PlayArrowIcon/>}
    </IconButton>
    <Slider
      size="small"
      defaultValue={0}
      aria-label="Small"
      valueLabelDisplay="auto"
      color="primary"
      //onChange={(e, value) => {if(handleJumpTo) handleJumpTo(value[0])}}
    />
  </div>
}


export default Controls;
