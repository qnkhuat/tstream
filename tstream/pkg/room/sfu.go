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
	// TODO: this should be requested by client, not server auto send it every 3 seconds
	// request a keyframe every 3 seconds
	//go func() {
	//	for range time.NewTicker(time.Second * 3).C {
	//		s.sendKeyFrame()
	//	}
	//}()
}

// TODO : break down this method
func (s *SFU) AddPeer(cl *Client) error {

	peerConn, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		log.Printf("Failed to init peer connection: %s", err)
		return err
	}
	defer peerConn.Close()
	defer cl.Close()

	participantID := s.newParticipantID()
	participant := &Participant{
		peer:   peerConn,
		client: cl,
	}

	s.lock.Lock()
	s.participants[participantID] = participant
	s.lock.Unlock()

	// Accept one audio and one video track incoming
	for _, typ := range []webrtc.RTPCodecType{webrtc.RTPCodecTypeVideo, webrtc.RTPCodecTypeAudio} {
		if _, err := peerConn.AddTransceiverFromKind(typ, webrtc.RTPTransceiverInit{
			Direction: webrtc.RTPTransceiverDirectionRecvonly,
		}); err != nil {
			log.Printf("Failed to add transeiver: %s", err)
			return err
		}
	}

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

	peerConn.OnConnectionStateChange(func(p webrtc.PeerConnectionState) {
		log.Printf("Pariticipant: %s changed stated to: %s", participantID, p)
		switch p {

		case webrtc.PeerConnectionStateFailed, webrtc.PeerConnectionStateClosed, webrtc.PeerConnectionStateDisconnected:
			s.removeParticipant(participantID)

		case webrtc.PeerConnectionStateConnected:
			// nothing yet

		default:
			log.Printf("Not implemented: %s", p)
		}

	})

	// Add all current tracks to this peer
	for _, track := range s.trackLocals {
		if _, err := peerConn.AddTrack(track); err != nil {
			log.Printf("Failed to add track: %s", err)
		}
	}

	// only producer can broadcast
	if cl.Role() == message.RProducerRTC {
		peerConn.OnTrack(func(t *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
			// Create a track to fan out our incoming video to all peerse
			log.Printf("added track :%s, %s", t.Kind(), t.ID())
			trackLocal := s.addLocalTrack(t)
			defer s.removeLocalTrack(t.ID())

			buf := make([]byte, 1500)
			for {
				// remote from remote
				i, _, err := t.Read(buf)
				if err != nil {
					log.Printf("Failed to read from track: %s", err)
					return
				}

				// send to all peers
				if _, err = trackLocal.Write(buf[:i]); err != nil {
					log.Printf("Failed to write to track local: %s", err)
					return
				}
			}
		})

	}

	// Signaling starts
	// 1. Server send offer to the other peer connection
	// 2. Server get answer
	// 3. Server send ice candidate
	// 4. Peer connection is established
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

	s.removeParticipant(participantID)
	return nil
}

// TODO: add filter so that only sending offer for who need to update
// Sync tracklocals with all peers so everyone get the right track at the right time
func (s *SFU) syncPeers() {

	attemptSync := func() (tryAgain bool) {
		s.lock.Lock()
		defer s.lock.Unlock()
		// if there is only one person in the room, no need to sync
		if len(s.participants) < 2 {
			return false
		}

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
				if _, ok := existingSenders[trackID]; !ok {
					if _, err := participant.peer.AddTrack(s.trackLocals[trackID]); err != nil {
						return true
					}
				}
			}

			err := s.sendOffer(participant)
			if err != nil {
				log.Printf("Failed to send offer :%s", err)
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
	if _, ok := s.participants[id]; !ok {
		return
	}

	// receiver in this context is the track producer send to server
	// if pariticipant is a producer => remove their track
	for _, receiver := range s.participants[id].peer.GetReceivers() {
		if receiver.Track() == nil {
			continue
		}
		s.removeLocalTrack(receiver.Track().ID())
	}

	s.lock.Lock()
	delete(s.participants, id)
	s.lock.Unlock()
	s.syncPeers()
	return
}

func (s *SFU) sendOffer(participant *Participant) error {
	offer, err := participant.peer.CreateOffer(nil)
	if err != nil {
		log.Printf("failed to create offer: %s", err)
		return err
	}

	if err = participant.peer.SetLocalDescription(offer); err != nil {
		log.Printf("Failed to set local description: %s", err)
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

// used for video broadcasting
// without sending keyframe user will receive a crappy video until the next keyframe is sent
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
