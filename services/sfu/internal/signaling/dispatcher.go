package signaling

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/awwal/voxmeet/sfu/internal/room"
)

// Dispatcher handles incoming signaling messages and manages room/peer state.
type Dispatcher struct {
	manager  *room.Manager
	producer *Producer
}

// NewDispatcher creates a new Dispatcher.
func NewDispatcher(mgr *room.Manager, producer *Producer) *Dispatcher {
	return &Dispatcher{
		manager:  mgr,
		producer: producer,
	}
}

// DispatchFunc returns a signal.DispatchFunc that delegates to the dispatcher.
func (d *Dispatcher) DispatchFunc() DispatchFunc {
	return func(ctx context.Context, msg Message) error {
		return d.Dispatch(ctx, msg)
	}
}

// Dispatch processes a single signaling message.
func (d *Dispatcher) Dispatch(ctx context.Context, msg Message) error {
	if d.manager == nil {
		return nil
	}

	switch msg.Action {
	case ActionJoinRoom:
		return d.handleJoinRoom(ctx, msg)
	case ActionLeaveRoom:
		return d.handleLeaveRoom(ctx, msg)
	case ActionPublish:
		return d.handlePublish(ctx, msg)
	case ActionUnpublish:
		return d.handleUnpublish(ctx, msg)
	case ActionSDPAnswer:
		return d.handleSDPAnswer(ctx, msg)
	case ActionICECandidate:
		return d.handleICE(ctx, msg)
	default:
		return fmt.Errorf("unknown action: %s", msg.Action)
	}
}

func (d *Dispatcher) handleJoinRoom(ctx context.Context, msg Message) error {
	r := d.manager.GetOrCreateRoom(msg.RoomID)
	r.AddPeer(msg.UserID)

	// If we have a producer, send room_state back
	if d.producer != nil {
		peers := r.GetPeers()
		peerInfos := make([]PeerInfo, 0, len(peers))
		for _, p := range peers {
			tracks := p.Tracks()
			trackInfos := make([]TrackInfo, 0, len(tracks))
			for _, t := range tracks {
				trackInfos = append(trackInfos, TrackInfo{
					TrackID: t.ID,
					Kind:    t.Kind,
					Source:  t.Source,
				})
			}
			peerInfos = append(peerInfos, PeerInfo{
				UserID:      p.UserID,
				DisplayName: p.UserID,
				Tracks:      trackInfos,
			})
		}

		data, _ := json.Marshal(RoomStateData{Peers: peerInfos})
		d.producer.Publish(ctx, Message{
			Action: ActionRoomState,
			RoomID: msg.RoomID,
			UserID: msg.UserID,
			Data:   data,
		})

		// Notify other peers
		for _, p := range peers {
			if p.UserID == msg.UserID {
				continue
			}
			joinData, _ := json.Marshal(PeerJoinedData{
				UserID:      msg.UserID,
				DisplayName: msg.UserID,
			})
			d.producer.Publish(ctx, Message{
				Action: ActionPeerJoined,
				RoomID: msg.RoomID,
				UserID: p.UserID,
				Data:   joinData,
			})
		}
	}

	return nil
}

func (d *Dispatcher) handleLeaveRoom(ctx context.Context, msg Message) error {
	r, exists := d.manager.GetRoom(msg.RoomID)
	if !exists {
		return nil
	}

	r.RemovePeer(msg.UserID)

	// Notify remaining peers
	if d.producer != nil {
		for _, p := range r.GetPeers() {
			data, _ := json.Marshal(PeerLeftData{UserID: msg.UserID})
			d.producer.Publish(ctx, Message{
				Action: ActionPeerLeft,
				RoomID: msg.RoomID,
				UserID: p.UserID,
				Data:   data,
			})
		}
	}

	// Clean up empty rooms
	if r.PeerCount() == 0 {
		d.manager.RemoveRoom(msg.RoomID)
	}

	return nil
}

func (d *Dispatcher) handlePublish(ctx context.Context, msg Message) error {
	r, exists := d.manager.GetRoom(msg.RoomID)
	if !exists {
		return fmt.Errorf("room not found: %s", msg.RoomID)
	}

	peer, exists := r.GetPeer(msg.UserID)
	if !exists {
		return fmt.Errorf("peer not found: %s", msg.UserID)
	}

	var pubData PublishData
	if err := json.Unmarshal(msg.Data, &pubData); err != nil {
		return fmt.Errorf("unmarshal publish data: %w", err)
	}

	peer.AddTrack(pubData.TrackID, pubData.Kind, pubData.Source)

	// Notify other peers about the new track
	if d.producer != nil {
		for _, p := range r.GetPeers() {
			if p.UserID == msg.UserID {
				continue
			}
			data, _ := json.Marshal(NewTrackData{
				PublisherID: msg.UserID,
				TrackID:     pubData.TrackID,
				Kind:        pubData.Kind,
				Source:      pubData.Source,
			})
			d.producer.Publish(ctx, Message{
				Action: ActionNewTrack,
				RoomID: msg.RoomID,
				UserID: p.UserID,
				Data:   data,
			})
		}
	}

	return nil
}

func (d *Dispatcher) handleUnpublish(ctx context.Context, msg Message) error {
	r, exists := d.manager.GetRoom(msg.RoomID)
	if !exists {
		return fmt.Errorf("room not found: %s", msg.RoomID)
	}

	peer, exists := r.GetPeer(msg.UserID)
	if !exists {
		return fmt.Errorf("peer not found: %s", msg.UserID)
	}

	var unpubData UnpublishData
	if err := json.Unmarshal(msg.Data, &unpubData); err != nil {
		return fmt.Errorf("unmarshal unpublish data: %w", err)
	}

	peer.RemoveTrack(unpubData.TrackID)

	// Notify other peers
	if d.producer != nil {
		for _, p := range r.GetPeers() {
			if p.UserID == msg.UserID {
				continue
			}
			data, _ := json.Marshal(RemoveTrackData{TrackID: unpubData.TrackID})
			d.producer.Publish(ctx, Message{
				Action: ActionRemoveTrack,
				RoomID: msg.RoomID,
				UserID: p.UserID,
				Data:   data,
			})
		}
	}

	return nil
}

func (d *Dispatcher) handleSDPAnswer(ctx context.Context, msg Message) error {
	// SDP handling will be implemented when Pion PeerConnection is wired in
	return nil
}

func (d *Dispatcher) handleICE(ctx context.Context, msg Message) error {
	// ICE handling will be implemented when Pion PeerConnection is wired in
	return nil
}
