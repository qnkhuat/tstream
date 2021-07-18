/*  SFU - Selective Forwarding Unit
Handle webrtc connections.
Primary used for real-time voice broadcasting

General knowledge about WebRTC API
https://developer.mozilla.org/en-US/docs/Web/API/WebRTC_API/Signaling_and_video_calling

Base code: https://github.com/pion/example-webrtc-applications/blob/master/sfu-ws/main.go
*/
package room

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
	"github.com/qnkhuat/tstream/pkg/message"
	"log"
	"sync"
	"time"
)

type Participant struct {
	peer   *webrtc.PeerConnection
	client *Client // contain role and websocket connection
}

type SFU struct {
	lock         sync.RWMutex
	trackLocals  map[string]*webrtc.TrackLocalStaticRTP
	participants map[string]*Participant
}

func NewSFU() *SFU {
	trackLocals := map[string]*webrtc.TrackLocalStaticRTP{}
	participants := map[string]*Participant{} // contain both producers and consumers
	return &SFU{
		trackLocals:  trackLocals,
		participants: participants,
	}
}

func (s *SFU) Start() {
	// request a keyframe every 3 seconds
	go func() {
		for range time.NewTicker(time.Second * 3).C {
			s.sendKeyFrame()
		}
	}()
}

// TODO : break down this func
func (s *SFU) AddPeer(cl *Client) error {

	log.Printf("")
	peerConn, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		log.Printf("Failed to init peer connection: %s", err)
		return err
	}
	defer peerConn.Close()
	defer cl.Close()
	log.Printf("")

	// Accept one audio and one video track incoming
	for _, typ := range []webrtc.RTPCodecType{webrtc.RTPCodecTypeVideo, webrtc.RTPCodecTypeAudio} {
		if _, err := peerConn.AddTransceiverFromKind(typ, webrtc.RTPTransceiverInit{
			Direction: webrtc.RTPTransceiverDirectionRecvonly,
		}); err != nil {
			log.Printf("Failed to add transeiver: %s", err)
			return err
		}
	}
	log.Printf("")

	participant := &Participant{
		peer:   peerConn,
		client: cl,
	}
	log.Printf("")
	participantID := s.newParticipantID()
	//s.lock.Lock()
	s.participants[participantID] = participant
	//s.lock.Unlock()
	log.Printf("")

	// Trickle ICE. Emit server candidate to client
	peerConn.OnICECandidate(func(ice *webrtc.ICECandidate) {
		if ice == nil {
			return
		}

		candidate, err := json.Marshal(ice.ToJSON())
		if err != nil {
			log.Println(err)
			return
		}

		payload := message.Wrapper{
			Type: message.TRTC,
			Data: message.RTC{
				Event: message.RTCCandidate,
				Data:  string(candidate),
			}}

		cl.Out <- payload
	})
	log.Printf("")

	peerConn.OnConnectionStateChange(func(p webrtc.PeerConnectionState) {
		switch p {

		case webrtc.PeerConnectionStateFailed:
			if err := peerConn.Close(); err != nil {
				log.Print(err)
			}

		case webrtc.PeerConnectionStateClosed, webrtc.PeerConnectionStateDisconnected:
			s.removeParticipant(participantID)

		case webrtc.PeerConnectionStateConnected:
			// nothing yet

		default:
			log.Printf("Not implemented: %s", p)
		}

	})
	log.Printf("")

	// Add all current tracks to this peer
	for _, track := range s.trackLocals {
		if _, err := peerConn.AddTrack(track); err != nil {
			log.Printf("Failed to add track: %s", err)
		}
	}

	log.Printf("")
	// only producer can broadcast
	if cl.Role() == message.RProducerRTC {
		peerConn.OnTrack(func(t *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
			// Create a track to fan out our incoming video to all peerse
			trackLocal := s.addLocalTrack(t)
			defer s.removeLocalTrack(t.ID())

			buf := make([]byte, 1500)
			for {
				log.Printf("Getting :%s", trackLocal.Kind())
				i, _, err := t.Read(buf)
				if err != nil {
					return
				}

				if _, err = trackLocal.Write(buf[:i]); err != nil {
					return
				}
			}
		})
	}
	log.Printf("")

	err = s.sendOffer(participant)

	for {
		msg := <-cl.In

		if msg.Type != message.TRTC {
			log.Printf("Expected RTCEvent, Got: %s", msg.Type)
			continue
		}

		rtcMsg := message.RTC{}
		if err := message.ToStruct(msg.Data, &rtcMsg); err != nil {
			log.Printf("Failed to decode RTC message: %v", msg.Data)
			continue
		}

		switch rtcMsg.Event {

		case message.RTCCandidate:
			candidate := webrtc.ICECandidateInit{}
			if err := json.Unmarshal([]byte(rtcMsg.Data), &candidate); err != nil {
				log.Println(err)
				return err
			}

			if err := peerConn.AddICECandidate(candidate); err != nil {
				log.Println(err)
				return err
			}

		case message.RTCAnswer:
			answer := webrtc.SessionDescription{}
			if err := json.Unmarshal([]byte(rtcMsg.Data), &answer); err != nil {
				log.Println(err)
				return err
			}

			if err := peerConn.SetRemoteDescription(answer); err != nil {
				log.Println(err)
				return err
			}
		default:
			log.Printf("Invalid RTCEvent: %s", rtcMsg.Event)
		}

	}

	return nil
}

// TODO: add filter so that only sending offer for who need to update
// Sync tracklocals with all peers so everyone get the right track at the right time
func (s *SFU) syncPeers() {

	attemptSync := func() (tryAgain bool) {
		s.lock.RLock()
		defer s.lock.RUnlock()
		for _, participant := range s.participants {

			// map of sender we already are seanding, so we don't double send
			existingSenders := map[string]bool{}

			for _, sender := range participant.peer.GetSenders() {
				if sender.Track() == nil {
					continue
				}

				existingSenders[sender.Track().ID()] = true

				// If we have a RTPSender that doesn't map to a existing track remove and signal
				if _, ok := s.trackLocals[sender.Track().ID()]; !ok {
					if err := participant.peer.RemoveTrack(sender); err != nil {
						return true
					}
				}
			}

			// Don't receive videos we are sending, make sure we don't have loopback
			for _, receiver := range participant.peer.GetReceivers() {
				if receiver.Track() == nil {
					continue
				}

				existingSenders[receiver.Track().ID()] = true
			}

			// Add all track we aren't sending yet to the PeerConnection
			for trackID := range s.trackLocals {
				log.Printf("Adding track: %s")
				if _, ok := existingSenders[trackID]; !ok {
					if _, err := participant.peer.AddTrack(s.trackLocals[trackID]); err != nil {
						return true
					}
				}
			}

			err := s.sendOffer(participant)
			if err != nil {
				return true
			}

			return false
		}
		return
	}

	for syncAttempt := 0; ; syncAttempt++ {
		if syncAttempt == 25 {
			// Release the lock and attempt a sync in 3 seconds. We might be blocking a RemoveTrack or AddTrack
			go func() {
				time.Sleep(time.Second * 3)
				s.syncPeers()
			}()
			return
		}

		if !attemptSync() {
			break
		}
	}

	return
}

func (s *SFU) addLocalTrack(t *webrtc.TrackRemote) *webrtc.TrackLocalStaticRTP {

	// Create a new TrackLocal with the same codec as our incoming
	trackLocal, err := webrtc.NewTrackLocalStaticRTP(t.Codec().RTPCodecCapability, t.ID(), t.StreamID())
	if err != nil {
		log.Printf("Failed to add track local: %s", err)
		return trackLocal
	}

	s.lock.Lock()
	s.trackLocals[t.ID()] = trackLocal
	s.lock.Unlock()

	// sync this new track with all current peers
	s.syncPeers()
	return trackLocal
}

func (s *SFU) removeLocalTrack(id string) {
	s.lock.Lock()
	delete(s.trackLocals, id)
	s.lock.Unlock()
	s.syncPeers()
}

func (s *SFU) newParticipantID() string {
	for {
		id := uuid.New().String()
		if _, ok := s.participants[id]; !ok {
			return id
		}
	}
}

func (s *SFU) removeParticipant(id string) {
	s.lock.Lock()
	delete(s.participants, id)
	s.lock.Unlock()
	s.syncPeers()
	return
}

func (s *SFU) sendOffer(participant *Participant) error {
	offer, err := participant.peer.CreateOffer(nil)
	if err != nil {
		return err
	}

	if err = participant.peer.SetLocalDescription(offer); err != nil {
		return err
	}

	offerByte, err := json.Marshal(offer)
	if err != nil {
		return err
	}
	payload := message.Wrapper{
		Type: message.TRTC,
		Data: message.RTC{
			Event: message.RTCOffer,
			Data:  string(offerByte),
		},
	}
	participant.client.Out <- payload
	return nil
}

func (s *SFU) Stop() {
	for _, participant := range s.participants {
		participant.peer.Close()
		participant.client.Close()
	}
}

func (s *SFU) sendKeyFrame() {
	s.lock.RLock()
	defer s.lock.RUnlock()

	for _, participant := range s.participants {
		for _, receiver := range participant.peer.GetReceivers() {
			if receiver.Track() == nil {
				continue
			}

			_ = participant.peer.WriteRTCP([]rtcp.Packet{
				&rtcp.PictureLossIndication{
					MediaSSRC: uint32(receiver.Track().SSRC()),
				},
			})
		}
	}
}
