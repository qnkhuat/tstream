import * as message from "../../types/message";
import dayjs from "dayjs";

export enum PlayerActionType {
  Play,
    Pause,
    SetManifest,
    JumpTo,
    ChangeRate,
    UpdateCurrentTime,
}

export type PlayerAction = 
  | { type: PlayerActionType.Play }
  | { type: PlayerActionType.Pause }
  | { type: PlayerActionType.SetManifest, payload: { manifest: message.Manifest} }
  | { type: PlayerActionType.JumpTo, payload: { to: number } }
  | { type: PlayerActionType.ChangeRate, payload: { rate: number } }
  | { type: PlayerActionType.UpdateCurrentTime, payload: { currentTime: number } }
;

export const playerActions: {[key: string]: (...args: any[]) => PlayerAction} = {
  play: () =>  ({ type: PlayerActionType.Play }),
  pause: () => ({ type: PlayerActionType.Pause }),
  setManifest: (manifest: message.Manifest) => ({ type:PlayerActionType.SetManifest, payload: { manifest } }),
  jumpTo: (to: number) => ({ type: PlayerActionType.JumpTo, payload: { to } }),
  updateCurrentTime: (currentTime: number) => ({ type: PlayerActionType.UpdateCurrentTime, payload: { currentTime} }),
  changeRate: (rate: number) => ({ type: PlayerActionType.ChangeRate, payload: { rate } }),
}

export interface PlayerState {
  play: boolean;
  currentTime: number;
  rate: number;
  manifest: message.Manifest | null;
  recordDuration: number;
  contentQueue: message.TermWriteBlock[];
}

export const initialState: PlayerState = {
  manifest: null,
  contentQueue: [],
  play: false,
  recordDuration: 0,
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

    case PlayerActionType.SetManifest:
      const { StartTime, StopTime } = action.payload.manifest;
      const duration = dayjs(StartTime).diff(dayjs(StopTime), "millisecond");
      return {
        ...state,
        recordDuration: duration,
        manifest: action.payload.manifest,
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
