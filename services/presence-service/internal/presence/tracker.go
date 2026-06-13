package presence

import (
	"sync"
	"time"
)

// UserStatus represents a user's current presence state.
type UserStatus struct {
	UserID    string `json:"user_id"`
	Status    string `json:"status"` // "online" or "offline"
	RoomID    string `json:"room_id,omitempty"`
	Typing    bool   `json:"typing,omitempty"`
	LastSeen  string `json:"last_seen"`
}

// userState is the internal state for a single user.
type userState struct {
	roomID   string
	typing   bool
	lastSeen time.Time
}

// Tracker manages online presence with heartbeat timeout.
type Tracker struct {
	users    map[string]*userState
	mu       sync.RWMutex
	timeout  time.Duration
	stopCh   chan struct{}
}

// NewTracker creates a new presence tracker with the given heartbeat timeout.
func NewTracker(heartbeatTimeout time.Duration) *Tracker {
	tr := &Tracker{
		users:   make(map[string]*userState),
		timeout: heartbeatTimeout,
		stopCh:  make(chan struct{}),
	}
	go tr.cleanupLoop()
	return tr
}

// UserOnline marks a user as online in a room.
func (t *Tracker) UserOnline(userID, roomID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.users[userID] = &userState{
		roomID:   roomID,
		lastSeen: time.Now(),
	}
}

// UserOffline marks a user as offline.
func (t *Tracker) UserOffline(userID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.users, userID)
}

// IsOnline returns true if the user is currently tracked as online.
func (t *Tracker) IsOnline(userID string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	_, ok := t.users[userID]
	return ok
}

// Heartbeat refreshes the user's last-seen timestamp.
func (t *Tracker) Heartbeat(userID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if state, ok := t.users[userID]; ok {
		state.lastSeen = time.Now()
	}
}

// GetRoomUsers returns all online user IDs in a room.
func (t *Tracker) GetRoomUsers(roomID string) []string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var result []string
	for id, state := range t.users {
		if state.roomID == roomID {
			result = append(result, id)
		}
	}
	return result
}

// SetTyping sets or clears the typing indicator for a user.
func (t *Tracker) SetTyping(userID string, typing bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if state, ok := t.users[userID]; ok {
		state.typing = typing
	}
}

// IsTyping returns whether the user is currently typing.
func (t *Tracker) IsTyping(userID string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if state, ok := t.users[userID]; ok {
		return state.typing
	}
	return false
}

// GetUserStatus returns the full status for a user.
func (t *Tracker) GetUserStatus(userID string) *UserStatus {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if state, ok := t.users[userID]; ok {
		return &UserStatus{
			UserID:   userID,
			Status:   "online",
			RoomID:   state.roomID,
			Typing:   state.typing,
			LastSeen: state.lastSeen.Format(time.RFC3339),
		}
	}

	return &UserStatus{
		UserID: userID,
		Status: "offline",
	}
}

// OnlineCount returns the number of currently online users.
func (t *Tracker) OnlineCount() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.users)
}

// Stop stops the cleanup loop.
func (t *Tracker) Stop() {
	select {
	case <-t.stopCh:
	default:
		close(t.stopCh)
	}
}

// cleanupLoop periodically removes users who haven't sent a heartbeat.
func (t *Tracker) cleanupLoop() {
	ticker := time.NewTicker(t.timeout / 2)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			t.mu.Lock()
			now := time.Now()
			for id, state := range t.users {
				if now.Sub(state.lastSeen) > t.timeout {
					delete(t.users, id)
				}
			}
			t.mu.Unlock()
		case <-t.stopCh:
			return
		}
	}
}
