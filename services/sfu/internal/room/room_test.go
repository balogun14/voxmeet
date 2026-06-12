package room_test

import (
	"testing"

	"github.com/awwal/voxmeet/sfu/internal/room"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Manager_GetOrCreateRoom(t *testing.T) {
	mgr := room.NewManager()
	defer mgr.StopAll()

	r1 := mgr.GetOrCreateRoom("room-1")
	require.NotNil(t, r1)
	assert.Equal(t, "room-1", r1.ID)

	r2 := mgr.GetOrCreateRoom("room-1")
	assert.Same(t, r1, r2, "should return the same instance")
}

func Test_Manager_GetRoom(t *testing.T) {
	mgr := room.NewManager()
	defer mgr.StopAll()

	_, exists := mgr.GetRoom("nonexistent")
	assert.False(t, exists)

	mgr.GetOrCreateRoom("room-abc")
	r, exists := mgr.GetRoom("room-abc")
	assert.True(t, exists)
	assert.Equal(t, "room-abc", r.ID)
}

func Test_Manager_RemoveRoom(t *testing.T) {
	mgr := room.NewManager()
	defer mgr.StopAll()

	mgr.GetOrCreateRoom("room-1")
	mgr.RemoveRoom("room-1")

	_, exists := mgr.GetRoom("room-1")
	assert.False(t, exists)
}

func Test_Room_AddAndRemovePeer(t *testing.T) {
	mgr := room.NewManager()
	defer mgr.StopAll()

	r := mgr.GetOrCreateRoom("room-1")

	peer := r.AddPeer("user-1")
	require.NotNil(t, peer)
	assert.Equal(t, "user-1", peer.UserID)
	assert.Equal(t, "room-1", peer.RoomID)
	assert.Equal(t, 1, r.PeerCount())

	r.RemovePeer("user-1")
	assert.Equal(t, 0, r.PeerCount())
}

func Test_Room_DuplicatePeer(t *testing.T) {
	mgr := room.NewManager()
	defer mgr.StopAll()

	r := mgr.GetOrCreateRoom("room-1")

	p1 := r.AddPeer("user-1")
	p2 := r.AddPeer("user-1")
	assert.Same(t, p1, p2, "should return existing peer for duplicate")
	assert.Equal(t, 1, r.PeerCount())
}

func Test_Room_PeerTracks(t *testing.T) {
	mgr := room.NewManager()
	defer mgr.StopAll()

	r := mgr.GetOrCreateRoom("room-1")
	peer := r.AddPeer("user-1")

	track := peer.AddTrack("track-1", "audio", "mic")
	require.NotNil(t, track)
	assert.Equal(t, "track-1", track.ID)
	assert.Equal(t, "audio", track.Kind)
	assert.Equal(t, "mic", track.Source)

	// Duplicate
	sameTrack := peer.AddTrack("track-1", "audio", "mic")
	assert.Same(t, track, sameTrack)

	// List
	tracks := peer.Tracks()
	assert.Len(t, tracks, 1)

	// Remove
	peer.RemoveTrack("track-1")
	assert.Len(t, peer.Tracks(), 0)
}

func Test_Room_GetPeers(t *testing.T) {
	mgr := room.NewManager()
	defer mgr.StopAll()

	r := mgr.GetOrCreateRoom("room-1")
	r.AddPeer("user-1")
	r.AddPeer("user-2")

	peers := r.GetPeers()
	assert.Len(t, peers, 2)
}

func Test_Manager_RoomCount(t *testing.T) {
	mgr := room.NewManager()
	defer mgr.StopAll()

	assert.Equal(t, 0, mgr.RoomCount())

	mgr.GetOrCreateRoom("room-1")
	assert.Equal(t, 1, mgr.RoomCount())

	mgr.GetOrCreateRoom("room-2")
	assert.Equal(t, 2, mgr.RoomCount())

	mgr.RemoveRoom("room-1")
	assert.Equal(t, 1, mgr.RoomCount())
}

func Test_Room_GetPeer(t *testing.T) {
	mgr := room.NewManager()
	defer mgr.StopAll()

	r := mgr.GetOrCreateRoom("room-1")
	r.AddPeer("user-1")

	p, exists := r.GetPeer("user-1")
	assert.True(t, exists)
	assert.Equal(t, "user-1", p.UserID)

	_, exists = r.GetPeer("nonexistent")
	assert.False(t, exists)
}

func Test_Room_Stop(t *testing.T) {
	mgr := room.NewManager()
	defer mgr.StopAll()

	r := mgr.GetOrCreateRoom("room-1")
	r.AddPeer("user-1")
	r.AddPeer("user-2")

	r.Stop()

	// After stop, room should be empty
	assert.Equal(t, 0, r.PeerCount())
}
