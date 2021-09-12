import urljoin from "url-join";

export function getWsUrl(sessionID: string): string{
  const wsHost: string = (process.env.REACT_APP_API_URL as string).replace("http", "ws");
  return urljoin(wsHost, "ws", sessionID);
}

export function sendWhenConnected(ws: WebSocket, msg: string, n: number = 0, maxTries: number = 100) {
  if (ws.readyState === 2 || ws.readyState === 3) return;
  setTimeout(() => {
    if (ws.readyState === 1) {
      ws.send(msg);
    } else if (n < maxTries) {
      sendWhenConnected(ws, msg, n + 1);
    } else{
      console.error("Exceed tries to send message: ", msg);
    }
  }, 10); // wait 10 milisecond for the connection...
}

// duration is in seconds
export function formatDuration(duration: number, full: boolean = false): string {

  let hours = Math.floor(duration / 3600);
  duration = duration - 3600 * hours;

  let minutes = Math.floor(duration / 60);
  duration = duration - minutes * 60;

  let seconds = duration;

  let text = `${minutes > 9 ? minutes : `0${minutes}` }:${seconds> 9 ? seconds: `0${seconds}`}` ;
  if (hours > 0 || full) {
    text = `${hours > 9 ? hours : `0${hours}`}:${text}`;
  }
  return text;
}

