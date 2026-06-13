package relay_test

import (
	"testing"

	"github.com/awwal/voxmeet/sfu/internal/relay"
	"github.com/stretchr/testify/assert"
)

func Test_NewRouter(t *testing.T) {
	r := relay.NewRouter()
	assert.NotNil(t, r)
}

func Test_Router_Subscribe(t *testing.T) {
	r := relay.NewRouter()

	// Add publisher track
	r.AddPublisherTrack("user-1", "track-1")

	// Subscribe user-2 to user-1's track
	r.Subscribe("user-2", "track-1")
	subs := r.GetSubscribers("track-1")
	assert.Contains(t, subs, "user-2")
}

func Test_Router_Unsubscribe(t *testing.T) {
	r := relay.NewRouter()

	r.AddPublisherTrack("user-1", "track-1")
	r.Subscribe("user-2", "track-1")
	r.Subscribe("user-3", "track-1")

	r.Unsubscribe("user-2", "track-1")
	subs := r.GetSubscribers("track-1")
	assert.Len(t, subs, 1)
	assert.Contains(t, subs, "user-3")
}

func Test_Router_RemoveTrack(t *testing.T) {
	r := relay.NewRouter()

	r.AddPublisherTrack("user-1", "track-1")
	r.AddPublisherTrack("user-1", "track-2")

	r.RemoveTrack("track-1")
	subs := r.GetSubscribers("track-1")
	assert.Len(t, subs, 0)
}

func Test_Router_RemoveAllPublisherTracks(t *testing.T) {
	r := relay.NewRouter()

	r.AddPublisherTrack("user-1", "track-1")
	r.AddPublisherTrack("user-1", "track-2")

	r.RemoveAllPublisherTracks("user-1")
	assert.Len(t, r.GetSubscribers("track-1"), 0)
	assert.Len(t, r.GetSubscribers("track-2"), 0)
}

func Test_Router_SubscribersAutoSubscribed(t *testing.T) {
	r := relay.NewRouter()

	// When a track is added, existing subscribers should auto-subscribe
	r.Subscribe("user-2", "track-1")
	r.AddPublisherTrack("user-1", "track-1")

	subs := r.GetSubscribers("track-1")
	assert.Contains(t, subs, "user-2")
}

func Test_Router_GetPublisherTracks(t *testing.T) {
	r := relay.NewRouter()

	r.AddPublisherTrack("user-1", "track-1")
	r.AddPublisherTrack("user-1", "track-2")

	tracks := r.GetPublisherTracks("user-1")
	assert.Len(t, tracks, 2)
	assert.Contains(t, tracks, "track-1")
	assert.Contains(t, tracks, "track-2")
}

func Test_Router_Cleanup(t *testing.T) {
	r := relay.NewRouter()

	r.AddPublisherTrack("user-1", "track-1")
	r.Subscribe("user-2", "track-1")
	r.Subscribe("user-3", "track-1")

	r.RemoveTrack("track-1")
	assert.Len(t, r.GetSubscribers("track-1"), 0)
	assert.Len(t, r.GetPublisherTracks("user-1"), 0)
}
