package metrics

import (
	"sync"
	"time"

	"github.com/stillnight88/infra-monitor/shared"
)

type AgentState struct {
	Payload shared.MetricsPayload
	Online  bool
}

// Store is a thread-safe registry of all connected agents, Every agent writes here. The /state endpoint reads from here.
type Store struct {
	mu   sync.RWMutex
	data map[string]AgentState
}

func New() *Store {
	return &Store{
		data: make(map[string]AgentState),
	}
}

// Set writes the latest metrics for an agent, Called from the WebSocket read loop — may be concurrent.
func (s *Store) Set(payload shared.MetricsPayload) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[payload.AgentID] = AgentState{
		Payload: payload,
		Online:  true,
	}
}

// All returns a snapshot of every agent's current state.
func (s *Store) All() map[string]AgentState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	snapshot := make(map[string]AgentState, len(s.data))
	for k, v := range s.data {
		snapshot[k] = v
	}
	return snapshot
}

// MarkOffline sets an agent's online flag to false.
func (s *Store) MarkOffline(agentID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if state, exists := s.data[agentID]; exists {
		state.Online = false
		s.data[agentID] = state
	}
}

// LastSeen returns the timestamp of the agent's last metric.
func (s *Store) LastSeen(agentID string) (int64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	state, ok := s.data[agentID]
	if !ok {
		return 0, false
	}
	return state.Payload.Timestamp, true
}

// isStale returns true if the agent hasn't sent metrics within the threshold.
func isStale(lastSeen int64, threshold time.Duration) bool {
	return time.Since(time.Unix(lastSeen, 0)) > threshold
}