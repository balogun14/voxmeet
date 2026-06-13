package peer_test

import (
	"testing"

	"github.com/awwal/voxmeet/sfu/internal/peer"
	"github.com/stretchr/testify/assert"
)

func Test_NewSession_InvalidConfig(t *testing.T) {
	_, err := peer.NewSession(peer.Config{})
	assert.Error(t, err)
}

func Test_NewSession_NilCallbacks(t *testing.T) {
	s, err := peer.NewSession(peer.Config{
		ICEServers: []string{"stun:stun.l.google.com:19302"},
	})
	assert.NoError(t, err)
	assert.NotNil(t, s)
	s.Close()
}

func Test_Session_State(t *testing.T) {
	s, err := peer.NewSession(peer.Config{
		ICEServers: []string{"stun:stun.l.google.com:19302"},
	})
	assert.NoError(t, err)
	defer s.Close()

	state := s.State()
	assert.NotNil(t, state)
	assert.Equal(t, "new", state.ConnectionState)
}

func Test_Session_CreateOffer(t *testing.T) {
	s, err := peer.NewSession(peer.Config{
		ICEServers: []string{"stun:stun.l.google.com:19302"},
	})
	assert.NoError(t, err)
	defer s.Close()

	sdp, err := s.CreateOffer()
	assert.NoError(t, err)
	assert.NotEmpty(t, sdp)
	assert.Contains(t, sdp, "v=0")
}

func Test_Session_SetAnswer_InvalidSDP(t *testing.T) {
	s, err := peer.NewSession(peer.Config{
		ICEServers: []string{"stun:stun.l.google.com:19302"},
	})
	assert.NoError(t, err)
	defer s.Close()

	err = s.SetAnswer("not-a-valid-sdp")
	assert.Error(t, err)
}

func Test_Session_AddICECandidate_Invalid(t *testing.T) {
	s, err := peer.NewSession(peer.Config{
		ICEServers: []string{"stun:stun.l.google.com:19302"},
	})
	assert.NoError(t, err)
	defer s.Close()

	_, err = s.CreateOffer()
	assert.NoError(t, err)

	err = s.AddICECandidate("invalid")
	assert.Error(t, err)
}

func Test_Session_OnTrack(t *testing.T) {
	s, err := peer.NewSession(peer.Config{
		ICEServers: []string{"stun:stun.l.google.com:19302"},
		OnTrack: func(kind, id string) {
			_ = kind
			_ = id
		},
	})
	assert.NoError(t, err)
	defer s.Close()

	assert.NotNil(t, s)
}

func Test_Session_MultipleCreateOffer(t *testing.T) {
	s, err := peer.NewSession(peer.Config{
		ICEServers: []string{"stun:stun.l.google.com:19302"},
	})
	assert.NoError(t, err)
	defer s.Close()

	sdp, err := s.CreateOffer()
	assert.NoError(t, err)
	assert.NotEmpty(t, sdp)

	// Second offer fails because PC is in have-local-offer state
	_, err = s.CreateOffer()
	assert.Error(t, err)
}

func Test_Session_Close_Twice(t *testing.T) {
	s, err := peer.NewSession(peer.Config{
		ICEServers: []string{"stun:stun.l.google.com:19302"},
	})
	assert.NoError(t, err)

	s.Close()
	s.Close() // Should not panic
}
