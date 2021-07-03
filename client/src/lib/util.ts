import urljoin from "url-join";
import dayjs from "dayjs";
export function getWsUrl(sessionID: string): string{
  const wsHost: string = (process.env.REACT_APP_API_URL as string).replace("http", "ws");
  return urljoin(wsHost, "ws", sessionID, "viewer");
}

export function sendWhenConnected(ws: WebSocket, msg: any) {
  setTimeout(() => {
    if (ws.readyState === 1) {
      ws.send(msg);
    } else {
      sendWhenConnected(ws, msg);
    }
  }, 5); // wait 5 milisecond for the connection...
}

export function getUpTime(time: dayjs.Dayjs): string {
  const now = dayjs();
  let diff = now.diff(time, "second");

  let hours = Math.floor(diff / 3600);
  diff = diff - 3600 * hours;

  let minutes = Math.floor(diff / 60);
  diff = diff - minutes * 60;

  let seconds = diff;

  return `${hours}:${minutes > 9 ? minutes : `0${minutes}` }:${seconds> 9 ? seconds: `0${seconds}` }`;
}

