import React from 'react';
import { RouteComponentProps } from "react-router-dom";
interface Params {
  roomID: string;
  roomKey?: string;
}

interface Props extends RouteComponentProps<Params> {
  id: number;
}

const Player: React.FC<Props> = () => {
  return <>
  </>
}

export default Player;
