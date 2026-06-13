package relay

import "sync"

// Router maps publisher tracks to subscribers for RTP forwarding.
type Router struct {
	// trackID → set of subscriber user IDs
	subscribers map[string]map[string]bool
	// userID → set of track IDs they publish
	publishers map[string]map[string]bool
	mu         sync.RWMutex
}

// NewRouter creates a new Router.
func NewRouter() *Router {
	return &Router{
		subscribers: make(map[string]map[string]bool),
		publishers:  make(map[string]map[string]bool),
	}
}

// AddPublisherTrack registers a track published by a user.
// Existing subscribers who previously subscribed to this track ID are preserved.
func (r *Router) AddPublisherTrack(userID, trackID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.publishers[userID] == nil {
		r.publishers[userID] = make(map[string]bool)
	}
	r.publishers[userID][trackID] = true

	if r.subscribers[trackID] == nil {
		r.subscribers[trackID] = make(map[string]bool)
	}
}

// Subscribe adds a subscriber to a track.
func (r *Router) Subscribe(subscriberID, trackID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.subscribers[trackID] == nil {
		r.subscribers[trackID] = make(map[string]bool)
	}
	r.subscribers[trackID][subscriberID] = true
}

// Unsubscribe removes a subscriber from a track.
func (r *Router) Unsubscribe(subscriberID, trackID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if subs, ok := r.subscribers[trackID]; ok {
		delete(subs, subscriberID)
		if len(subs) == 0 {
			delete(r.subscribers, trackID)
		}
	}
}

// RemoveTrack removes a track and all its subscribers.
func (r *Router) RemoveTrack(trackID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.subscribers, trackID)

	// Remove from all publishers' track lists
	for userID := range r.publishers {
		delete(r.publishers[userID], trackID)
	}
}

// RemoveAllPublisherTracks removes all tracks for a given publisher.
func (r *Router) RemoveAllPublisherTracks(userID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if tracks, ok := r.publishers[userID]; ok {
		for trackID := range tracks {
			delete(r.subscribers, trackID)
		}
		delete(r.publishers, userID)
	}
}

// GetSubscribers returns all subscriber IDs for a track.
func (r *Router) GetSubscribers(trackID string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	subs := r.subscribers[trackID]
	result := make([]string, 0, len(subs))
	for id := range subs {
		result = append(result, id)
	}
	return result
}

// GetPublisherTracks returns all track IDs for a publisher.
func (r *Router) GetPublisherTracks(userID string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tracks := r.publishers[userID]
	result := make([]string, 0, len(tracks))
	for id := range tracks {
		result = append(result, id)
	}
	return result
}
