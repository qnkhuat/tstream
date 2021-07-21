import React, { useState, useEffect, useRef } from "react";
import * as util from "../lib/util";
import * as constants from "../lib/constants";

import Button from '@material-ui/core/Button';
import Modal from '@material-ui/core/Modal';
import IconButton from '@material-ui/core/IconButton';
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

interface State {
  lastVolume: number;
  volume: number;
  isPlayed: boolean;
  trackIDs: string[];
  openModal: boolean;
}

class AudioRTC extends React.Component<Props, State> {
  mediaDivRef: React.RefObject<HTMLDivElement> = React.createRef<HTMLDivElement>();
  wsConn: WebSocket | null = null;
  peerConn: RTCPeerConnection | null = null;

  constructor(props: Props) {
    super(props);

    this.state =  {
      volume: props.volume || 0,
      lastVolume: 1,
      isPlayed: false,
      trackIDs: [],
      openModal: false,
    }
  }

  playTracks(volume: number)   {
    if (this.mediaDivRef.current) {
      Array.from(this.mediaDivRef.current?.getElementsByTagName("audio")).forEach((el) => {
        this.setState({isPlayed: true});
        if (el.paused) { 
          el.play().
            catch((e) => {
              this.setState({isPlayed: false, volume: 0, openModal: true});
            });
        }
        el.volume = volume as number;
      });
    } 
  }

  handleChangeVolume(event: Event | null, newValue: number | number[]) {
    this.setState({volume: newValue as number});
    this.playTracks(newValue as number);
  };

  toggleMute(event: React.MouseEvent) {
    this.setState({lastVolume: this.state.volume});
    if (this.state.volume > 0) this.handleChangeVolume(null, 0);
    else this.handleChangeVolume(null, this.state.lastVolume);
  }

  componentDidMount() {
    let peerConn = new RTCPeerConnection();
    this.peerConn = peerConn;

    const wsUrl = util.getWsUrl(this.props.roomID);
    const wsConn = new WebSocket(wsUrl);
    this.wsConn = wsConn;

    // Send client info to enter the room
    const clientInfo = {
      Type: constants.MSG_TCLIENT_INFO,
      Data:  {
        Name: this.props.roomID,
        Role: constants.MSG_FRTC_TYPE_CONSUMER,
      }
    };

    util.sendWhenConnected(this.wsConn, JSON.stringify(clientInfo));

    this.peerConn.ontrack = (ev: RTCTrackEvent)  => {
      if (!this.state.isPlayed) {
        this.setState({openModal:true});
      }

      const tempTrackIDs = [...this.state.trackIDs];
      tempTrackIDs.push(ev.track.id);
      this.setState({trackIDs: tempTrackIDs});

      let el = document.createElement(ev.track.kind) as HTMLMediaElement;
      el.srcObject = ev.streams[0];
      el.autoplay = true;

      ev.streams[0].onremovetrack = ({track}) => {
        // remove track from track list
        const tempTrackIDs = [...this.state.trackIDs];
        const index = tempTrackIDs.indexOf(ev.track.id);
        if (index > -1) {
          tempTrackIDs.splice(index, 1);
          this.setState({trackIDs: tempTrackIDs});
        }

        if (el.parentNode) {
          el.parentNode.removeChild(el);
        }
      }
      this.mediaDivRef.current?.appendChild(el);
    }

    this.peerConn.onconnectionstatechange = (ev: Event) => {
      switch (this.peerConn?.connectionState) {
        case "connected":
          break;
        case "disconnected":
        case "failed":
        case "closed":
          // remove all tracks
          this.setState({trackIDs: []});
          break;
      }
      console.log("State change: ", this.peerConn?.connectionState);
    }

    // listen to onicecandidate event and send it back to server
    this.peerConn.onicecandidate = (ev) => {
      if (ev.candidate) {
        const candidate = {
          Type: constants.MSG_TRTC,
          Data: {
            Event: constants.MSG_FRTC_EVENT_CANDIDATE,
            Data: JSON.stringify(ev.candidate),
          }
        };
        if (this.wsConn) util.sendWhenConnected(this.wsConn, JSON.stringify(candidate));
      }
    }

    this.wsConn.onclose = (err) => {
      console.log("Websocket has closed", err);
    }

    this.wsConn.onmessage = (ev: MessageEvent) => {
      const msg = JSON.parse(ev.data)

      if (msg.Type != constants.MSG_TRTC) console.error("Expected RTC message");

      const event = msg.Data;

      switch (event.Event) {
        case constants.MSG_FRTC_EVENT_OFFER:
          let offer = JSON.parse(event.Data)
          if (!offer) {
            return console.log('failed to parse answer')
          }

          this.peerConn?.setRemoteDescription(offer);
          this.peerConn?.createAnswer().then(answer => {
            this.peerConn?.setLocalDescription(answer)
            this.wsConn?.send(JSON.stringify({
              Type: constants.MSG_TRTC,
              Data: {Event: constants.MSG_FRTC_EVENT_ANSWER, Data: JSON.stringify(answer)}
            }));
          }).catch((e) => console.error("Failed to set local description: ", e));
          return

        case constants.MSG_FRTC_EVENT_CANDIDATE:
            let candidate = JSON.parse(msg.Data.Data)
          if (!candidate) {
            return console.log('failed to parse candidate')
          }

          this.peerConn?.addIceCandidate(candidate)
          return

        default:
            console.error("Invalid event: ", event.Event);
          return
      } 
    }

  }

  componentWillUnmount() {
    this.wsConn?.close();
    this.peerConn?.close();
  }

  render() {
    const hasTracks = this.state.trackIDs.length > 0;
    return (
      <div className={`${this.props.className}`} >
        <div className="hidden" ref={this.mediaDivRef}/>
        {hasTracks &&
        <>
          <div className="w-auto bg-gray-700 rounded-full group hover:block hover:mr-2 flex flex-row items-center p-1">
            <IconButton onClick={this.toggleMute.bind(this)} className="p-1">
              {(!this.state.isPlayed || this.state.volume === 0) && <VolumeMuteIcon fontSize="small" />}
              {this.state.isPlayed && this.state.volume <= 0.4 && this.state.volume > 0 && <VolumeDown fontSize="small" />}
              {this.state.isPlayed && this.state.volume > 0.4 && <VolumeUp fontSize="small" />}
            </IconButton>
            <Slider className="py-0 mx-4 hidden group-hover:block w-28" aria-label="Volume" step={0.1} min={0} max={1} 
              value={this.state.volume} onChange={this.handleChangeVolume.bind(this)} />
          </div>
          <Modal
            open={this.state.openModal}
            onClose={() => this.setState({openModal: false})}
            aria-labelledby="modal-modal-title"
            aria-describedby="modal-modal-description">
            <div className="fixed top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 p-4 bg-gray-900 border border-gray-800 rounded">
              <p className="font-bold text-center">Streamer has enabled voice chat<br></br>Do you want to turn on audio?</p>
              <br></br>
              <div className="flex justify-center">
                <Button
                  className="font-bold mr-4"
                  variant="outlined"
                  onClick={(e) => {
                    this.playTracks(1);
                    this.setState({openModal: false, volume:1});
                  }}>Yes</Button>
                <Button
                  variant="outlined"
                  className="font-bold bg"
                  onClick={(e) => this.setState({openModal: false})}>No</Button>
              </div>
            </div>

          </Modal>
        </>
        }
      </div>)
  }
}

export default AudioRTC;
