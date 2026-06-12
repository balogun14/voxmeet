package signaling_test

import (
	"encoding/json"
	"testing"

	"github.com/awwal/voxmeet/sfu/internal/signaling"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_SignalMessage_Marshal(t *testing.T) {
	msg := signaling.Message{
		Action: signaling.ActionJoinRoom,
		RoomID: "room-abc",
		UserID: "user-123",
		Data:   json.RawMessage(`{"display_name":"Alice"}`),
	}

	data, err := json.Marshal(msg)
	require.NoError(t, err)

	var decoded signaling.Message
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, signaling.ActionJoinRoom, decoded.Action)
	assert.Equal(t, "room-abc", decoded.RoomID)
	assert.Equal(t, "user-123", decoded.UserID)
}

func Test_SignalMessage_JoinRoomData(t *testing.T) {
	data := signaling.JoinRoomData{DisplayName: "Alice"}
	raw, err := json.Marshal(data)
	require.NoError(t, err)

	msg := signaling.Message{
		Action: signaling.ActionJoinRoom,
		Data:   raw,
	}

	var decoded signaling.JoinRoomData
	err = json.Unmarshal(msg.Data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, "Alice", decoded.DisplayName)
}

func Test_SignalMessage_SDPData(t *testing.T) {
	data := signaling.SDPData{SDP: "v=0\r\no=- ...", Type: "offer"}
	raw, err := json.Marshal(data)
	require.NoError(t, err)

	msg := signaling.Message{
		Action: signaling.ActionSDPOffer,
		Data:   raw,
	}

	var decoded signaling.SDPData
	err = json.Unmarshal(msg.Data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, "v=0\r\no=- ...", decoded.SDP)
	assert.Equal(t, "offer", decoded.Type)
}

func Test_SignalMessage_ICEData(t *testing.T) {
	data := signaling.ICEData{
		Candidate:     "candidate:1 1 UDP 2122252543 192.168.1.1 54321 typ host",
		SDPMid:        "0",
		SDPMLineIndex: 0,
	}
	raw, err := json.Marshal(data)
	require.NoError(t, err)

	var decoded signaling.ICEData
	err = json.Unmarshal(raw, &decoded)
	require.NoError(t, err)
	assert.Equal(t, "candidate:1 1 UDP 2122252543 192.168.1.1 54321 typ host", decoded.Candidate)
	assert.Equal(t, "0", decoded.SDPMid)
	assert.EqualValues(t, 0, decoded.SDPMLineIndex)
}

func Test_SignalMessage_PublishData(t *testing.T) {
	data := signaling.PublishData{TrackID: "track-1", Kind: "audio", Source: "mic"}
	raw, err := json.Marshal(data)
	require.NoError(t, err)

	var decoded signaling.PublishData
	err = json.Unmarshal(raw, &decoded)
	require.NoError(t, err)
	assert.Equal(t, "track-1", decoded.TrackID)
	assert.Equal(t, "audio", decoded.Kind)
	assert.Equal(t, "mic", decoded.Source)
}

func Test_ActionConstants(t *testing.T) {
	assert.Equal(t, "join_room", signaling.ActionJoinRoom)
	assert.Equal(t, "sdp_offer", signaling.ActionSDPOffer)
	assert.Equal(t, "sdp_answer", signaling.ActionSDPAnswer)
	assert.Equal(t, "ice_candidate", signaling.ActionICECandidate)
	assert.Equal(t, "publish", signaling.ActionPublish)
	assert.Equal(t, "unpublish", signaling.ActionUnpublish)
	assert.Equal(t, "new_track", signaling.ActionNewTrack)
	assert.Equal(t, "remove_track", signaling.ActionRemoveTrack)
	assert.Equal(t, "leave_room", signaling.ActionLeaveRoom)
	assert.Equal(t, "peer_joined", signaling.ActionPeerJoined)
	assert.Equal(t, "peer_left", signaling.ActionPeerLeft)
	assert.Equal(t, "room_state", signaling.ActionRoomState)
	assert.Equal(t, "error", signaling.ActionError)
}

func Test_SignalMessage_RoomStateData(t *testing.T) {
	peers := []signaling.PeerInfo{
		{
			UserID:      "user-1",
			DisplayName: "Alice",
			Tracks: []signaling.TrackInfo{
				{TrackID: "t1", Kind: "audio", Source: "mic"},
			},
		},
		{
			UserID:      "user-2",
			DisplayName: "Bob",
			Tracks: []signaling.TrackInfo{
				{TrackID: "t2", Kind: "video", Source: "camera"},
			},
		},
	}

	data := signaling.RoomStateData{Peers: peers}
	raw, err := json.Marshal(data)
	require.NoError(t, err)

	var decoded signaling.RoomStateData
	err = json.Unmarshal(raw, &decoded)
	require.NoError(t, err)
	assert.Len(t, decoded.Peers, 2)
	assert.Equal(t, "user-1", decoded.Peers[0].UserID)
	assert.Equal(t, "Alice", decoded.Peers[0].DisplayName)
	assert.Len(t, decoded.Peers[0].Tracks, 1)
	assert.Equal(t, "audio", decoded.Peers[0].Tracks[0].Kind)
}

func Test_SignalMessage_NewTrackData(t *testing.T) {
	data := signaling.NewTrackData{
		PublisherID: "user-1",
		TrackID:     "t1",
		Kind:        "video",
		Source:      "camera",
	}
	raw, err := json.Marshal(data)
	require.NoError(t, err)

	var decoded signaling.NewTrackData
	err = json.Unmarshal(raw, &decoded)
	require.NoError(t, err)
	assert.Equal(t, "user-1", decoded.PublisherID)
	assert.Equal(t, "t1", decoded.TrackID)
}
