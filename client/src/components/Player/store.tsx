import * as message from "../../types/message";

export enum PlayerActionType {
  Play,
    Pause,
    JumpTo,
    ChangeRate,
    UpdateCurrentTime,
}

export type PlayerAction = 
  | { type: PlayerActionType.Play }
  | { type: PlayerActionType.Pause }
  | { type: PlayerActionType.JumpTo, payload: { to: number } }
  | { type: PlayerActionType.ChangeRate, payload: { rate: number } }
  | { type: PlayerActionType.UpdateCurrentTime, payload: { currentTime: number } }
;

export const a: {[key: string]: number} = {
  a: 3
}
export const playerActions: {[key: string]: (...args: any[]) => PlayerAction} = {
  play: () =>  ({ type: PlayerActionType.Play }),
  pause: () => ({ type: PlayerActionType.Pause }),
  jumpTo: (to: number) => ({ type: PlayerActionType.JumpTo, payload: { to } }),
  updateCurrentTime: (currentTime: number) => ({ type: PlayerActionType.UpdateCurrentTime, payload: { currentTime} }),
  changeRate: (rate: number) => ({ type: PlayerActionType.ChangeRate, payload: { rate } }),
}

export interface PlayerState {
  play: boolean;
  currentTime: number;
  rate: number;
  manifest: message.Manifest | null;
  contentQueue: message.TermWriteBlock[];
}

export const initialState: PlayerState = {
  manifest: null,
  contentQueue: [],
  play: false,
  currentTime: 0,
  rate: 1,
};

export const playerReducer = (state: PlayerState, action: PlayerAction) => {
  switch (action.type) {
    case PlayerActionType.Play:
      return {
        ...state,
        play: true,
      }

    case PlayerActionType.Pause:
      return {
        ...state,
        play: false,
      }

    case PlayerActionType.JumpTo:
      return {
        ...state,
        currentTime: action.payload.to,
      }

    case PlayerActionType.ChangeRate:
      return {
        ...state,
        rate: action.payload.rate
      }

    default:
      return {
        ...state,
      }

  }
}
