package signaling

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/awwal/voxmeet/sfu/internal/peer"
	"github.com/awwal/voxmeet/sfu/internal/relay"
	"github.com/awwal/voxmeet/sfu/internal/room"
	"github.com/pion/webrtc/v4"
)

// MediaDispatcher extends the base Dispatcher with Pion WebRTC session management.
type MediaDispatcher struct {
	*Dispatcher
	iceServers []string
	sessions   map[string]*peer.Session // userID → active Pion session
	relays     map[string]*relay.Router // roomID → Router
	mu         sync.RWMutex
}

// NewMediaDispatcher creates a dispatcher that creates Pion sessions on join.
func NewMediaDispatcher(mgr *room.Manager, producer *Producer, iceServers []string) *MediaDispatcher {
	return &MediaDispatcher{
		Dispatcher: NewDispatcher(mgr, producer),
		iceServers: iceServers,
		sessions:   make(map[string]*peer.Session),
		relays:     make(map[string]*relay.Router),
	}
}

// Dispatch processes a signal, creating/closing Pion sessions as needed.
func (md *MediaDispatcher) Dispatch(ctx context.Context, msg Message) error {
	switch msg.Action {
	case ActionJoinRoom:
		return md.handleJoinRoom(ctx, msg)
	case ActionSDPAnswer:
		return md.handleSDPAnswer(ctx, msg)
	case ActionICECandidate:
		return md.handleICE(ctx, msg)
	case ActionLeaveRoom:
		return md.handleLeaveRoom(ctx, msg)
	case ActionPublish:
		return md.handlePublish(ctx, msg)
	case ActionUnpublish:
		return md.handleUnpublish(ctx, msg)
	default:
		return fmt.Errorf("unknown action: %s", msg.Action)
	}
}

// DispatchFunc returns a DispatchFunc that delegates to the media dispatcher.
func (md *MediaDispatcher) DispatchFunc() DispatchFunc {
	return func(ctx context.Context, msg Message) error {
		return md.Dispatch(ctx, msg)
	}
}

// CloseAll closes all Pion sessions.
func (md *MediaDispatcher) CloseAll() {
	md.mu.Lock()
	defer md.mu.Unlock()
	for id, s := range md.sessions {
		s.Close()
		delete(md.sessions, id)
	}
}

// --- Private helpers ---

func (md *MediaDispatcher) getOrCreateRouter(roomID string) *relay.Router {
	md.mu.Lock()
	defer md.mu.Unlock()
	if r, ok := md.relays[roomID]; ok {
		return r
	}
	r := relay.NewRouter()
	md.relays[roomID] = r
	return r
}

func (md *MediaDispatcher) handleJoinRoom(ctx context.Context, msg Message) error {
	r := md.manager.GetOrCreateRoom(msg.RoomID)
	r.AddPeer(msg.UserID)

	cfg := peer.Config{
		ICEServers: md.iceServers,
		OnTrack: func(remoteTrack *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
			kind := remoteTrack.Kind().String()
			trackID := remoteTrack.ID()

			r, ok := md.manager.GetRoom(msg.RoomID)
			if !ok {
				return
			}
			p, ok := r.GetPeer(msg.UserID)
			if !ok {
				return
			}
			p.AddTrack(trackID, kind, "")

			router := md.getOrCreateRouter(msg.RoomID)
			router.AddPublisherTrack(msg.UserID, trackID)

			for _, other := range r.GetPeers() {
				if other.UserID == msg.UserID {
					continue
				}
				router.Subscribe(other.UserID, trackID)

				md.mu.RLock()
				subSession, subOk := md.sessions[other.UserID]
				md.mu.RUnlock()
				if subOk {
					go md.forwardRTP(context.Background(), remoteTrack, subSession, trackID)
				}

				if md.producer != nil {
					data, _ := json.Marshal(NewTrackData{
						PublisherID: msg.UserID,
						TrackID:     trackID,
						Kind:        kind,
						Source:      "media",
					})
					md.producer.Publish(ctx, Message{
						Action: ActionNewTrack,
						RoomID: msg.RoomID,
						UserID: other.UserID,
						Data:   data,
					})
				}
			}
		},
		OnICE: func(candidate, sdpMid string, sdpMLineIndex uint16) {
			if md.producer == nil {
				return
			}
			data, _ := json.Marshal(ICEData{
				Candidate:     candidate,
				SDPMid:        sdpMid,
				SDPMLineIndex: sdpMLineIndex,
			})
			md.producer.Publish(ctx, Message{
				Action: ActionICECandidate,
				RoomID: msg.RoomID,
				UserID: msg.UserID,
				Data:   data,
			})
		},
		OnStateChange: func(state string) {},
	}

	s, err := peer.NewSession(cfg)
	if err != nil {
		return nil
	}
	offer, err := s.CreateOffer()
	if err != nil {
		s.Close()
		return nil
	}

	md.mu.Lock()
	md.sessions[msg.UserID] = s
	md.mu.Unlock()

	if md.producer != nil {
		data, _ := json.Marshal(SDPData{SDP: offer, Type: "offer"})
		md.producer.Publish(ctx, Message{
			Action: ActionSDPOffer, RoomID: msg.RoomID, UserID: msg.UserID, Data: data,
		})

		peers := r.GetPeers()
		peerInfos := make([]PeerInfo, 0, len(peers))
		for _, p := range peers {
			tracks := p.Tracks()
			trackInfos := make([]TrackInfo, 0, len(tracks))
			for _, t := range tracks {
				trackInfos = append(trackInfos, TrackInfo{TrackID: t.ID, Kind: t.Kind, Source: t.Source})
			}
			peerInfos = append(peerInfos, PeerInfo{UserID: p.UserID, DisplayName: p.UserID, Tracks: trackInfos})
		}
		stateData, _ := json.Marshal(RoomStateData{Peers: peerInfos})
		md.producer.Publish(ctx, Message{Action: ActionRoomState, RoomID: msg.RoomID, UserID: msg.UserID, Data: stateData})

		for _, p := range peers {
			if p.UserID == msg.UserID {
				continue
			}
			joinData, _ := json.Marshal(PeerJoinedData{UserID: msg.UserID, DisplayName: msg.UserID})
			md.producer.Publish(ctx, Message{Action: ActionPeerJoined, RoomID: msg.RoomID, UserID: p.UserID, Data: joinData})
		}
	}
	return nil
}

func (md *MediaDispatcher) handleSDPAnswer(ctx context.Context, msg Message) error {
	md.mu.RLock()
	s, ok := md.sessions[msg.UserID]
	md.mu.RUnlock()
	if !ok {
		return fmt.Errorf("no session for user %s", msg.UserID)
	}
	var data SDPData
	if err := json.Unmarshal(msg.Data, &data); err != nil {
		return fmt.Errorf("unmarshal sdp: %w", err)
	}
	return s.SetAnswer(data.SDP)
}

func (md *MediaDispatcher) handleICE(ctx context.Context, msg Message) error {
	md.mu.RLock()
	s, ok := md.sessions[msg.UserID]
	md.mu.RUnlock()
	if !ok {
		return fmt.Errorf("no session for user %s", msg.UserID)
	}
	var data ICEData
	if err := json.Unmarshal(msg.Data, &data); err != nil {
		return fmt.Errorf("unmarshal ice: %w", err)
	}
	return s.AddICECandidate(data.Candidate)
}

func (md *MediaDispatcher) handleLeaveRoom(ctx context.Context, msg Message) error {
	md.mu.Lock()
	if s, ok := md.sessions[msg.UserID]; ok {
		s.Close()
		delete(md.sessions, msg.UserID)
	}
	md.mu.Unlock()
	return md.Dispatcher.handleLeaveRoom(ctx, msg)
}

func (md *MediaDispatcher) handlePublish(ctx context.Context, msg Message) error {
	return md.Dispatcher.handlePublish(ctx, msg)
}

func (md *MediaDispatcher) handleUnpublish(ctx context.Context, msg Message) error {
	return md.Dispatcher.handleUnpublish(ctx, msg)
}

// forwardRTP reads RTP packets from a publisher and writes them to a subscriber.
func (md *MediaDispatcher) forwardRTP(ctx context.Context, remote *webrtc.TrackRemote, subSession *peer.Session, trackID string) {
	codec := remote.Codec().RTPCodecCapability
	localTrack, err := subSession.CreateDownTrack(codec, trackID, remote.StreamID())
	if err != nil {
		return
	}

	buf := make([]byte, 1500)
	for {
		n, _, err := remote.Read(buf)
		if err != nil {
			if err == io.EOF {
				return
			}
			return
		}
		if _, err := localTrack.Write(buf[:n]); err != nil {
			return
		}
	}
}
