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
	"github.com/pion/webrtc/v3"
	"github.com/qnkhuat/tstream/pkg/message"
	"log"
	"sync"
)

type participant struct {
	peer   *webrtc.PeerConnection
	client *Client // contain role and websocket connection
}

type SFU struct {
	lock        sync.RWMutex
	trackLocals map[string]*webrtc.TrackLocalStaticRTP
	consumers   map[string]*participant
	producers   map[string]*participant
}

func NewSFU() *SFU {
	trackLocals := map[string]*webrtc.TrackLocalStaticRTP{}
	consumers := map[string]*participant{}
	producers := map[string]*participant{}

	return &SFU{
		trackLocals: trackLocals,
		consumers:   consumers,
		producers:   producers,
	}
}

func (s *SFU) AddProducer(cl *Client) error {
	peerConn, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		log.Printf("Failed to init peer connection: %s", err)
		return err
	}
	defer peerConn.Close()

	// Accept one audio and one video track incoming
	for _, typ := range []webrtc.RTPCodecType{webrtc.RTPCodecTypeVideo, webrtc.RTPCodecTypeAudio} {
		if _, err := peerConn.AddTransceiverFromKind(typ, webrtc.RTPTransceiverInit{
			Direction: webrtc.RTPTransceiverDirectionRecvonly,
		}); err != nil {
			log.Printf("Failed to add transeiver: %s", err)
			return err
		}
	}

	log.Printf("About to add producer")

	//s.lock.Lock()
	producerID := s.newProducerID()
	s.producers[producerID] = &participant{
		peer:   peerConn,
		client: cl,
	}
	//s.lock.Unlock()

	// Trickle ICE. Emit server candidate to client
	peerConn.OnICECandidate(func(i *webrtc.ICECandidate) {
		if i == nil {
			return
		}

		candidate, err := json.Marshal(i.ToJSON())
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
		log.Printf("Sent candidate")

		cl.Out <- payload
	})

	// If PeerConnection is closed remove it from global list
	peerConn.OnConnectionStateChange(func(p webrtc.PeerConnectionState) {
		switch p {
		case webrtc.PeerConnectionStateFailed:
			if err := peerConn.Close(); err != nil {
				log.Print(err)
			}

		case webrtc.PeerConnectionStateClosed:
			s.removeProducer(producerID)

		default:
			log.Printf("Not implemented: %s", p)
		}
	})

	peerConn.OnTrack(func(t *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		log.Printf("Got a new track: %s", t.Kind())
		// Create a track to fan out our incoming video to all peers
		trackLocal := s.addTrack(t)
		defer s.removeTrack(t.ID())

		buf := make([]byte, 1500)
		for {
			i, _, err := t.Read(buf)
			log.Printf("(Track: %s) Reading: %d", trackLocal.Kind(), len(buf))
			if err != nil {
				return
			}

			if _, err = trackLocal.Write(buf[:i]); err != nil {
				return
			}
		}
	})

	// send offer
	offer, err := peerConn.CreateOffer(nil)
	if err != nil {
		return err
	}

	if err = peerConn.SetLocalDescription(offer); err != nil {
		return err
	}

	offerByte, err := json.Marshal(offer)
	if err != nil {
		return err
	}

	offerMsg := message.RTC{
		Event: message.RTCOffer,
		Data:  string(offerByte),
	}
	payload := message.Wrapper{
		Type: message.TRTC,
		Data: offerMsg,
	}
	cl.Out <- payload
	log.Printf("Offer sent")

	for {

		msg := <-cl.In
		if msg.Type != message.TRTC {
			log.Printf("not RTC message: %s", msg.Type)
			continue
		}

		rtcMsg := message.RTC{}
		if err := message.ToStruct(msg.Data, &rtcMsg); err != nil {
			log.Printf("Failed to decode RTC message: %v", msg.Data)
			continue
		}

		log.Printf("Got RTC Event: %s", rtcMsg.Event)
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
		}

	}

	return nil
}

func (s *SFU) AddConsumer(cl *Client) error {
	peerConn, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		log.Printf("Failed to init peer connection: %s", err)
		return err
	}
	defer peerConn.Close()

	log.Printf("About to add consumer")

	//s.lock.Lock()
	ID := s.newConsumerID()
	s.producers[ID] = &participant{
		peer:   peerConn,
		client: cl,
	}
	//s.lock.Unlock()
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
	peerConn.OnICECandidate(func(i *webrtc.ICECandidate) {
		if i == nil {
			return
		}

		candidate, err := json.Marshal(i.ToJSON())
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
		log.Printf("Sent candidate")

		cl.Out <- payload
	})

	// If PeerConnection is closed remove it from global list
	peerConn.OnConnectionStateChange(func(p webrtc.PeerConnectionState) {
		switch p {
		case webrtc.PeerConnectionStateFailed:
			if err := peerConn.Close(); err != nil {
				log.Print(err)
			}

		case webrtc.PeerConnectionStateClosed:
		//c.removeProducer(producerID)

		default:
			log.Printf("Not implemented: %s", p)
		}
	})

	peerConn.OnTrack(func(t *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		log.Printf("Got a new track from consumer: %s", t.Kind())
		log.Printf("Consumer tried to addtrack. Nice try")
		//// Create a track to fan out our incoming video to all peers
		//trackLocal := s.addTrack(t)
		//defer s.removeTrack(t.ID())

		//buf := make([]byte, 1500)
		//for {
		//	i, _, err := t.Read(buf)
		//	log.Printf("(Track: %s) Consumer Reading: %d", trackLocal.Kind(), len(buf))
		//	if err != nil {
		//		return
		//	}

		//	if _, err = trackLocal.Write(buf[:i]); err != nil {
		//		return
		//	}
		//}
	})

	for _, track := range s.trackLocals {
		if _, err := peerConn.AddTrack(track); err != nil {
			log.Printf("Failed to add track")
		}
	}

	// send offer
	offer, err := peerConn.CreateOffer(nil)
	if err != nil {
		return err
	}

	if err = peerConn.SetLocalDescription(offer); err != nil {
		return err
	}

	offerByte, err := json.Marshal(offer)
	if err != nil {
		return err
	}

	offerMsg := message.RTC{
		Event: message.RTCOffer,
		Data:  string(offerByte),
	}
	payload := message.Wrapper{
		Type: message.TRTC,
		Data: offerMsg,
	}
	cl.Out <- payload
	log.Printf("Offer sent")

	for {

		msg := <-cl.In
		if msg.Type != message.TRTC {
			log.Printf("not RTC message: %s", msg.Type)
			continue
		}

		rtcMsg := message.RTC{}
		if err := message.ToStruct(msg.Data, &rtcMsg); err != nil {
			log.Printf("Failed to decode RTC message: %v", msg.Data)
			continue
		}

		log.Printf("Got RTC Event: %s", rtcMsg.Event)
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
		}

	}

	return nil
}

func (s *SFU) addTrack(t *webrtc.TrackRemote) *webrtc.TrackLocalStaticRTP {
	s.lock.Lock()

	// Create a new TrackLocal with the same codec as our incoming
	trackLocal, err := webrtc.NewTrackLocalStaticRTP(t.Codec().RTPCodecCapability, t.ID(), t.StreamID())
	if err != nil {
		panic(err)
	}

	s.trackLocals[t.ID()] = trackLocal

	return trackLocal
}

func (s *SFU) removeTrack(id string) {
	s.lock.Lock()
	delete(s.trackLocals, id)
	//s.updateProducers()
	s.lock.Unlock()
}

func (s *SFU) newConsumerID() string {
	var id string
	for {
		id = uuid.New().String()
		if _, ok := s.consumers[id]; !ok {
			return id
		}
	}
}

func (s *SFU) newProducerID() string {
	var id string
	for {
		id = uuid.New().String()
		if _, ok := s.consumers[id]; !ok {
			return id
		}
	}
}

func (s *SFU) removeProducer(id string) {
	s.lock.Lock()
	delete(s.producers, id)
	s.lock.Unlock()
	return
}
