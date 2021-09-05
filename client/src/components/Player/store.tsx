export enum PlayerAction {
  Play,
    Pause,
    JumpTo,
    ChangeRate,
}

type Action = 
  | { type: PlayerAction.Play }
  | { type: PlayerAction.Pause}
  | { type: PlayerAction.JumpTo, payload: { to: number } }
  | { type: PlayerAction.ChangeRate, payload: { rate: number } }
;
interface State {
  play: boolean;
  currentTime: number;
  rate: number;
}

export const initialState: State = {
  play: false,
  currentTime: 0,
  rate: 1,
};

export const reducer = (state: State, action: Action) => {
  switch (action.type) {
    case PlayerAction.Play:
      return {
        ...state,
        play: true,
      }

    case PlayerAction.Pause:
      return {
        ...state,
        play: false,
      }

    case PlayerAction.JumpTo:
      return {
        ...state,
        currentTime: action.payload.to,
      }

    case PlayerAction.ChangeRate:
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
