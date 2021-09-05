import { useState } from "react";

export default function useAsyncReducer(reducer: any, initState:any) {
  const [state, setState] = useState(initState),
    dispatchState = async (action: any) => setState(await reducer(state, action));
  return [state, dispatchState];
}
