import React from "react";
import { RouteComponentProps, match, useParams } from "react-router-dom";

interface Params {
  username: string;
}

function Room() {
  const params: Params = useParams();
  return (
    <>
      <h3>Rooommmmmmmmmmmm</h3>
      <h3>{params.username}</h3>
    </>
  )
}

export default Room;
