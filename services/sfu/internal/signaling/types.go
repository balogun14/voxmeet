package signaling

import "encoding/json"

// Message is the envelope format for all signaling messages over RabbitMQ.
type Message struct {
	Action string          `json:"action"`
	RoomID string          `json:"room_id"`
	UserID string          `json:"user_id"`
	Data   json.RawMessage `json:"data"`
}

// Action constants
const (
	ActionJoinRoom     = "join_room"
	ActionSDPOffer     = "sdp_offer"
	ActionSDPAnswer    = "sdp_answer"
	ActionICECandidate = "ice_candidate"
	ActionPublish      = "publish"
	ActionUnpublish    = "unpublish"
	ActionNewTrack     = "new_track"
	ActionRemoveTrack  = "remove_track"
	ActionLeaveRoom    = "leave_room"
	ActionPeerJoined   = "peer_joined"
	ActionPeerLeft     = "peer_left"
	ActionRoomState    = "room_state"
	ActionError        = "error"
)

// JoinRoomData is the payload for a join_room request.
type JoinRoomData struct {
	DisplayName string `json:"display_name"`
}

// SDPData is the payload for SDP offer/answer messages.
type SDPData struct {
	SDP  string `json:"sdp"`
	Type string `json:"type"`
}

// ICEData is the payload for ICE candidate messages.
type ICEData struct {
	Candidate     string `json:"candidate"`
	SDPMid        string `json:"sdp_mid"`
	SDPMLineIndex uint16 `json:"sdp_mline_index"`
}

// PublishData is the payload when a client announces a track it wants to publish.
type PublishData struct {
	TrackID string `json:"track_id"`
	Kind    string `json:"kind"`
	Source  string `json:"source"`
}

// UnpublishData is the payload when a client stops publishing a track.
type UnpublishData struct {
	TrackID string `json:"track_id"`
}

// TrackInfo describes a media track in room state or new_track messages.
type TrackInfo struct {
	TrackID string `json:"track_id"`
	Kind    string `json:"kind"`
	Source  string `json:"source"`
}

// PeerInfo describes a peer in room state messages.
type PeerInfo struct {
	UserID      string      `json:"user_id"`
	DisplayName string      `json:"display_name"`
	Tracks      []TrackInfo `json:"tracks"`
}

// RoomStateData is the full room state sent to a newly joined peer.
type RoomStateData struct {
	Peers []PeerInfo `json:"peers"`
}

// NewTrackData is sent to subscribers when a new track is available.
type NewTrackData struct {
	PublisherID string `json:"publisher_id"`
	TrackID     string `json:"track_id"`
	Kind        string `json:"kind"`
	Source      string `json:"source"`
}

// RemoveTrackData is sent to subscribers when a track is removed.
type RemoveTrackData struct {
	TrackID string `json:"track_id"`
}

// PeerJoinedData is broadcast when a peer joins the room.
type PeerJoinedData struct {
	UserID      string      `json:"user_id"`
	DisplayName string      `json:"display_name"`
	Tracks      []TrackInfo `json:"tracks"`
}

// PeerLeftData is broadcast when a peer leaves the room.
type PeerLeftData struct {
	UserID string `json:"user_id"`
}

// ErrorData is sent when an error occurs.
type ErrorData struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
