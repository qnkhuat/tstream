import React, { useState, useEffect, useRef } from "react";
import * as util from "../lib/util";
import * as constants from "../lib/constants";

import Box from '@material-ui/core/Box';
import Stack from '@material-ui/core/Stack';
import Slider from '@material-ui/core/Slider';
import VolumeMuteIcon from '@material-ui/icons/VolumeMute';
import VolumeDown from '@material-ui/icons/VolumeDown';
import VolumeUp from '@material-ui/icons/VolumeUp';

interface Props {
  roomID: string;
  volume?: number;
  className?: string;
}

const AudioRTC: React.FC<Props> = ({ roomID, volume = 1, className }) => {

  const [ vol, setVol] = useState(volume);
  const mediaDivRef = useRef<HTMLDivElement>(null);

  const handleChangeVol =  (event: Event, newValue: number | number[]) => {
    setVol(newValue as number);
    if (mediaDivRef.current) {
      Array.from(mediaDivRef.current?.getElementsByTagName("audio")).forEach((el) => {
        el.volume = newValue as number
        if (el.paused) el.play();
      });
    }
  };


  useEffect(() => {
    let peerConn = new RTCPeerConnection();

    const wsUrl = util.getWsUrl(roomID);
    const ws = new WebSocket(wsUrl);

    const clientInfo = {
      Type: constants.MSG_TCLIENT_INFO,
      Data:  {
        Name: roomID,
        Role: constants.MSG_FRTC_TYPE_CONSUMER,
      }
    };

    util.sendWhenConnected(ws, JSON.stringify(clientInfo));

    peerConn.ontrack = (ev: RTCTrackEvent)  => {
      let el = document.createElement(ev.track.kind) as HTMLMediaElement;
      el.srcObject = ev.streams[0];
      el.autoplay = true;
      el.controls = true;

      ev.track.onmute = function(ev) {
        el.play();
      }

      ev.streams[0].onremovetrack = ({track}) => {
        if (el.parentNode) {
          el.parentNode.removeChild(el);
        }
      }
      mediaDivRef.current?.appendChild(el);
    }

    peerConn.onconnectionstatechange = (ev: Event) => {
      console.log("State change: ", ev);
    }

    // listen to onicecandidate event and send it back to server
    peerConn.onicecandidate = (ev) => {
      if (ev.candidate) {
        const candidate = {
          Type: constants.MSG_TRTC,
          Data: {
            Event: constants.MSG_FRTC_EVENT_CANDIDATE,
            Data: JSON.stringify(ev.candidate),
          }
        };
        util.sendWhenConnected(ws, JSON.stringify(candidate));
      }
    }

    ws.onclose = (err) => {
      console.log("Websocket has closed", err);
    }

    ws.onmessage = (ev: MessageEvent) => {
      const msg = JSON.parse(ev.data)

      if (msg.Type != constants.MSG_TRTC) console.error("Expected RTC message");

      const event = msg.Data;

      switch (event.Event) {
        case constants.MSG_FRTC_EVENT_OFFER:
          let offer = JSON.parse(event.Data)
          if (!offer) {
            return console.log('failed to parse answer')
          }
          peerConn.setRemoteDescription(offer)
          peerConn.createAnswer().then(answer => {
            peerConn.setLocalDescription(answer)
            ws.send(JSON.stringify({
              Type: constants.MSG_TRTC,
              Data: {Event: constants.MSG_FRTC_EVENT_ANSWER, Data: JSON.stringify(answer)}
            }))
          }).catch((e) => console.error("Failed to set local description: ", e));
          return

        case constants.MSG_FRTC_EVENT_CANDIDATE:
          let candidate = JSON.parse(msg.Data.Data)
          if (!candidate) {
            return console.log('failed to parse candidate')
          }

          peerConn.addIceCandidate(candidate)
          return

        default:
          console.error("Invalid event: ", event.Event);
          return
      } 
    }
  }, [roomID]);

  return (
    <div className={`${className}`} >
      <div className="hidden" ref={mediaDivRef}/>
      <div className="w-auto bg-gray-700 rounded-full group hover:block hover:mr-2 flex flex-row items-center p-1">
        {volume === 0 && <VolumeMuteIcon fontSize="small"/>}
        {volume <= 0.4 && volume > 0 && <VolumeDown fontSize="small"/>}
        {volume > 0.4 && <VolumeUp fontSize="small"/>}
        <Slider className="py-0 mx-2 hidden group-hover:block w-28" size="small" aria-label="Volume" step={0.1} min={0} max={1} value={volume} onChange={handleChangeVol} />
      </div>
    </div>)
}

export default AudioRTC;
