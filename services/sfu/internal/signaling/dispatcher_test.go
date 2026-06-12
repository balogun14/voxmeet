package signaling_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/awwal/voxmeet/sfu/internal/room"
	"github.com/awwal/voxmeet/sfu/internal/signaling"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Dispatch_JoinRoom_CreatesPeer(t *testing.T) {
	mgr := room.NewManager()
	defer mgr.StopAll()

	disp := signaling.NewDispatcher(mgr, nil)
	require.NotNil(t, disp)

	data, _ := json.Marshal(signaling.JoinRoomData{DisplayName: "Alice"})
	msg := signaling.Message{
		Action: signaling.ActionJoinRoom,
		RoomID: "room-1",
		UserID: "user-1",
		Data:   data,
	}

	err := disp.Dispatch(context.Background(), msg)
	require.NoError(t, err)

	r, exists := mgr.GetRoom("room-1")
	require.True(t, exists)
	assert.Equal(t, 1, r.PeerCount())

	_, exists = r.GetPeer("user-1")
	assert.True(t, exists)
}

func Test_Dispatch_JoinRoom_Twice(t *testing.T) {
	mgr := room.NewManager()
	defer mgr.StopAll()

	disp := signaling.NewDispatcher(mgr, nil)

	data, _ := json.Marshal(signaling.JoinRoomData{DisplayName: "Alice"})
	msg := signaling.Message{
		Action: signaling.ActionJoinRoom,
		RoomID: "room-1",
		UserID: "user-1",
		Data:   data,
	}

	err := disp.Dispatch(context.Background(), msg)
	require.NoError(t, err)

	err = disp.Dispatch(context.Background(), msg)
	require.NoError(t, err)

	r, _ := mgr.GetRoom("room-1")
	assert.Equal(t, 1, r.PeerCount())
}

func Test_Dispatch_LeaveRoom(t *testing.T) {
	mgr := room.NewManager()
	defer mgr.StopAll()

	disp := signaling.NewDispatcher(mgr, nil)

	// Join
	joinData, _ := json.Marshal(signaling.JoinRoomData{DisplayName: "Alice"})
	disp.Dispatch(context.Background(), signaling.Message{
		Action: signaling.ActionJoinRoom,
		RoomID: "room-1",
		UserID: "user-1",
		Data:   joinData,
	})
	disp.Dispatch(context.Background(), signaling.Message{
		Action: signaling.ActionJoinRoom,
		RoomID: "room-1",
		UserID: "user-2",
		Data:   joinData,
	})

	// Leave
	leaveData, _ := json.Marshal(struct{}{})
	err := disp.Dispatch(context.Background(), signaling.Message{
		Action: signaling.ActionLeaveRoom,
		RoomID: "room-1",
		UserID: "user-1",
		Data:   leaveData,
	})
	require.NoError(t, err)

	r, _ := mgr.GetRoom("room-1")
	assert.Equal(t, 1, r.PeerCount())

	_, exists := r.GetPeer("user-1")
	assert.False(t, exists)

	_, exists = r.GetPeer("user-2")
	assert.True(t, exists)
}

func Test_Dispatch_LeaveRoom_RemovesEmptyRoom(t *testing.T) {
	mgr := room.NewManager()
	defer mgr.StopAll()

	disp := signaling.NewDispatcher(mgr, nil)

	joinData, _ := json.Marshal(signaling.JoinRoomData{DisplayName: "Alice"})
	disp.Dispatch(context.Background(), signaling.Message{
		Action: signaling.ActionJoinRoom,
		RoomID: "room-1",
		UserID: "user-1",
		Data:   joinData,
	})

	emptyData, _ := json.Marshal(struct{}{})
	disp.Dispatch(context.Background(), signaling.Message{
		Action: signaling.ActionLeaveRoom,
		RoomID: "room-1",
		UserID: "user-1",
		Data:   emptyData,
	})

	_, exists := mgr.GetRoom("room-1")
	assert.False(t, exists, "room should be removed when last peer leaves")
}

func Test_Dispatch_PublishTrack(t *testing.T) {
	mgr := room.NewManager()
	defer mgr.StopAll()

	disp := signaling.NewDispatcher(mgr, nil)

	joinData, _ := json.Marshal(signaling.JoinRoomData{DisplayName: "Alice"})
	disp.Dispatch(context.Background(), signaling.Message{
		Action: signaling.ActionJoinRoom,
		RoomID: "room-1",
		UserID: "user-1",
		Data:   joinData,
	})

	pubData, _ := json.Marshal(signaling.PublishData{
		TrackID: "track-1",
		Kind:    "audio",
		Source:  "mic",
	})
	err := disp.Dispatch(context.Background(), signaling.Message{
		Action: signaling.ActionPublish,
		RoomID: "room-1",
		UserID: "user-1",
		Data:   pubData,
	})
	require.NoError(t, err)

	r, _ := mgr.GetRoom("room-1")
	p, _ := r.GetPeer("user-1")
	assert.Len(t, p.Tracks(), 1)
	assert.Equal(t, "audio", p.Tracks()[0].Kind)
}

func Test_Dispatch_UnpublishTrack(t *testing.T) {
	mgr := room.NewManager()
	defer mgr.StopAll()

	disp := signaling.NewDispatcher(mgr, nil)

	joinData, _ := json.Marshal(signaling.JoinRoomData{DisplayName: "Alice"})
	disp.Dispatch(context.Background(), signaling.Message{
		Action: signaling.ActionJoinRoom,
		RoomID: "room-1",
		UserID: "user-1",
		Data:   joinData,
	})

	pubData, _ := json.Marshal(signaling.PublishData{
		TrackID: "track-1",
		Kind:    "audio",
		Source:  "mic",
	})
	disp.Dispatch(context.Background(), signaling.Message{
		Action: signaling.ActionPublish,
		RoomID: "room-1",
		UserID: "user-1",
		Data:   pubData,
	})

	unpubData, _ := json.Marshal(signaling.UnpublishData{TrackID: "track-1"})
	err := disp.Dispatch(context.Background(), signaling.Message{
		Action: signaling.ActionUnpublish,
		RoomID: "room-1",
		UserID: "user-1",
		Data:   unpubData,
	})
	require.NoError(t, err)

	r, _ := mgr.GetRoom("room-1")
	p, _ := r.GetPeer("user-1")
	assert.Len(t, p.Tracks(), 0)
}

func Test_Dispatch_UnknownAction(t *testing.T) {
	mgr := room.NewManager()
	defer mgr.StopAll()

	disp := signaling.NewDispatcher(mgr, nil)

	err := disp.Dispatch(context.Background(), signaling.Message{
		Action: "unknown_action",
	})
	assert.Error(t, err)
}

func Test_Dispatch_NilDispatcher(t *testing.T) {
	disp := signaling.NewDispatcher(nil, nil)
	assert.NotNil(t, disp)
}
