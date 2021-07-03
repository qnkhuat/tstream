import urljoin from "url-join";

export function getWsUrl(sessionID: string): string{
  const wsHost: string = (process.env.REACT_APP_API_URL as string).replace("http", "ws");
  return urljoin(wsHost, "ws", sessionID, "viewer");
}

export function sendWhenConnected(ws: WebSocket, msg: any) {
  setTimeout(() => {
    if (ws.readyState === 1) {
      console.log("Connection is made")
      ws.send(msg);
    } else {
      console.log("wait for connection...")
      sendWhenConnected(ws, msg);
    }
  }, 5); // wait 5 milisecond for the connection...
}
