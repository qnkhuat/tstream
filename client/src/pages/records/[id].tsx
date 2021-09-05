import React, { useState, useEffect, useRef } from 'react';

import { RouteComponentProps } from "react-router-dom";

import useWindowSize from "../../hooks/windonwsize";
import useDimension from "../../hooks/dimension";
import Player from "../../components/Player";
import Loading from "../../components/Loading";
import Navbar from "../../components/Navbar";

interface Params {
  Id: string;
  roomKey?: string;
}

interface Props extends RouteComponentProps<Params> {}

const Records: React.FC<Props> = (props: Props) => {
  const [ navbarSize, navbarRef ] = useDimension();
  const windowSize = useWindowSize();

  if (!windowSize) return <Loading/>;
  return <>
    <Navbar ref={navbarRef}/>
    <Player
      id={props.match.params.Id}
      width={windowSize.width}
      height={windowSize.height - navbarSize.height}
    />
  </>
}

export default Records;
