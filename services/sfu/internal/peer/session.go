package peer

import (
	"fmt"
	"sync"

	"github.com/pion/webrtc/v4"
)

// Config holds Pion configuration for a peer session.
type Config struct {
	ICEServers    []string
	OnTrack       func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver)
	OnICE         func(candidate string, sdpMid string, sdpMLineIndex uint16)
	OnStateChange func(state string)
}

// Session wraps a Pion PeerConnection for media.
type Session struct {
	pc     *webrtc.PeerConnection
	cfg    Config
	mu     sync.Mutex
	closed bool
}

// NewSession creates a new PeerConnection session.
func NewSession(cfg Config) (*Session, error) {
	if len(cfg.ICEServers) == 0 {
		return nil, fmt.Errorf("at least one ICE server is required")
	}

	m := &webrtc.MediaEngine{}
	if err := m.RegisterDefaultCodecs(); err != nil {
		return nil, fmt.Errorf("register codecs: %w", err)
	}

	api := webrtc.NewAPI(webrtc.WithMediaEngine(m))

	iceServers := make([]webrtc.ICEServer, 0, len(cfg.ICEServers))
	for _, u := range cfg.ICEServers {
		iceServers = append(iceServers, webrtc.ICEServer{URLs: []string{u}})
	}

	pc, err := api.NewPeerConnection(webrtc.Configuration{
		ICEServers: iceServers,
	})
	if err != nil {
		return nil, fmt.Errorf("create peer connection: %w", err)
	}

	s := &Session{pc: pc, cfg: cfg}

	if cfg.OnStateChange != nil {
		pc.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
			cfg.OnStateChange(state.String())
		})
	}

	if cfg.OnTrack != nil {
		pc.OnTrack(func(remote *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
			cfg.OnTrack(remote, receiver)
		})
	}

	if cfg.OnICE != nil {
		pc.OnICECandidate(func(c *webrtc.ICECandidate) {
			if c == nil {
				return
			}
			init := c.ToJSON()
			line := uint16(0)
			if init.SDPMLineIndex != nil {
				line = uint16(*init.SDPMLineIndex)
			}
			mid := ""
			if init.SDPMid != nil {
				mid = *init.SDPMid
			}
			cfg.OnICE(init.Candidate, mid, line)
		})
	}

	return s, nil
}

// CreateOffer generates an SDP offer and sets it as local description.
func (s *Session) CreateOffer() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return "", fmt.Errorf("session closed")
	}
	offer, err := s.pc.CreateOffer(nil)
	if err != nil {
		return "", fmt.Errorf("create offer: %w", err)
	}
	if err := s.pc.SetLocalDescription(offer); err != nil {
		return "", fmt.Errorf("set local description: %w", err)
	}
	return offer.SDP, nil
}

// SetAnswer sets the remote SDP description from the client.
func (s *Session) SetAnswer(sdp string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return fmt.Errorf("session closed")
	}
	return s.pc.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer,
		SDP:  sdp,
	})
}

// AddICECandidate adds a remote ICE candidate.
func (s *Session) AddICECandidate(candidate string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return fmt.Errorf("session closed")
	}
	return s.pc.AddICECandidate(webrtc.ICECandidateInit{Candidate: candidate})
}

// AddTrack creates a local track from a remote one and adds it to the PC.
func (s *Session) AddTrack(remote *webrtc.TrackRemote) (*webrtc.TrackLocalStaticRTP, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil, fmt.Errorf("session closed")
	}
	local, err := webrtc.NewTrackLocalStaticRTP(remote.Codec().RTPCodecCapability, remote.ID(), remote.StreamID())
	if err != nil {
		return nil, fmt.Errorf("create local track: %w", err)
	}
	if _, err := s.pc.AddTrack(local); err != nil {
		return nil, fmt.Errorf("add track: %w", err)
	}
	return local, nil
}

// CreateDownTrack creates a new local track for sending media to this peer.
func (s *Session) CreateDownTrack(codec webrtc.RTPCodecCapability, id, streamID string) (*webrtc.TrackLocalStaticRTP, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil, fmt.Errorf("session closed")
	}
	local, err := webrtc.NewTrackLocalStaticRTP(codec, id, streamID)
	if err != nil {
		return nil, fmt.Errorf("create down track: %w", err)
	}
	if _, err := s.pc.AddTrack(local); err != nil {
		return nil, fmt.Errorf("add down track: %w", err)
	}
	return local, nil
}

// Close closes the peer connection.
func (s *Session) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	s.closed = true
	_ = s.pc.Close()
}

// State returns the current ICE connection state.
func (s *Session) State() *State {
	s.mu.Lock()
	defer s.mu.Unlock()
	state := "new"
	if s.pc != nil {
		state = s.pc.ICEConnectionState().String()
	}
	return &State{ConnectionState: state}
}

// State represents the current state of a peer session.
type State struct {
	ConnectionState string `json:"connection_state"`
}
