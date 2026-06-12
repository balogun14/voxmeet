package room

import (
	"sync"
)

// Manager manages all active rooms.
type Manager struct {
	rooms map[string]*Room
	mu    sync.RWMutex
}

// NewManager creates a new room Manager.
func NewManager() *Manager {
	return &Manager{
		rooms: make(map[string]*Room),
	}
}

// GetOrCreateRoom returns the room with the given ID, creating one if it doesn't exist.
func (m *Manager) GetOrCreateRoom(roomID string) *Room {
	m.mu.Lock()
	defer m.mu.Unlock()

	if r, ok := m.rooms[roomID]; ok {
		return r
	}

	r := NewRoom(roomID)
	m.rooms[roomID] = r
	return r
}

// GetRoom returns the room with the given ID, or false if it doesn't exist.
func (m *Manager) GetRoom(roomID string) (*Room, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	r, ok := m.rooms[roomID]
	return r, ok
}

// RemoveRoom removes and stops the room with the given ID.
func (m *Manager) RemoveRoom(roomID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if r, ok := m.rooms[roomID]; ok {
		r.Stop()
		delete(m.rooms, roomID)
	}
}

// RoomCount returns the number of active rooms.
func (m *Manager) RoomCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.rooms)
}

// StopAll stops all rooms and clears the manager.
func (m *Manager) StopAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, r := range m.rooms {
		r.Stop()
		delete(m.rooms, id)
	}
}

// Room represents a media session with multiple peers.
type Room struct {
	ID    string
	peers map[string]*Peer
	mu    sync.RWMutex
	done  chan struct{}
}

// NewRoom creates a new Room.
func NewRoom(id string) *Room {
	return &Room{
		ID:    id,
		peers: make(map[string]*Peer),
		done:  make(chan struct{}),
	}
}

// AddPeer adds a peer to the room. Returns the existing peer if already present.
func (r *Room) AddPeer(userID string) *Peer {
	r.mu.Lock()
	defer r.mu.Unlock()

	if p, ok := r.peers[userID]; ok {
		return p
	}

	p := NewPeer(userID, r.ID)
	r.peers[userID] = p
	return p
}

// RemovePeer removes a peer from the room.
func (r *Room) RemovePeer(userID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if p, ok := r.peers[userID]; ok {
		p.Stop()
		delete(r.peers, userID)
	}
}

// GetPeer returns a peer by user ID.
func (r *Room) GetPeer(userID string) (*Peer, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.peers[userID]
	return p, ok
}

// GetPeers returns all peers in the room.
func (r *Room) GetPeers() []*Peer {
	r.mu.RLock()
	defer r.mu.RUnlock()

	peers := make([]*Peer, 0, len(r.peers))
	for _, p := range r.peers {
		peers = append(peers, p)
	}
	return peers
}

// PeerCount returns the number of peers in the room.
func (r *Room) PeerCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.peers)
}

// Stop closes all peer connections and cleans up the room.
func (r *Room) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for id, p := range r.peers {
		p.Stop()
		delete(r.peers, id)
	}

	select {
	case <-r.done:
	default:
		close(r.done)
	}
}

// Done returns a channel that's closed when the room is stopped.
func (r *Room) Done() <-chan struct{} {
	return r.done
}

// Peer represents a participant in a room.
type Peer struct {
	UserID string
	RoomID string
	tracks map[string]*Track
	mu     sync.RWMutex
	done   chan struct{}
}

// NewPeer creates a new Peer.
func NewPeer(userID, roomID string) *Peer {
	return &Peer{
		UserID: userID,
		RoomID: roomID,
		tracks: make(map[string]*Track),
		done:   make(chan struct{}),
	}
}

// AddTrack adds a track to this peer. Returns existing track if duplicate.
func (p *Peer) AddTrack(id, kind, source string) *Track {
	p.mu.Lock()
	defer p.mu.Unlock()

	if t, ok := p.tracks[id]; ok {
		return t
	}

	t := &Track{
		ID:     id,
		Kind:   kind,
		Source: source,
	}
	p.tracks[id] = t
	return t
}

// RemoveTrack removes a track from this peer.
func (p *Peer) RemoveTrack(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.tracks, id)
}

// Tracks returns all tracks for this peer.
func (p *Peer) Tracks() []*Track {
	p.mu.RLock()
	defer p.mu.RUnlock()

	tracks := make([]*Track, 0, len(p.tracks))
	for _, t := range p.tracks {
		tracks = append(tracks, t)
	}
	return tracks
}

// Stop cleans up the peer.
func (p *Peer) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.tracks = make(map[string]*Track)

	select {
	case <-p.done:
	default:
		close(p.done)
	}
}

// Done returns a channel that's closed when the peer is stopped.
func (p *Peer) Done() <-chan struct{} {
	return p.done
}

// Track represents a media track published by a peer.
type Track struct {
	ID     string
	Kind   string
	Source string
}
