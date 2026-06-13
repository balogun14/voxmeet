package presence_test

import (
	"testing"
	"time"

	"github.com/awwal/voxmeet/presence-service/internal/presence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Tracker_UserOnline(t *testing.T) {
	tr := presence.NewTracker(100 * time.Millisecond)
	defer tr.Stop()

	tr.UserOnline("user-1", "room-1")
	assert.True(t, tr.IsOnline("user-1"))
}

func Test_Tracker_UserOffline(t *testing.T) {
	tr := presence.NewTracker(100 * time.Millisecond)
	defer tr.Stop()

	tr.UserOnline("user-1", "room-1")
	tr.UserOffline("user-1")
	assert.False(t, tr.IsOnline("user-1"))
}

func Test_Tracker_GetRoomUsers(t *testing.T) {
	tr := presence.NewTracker(100 * time.Millisecond)
	defer tr.Stop()

	tr.UserOnline("user-1", "room-1")
	tr.UserOnline("user-2", "room-1")
	tr.UserOnline("user-3", "room-2")

	users := tr.GetRoomUsers("room-1")
	assert.Len(t, users, 2)

	users = tr.GetRoomUsers("room-2")
	assert.Len(t, users, 1)

	users = tr.GetRoomUsers("nonexistent")
	assert.Len(t, users, 0)
}

func Test_Tracker_Heartbeat(t *testing.T) {
	timeout := 200 * time.Millisecond
	tr := presence.NewTracker(timeout)
	defer tr.Stop()

	tr.UserOnline("user-1", "room-1")
	assert.True(t, tr.IsOnline("user-1"))

	// Heartbeat within timeout
	tr.Heartbeat("user-1")
	assert.True(t, tr.IsOnline("user-1"))

	// Wait for timeout
	time.Sleep(timeout + 50*time.Millisecond)
	assert.False(t, tr.IsOnline("user-1"), "user should be timed out")
}

func Test_Tracker_HeartbeatExtendsLife(t *testing.T) {
	timeout := 150 * time.Millisecond
	tr := presence.NewTracker(timeout)
	defer tr.Stop()

	tr.UserOnline("user-1", "room-1")

	time.Sleep(100 * time.Millisecond)  // within timeout
	tr.Heartbeat("user-1")

	time.Sleep(100 * time.Millisecond)  // would expire if not for heartbeat
	assert.True(t, tr.IsOnline("user-1"), "heartbeat should extend lifetime")
}

func Test_Tracker_Typing(t *testing.T) {
	tr := presence.NewTracker(100 * time.Millisecond)
	defer tr.Stop()

	tr.UserOnline("user-1", "room-1")

	tr.SetTyping("user-1", true)
	assert.True(t, tr.IsTyping("user-1"))

	tr.SetTyping("user-1", false)
	assert.False(t, tr.IsTyping("user-1"))
}

func Test_Tracker_GetUserStatus(t *testing.T) {
	tr := presence.NewTracker(100 * time.Millisecond)
	defer tr.Stop()

	tr.UserOnline("user-1", "room-1")

	status := tr.GetUserStatus("user-1")
	require.NotNil(t, status)
	assert.Equal(t, "online", status.Status)
	assert.Equal(t, "room-1", status.RoomID)

	tr.UserOffline("user-1")
	status = tr.GetUserStatus("user-1")
	require.NotNil(t, status)
	assert.Equal(t, "offline", status.Status)
}

func Test_Tracker_OnlineCount(t *testing.T) {
	tr := presence.NewTracker(100 * time.Millisecond)
	defer tr.Stop()

	assert.Equal(t, 0, tr.OnlineCount())

	tr.UserOnline("user-1", "room-1")
	tr.UserOnline("user-2", "room-1")
	assert.Equal(t, 2, tr.OnlineCount())

	tr.UserOffline("user-1")
	assert.Equal(t, 1, tr.OnlineCount())
}
