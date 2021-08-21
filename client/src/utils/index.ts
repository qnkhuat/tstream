import urljoin from "url-join";
import dayjs from "dayjs";

export function getWsUrl(sessionID: string): string{
  const wsHost: string = (process.env.REACT_APP_API_URL as string).replace("http", "ws");
  return urljoin(wsHost, "ws", sessionID);
}

export function sendWhenConnected(ws: WebSocket, msg: string, n: number = 0, maxTries: number = 100) {
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

export function formatDuration(from:dayjs.Dayjs, to: dayjs.Dayjs): string {
  let diff = from.diff(to, "second");

  let hours = Math.floor(diff / 3600);
  diff = diff - 3600 * hours;

  let minutes = Math.floor(diff / 60);
  diff = diff - minutes * 60;

  let seconds = diff;

  return `${hours > 9 ? hours : `0${hours}`}:${minutes > 9 ? minutes : `0${minutes}` }:${seconds> 9 ? seconds: `0${seconds}` }`;
}

