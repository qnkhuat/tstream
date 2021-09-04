import React, { createContext, useReducer } from "react";

const initialState = {};

const PlaybackContext = createContext();

interface State {
  play: boolean;
  currentTime: number;
  rate: number;
}

export enum PlaybackAction {
  Play,
    Pause,
    JumpTo,
    ChangeRate,

}

type Action = 
  | { type: PlaybackAction.Play }
  | { type: PlaybackAction.Pause}
  | { type: PlaybackAction.JumpTo, payload: { to: number } }
  | { type: PlaybackAction.ChangeRate, payload: { rate: number } }
;

const reducer = (state: State, action: Action) => {
  switch (action.type) {
    case  PlaybackAction.Play:
      return {
        ...state,
        play: true,

      }

    case  PlaybackAction.Pause:
      return {
        ...state,
        play: false,

      }

    default:
      return {
        ...state,
      }

  }

}

const PlaybackContextProvider: React.FC = (props) => {

  const [ state, dispatch ] = useReducer(reducer, {
    play: false,
    currentTime: 0,
    rate: 1,
  });

  return <PlaybackContext.Provider value={{state, dispatch}}>
    {props.children}
  </PlaybackContext.Provider >

}
export default PlaybackContextProvider;
