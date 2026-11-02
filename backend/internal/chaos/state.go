package chaos

import (
	"sync"
	"time"
)

type State struct {
	mu       sync.RWMutex
	redisEnd time.Time
	dbDelay  time.Duration
}

func New() *State { return &State{} }

func (s *State) RedisDownFor(duration time.Duration) {
	s.mu.Lock()
	s.redisEnd = time.Now().Add(duration)
	s.mu.Unlock()
}

func (s *State) RedisDown() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return time.Now().Before(s.redisEnd)
}

func (s *State) SetDBLatency(duration time.Duration) {
	s.mu.Lock()
	s.dbDelay = duration
	s.mu.Unlock()
}

func (s *State) DBLatency() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.dbDelay
}

func (s *State) Snapshot() map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	remaining := time.Until(s.redisEnd)
	if remaining < 0 {
		remaining = 0
	}
	return map[string]any{
		"redis_down":        remaining > 0,
		"redis_down_for_ms": remaining.Milliseconds(),
		"db_latency_ms":     s.dbDelay.Milliseconds(),
	}
}
