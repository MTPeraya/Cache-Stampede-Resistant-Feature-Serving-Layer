package features

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/MTPeraya/stampede-shield/backend/internal/cache"
	"github.com/MTPeraya/stampede-shield/backend/internal/chaos"
	"github.com/MTPeraya/stampede-shield/backend/internal/metrics"
	"golang.org/x/sync/singleflight"
)

type StoreReader interface {
	Get(context.Context, int64) (Feature, error)
}

type CacheReader interface {
	Get(context.Context, string) (cache.Entry, error)
	Set(context.Context, string, cache.Entry, time.Duration, time.Duration) error
}

type Mode string

const (
	ModeNaive    Mode = "naive"
	ModeHardened Mode = "hardened"
)

type Service struct {
	store   StoreReader
	cache   CacheReader
	metrics *metrics.Metrics
	group   singleflight.Group
	mode    Mode
	mu      sync.RWMutex
	randMu  sync.Mutex
	rand    *rand.Rand
	chaos   *chaos.State
	dbSlots chan struct{}
}

func NewService(store StoreReader, cacheClient CacheReader, m *metrics.Metrics, states ...*chaos.State) *Service {
	state := chaos.New()
	if len(states) > 0 && states[0] != nil {
		state = states[0]
	}
	return &Service{store: store, cache: cacheClient, metrics: m, mode: ModeNaive, rand: rand.New(rand.NewSource(time.Now().UnixNano())), chaos: state, dbSlots: make(chan struct{}, 100)}
}

func (s *Service) SetMode(mode Mode) error {
	if mode != ModeNaive && mode != ModeHardened {
		return fmt.Errorf("unsupported mode %q", mode)
	}
	s.mu.Lock()
	s.mode = mode
	s.mu.Unlock()
	return nil
}

func (s *Service) Mode() Mode {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.mode
}

func (s *Service) Get(ctx context.Context, entityID int64) (Feature, error) {
	s.metrics.IncRequests()
	key := fmt.Sprintf("feature:%d", entityID)
	entry, err := s.cache.Get(ctx, key)
	if err == nil {
		feature, decodeErr := decode(entry.Data)
		if decodeErr != nil {
			return Feature{}, decodeErr
		}
		if s.Mode() == ModeHardened && time.Now().After(entry.SoftExpires) {
			s.metrics.IncStaleServes()
			go s.refresh(key, entityID)
		} else {
			s.metrics.IncCacheHits()
		}
		return feature, nil
	}
	s.metrics.IncCacheMisses()
	return s.loadMiss(ctx, key, entityID)
}

func (s *Service) loadMiss(ctx context.Context, key string, entityID int64) (Feature, error) {
	if s.Mode() == ModeNaive {
		return s.queryAndCache(ctx, key, entityID)
	}
	value, err, shared := s.group.Do(key, func() (any, error) {
		callCtx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
		defer cancel()
		return s.queryAndCache(callCtx, key, entityID)
	})
	if shared {
		s.metrics.IncCoalesced()
	}
	if err != nil {
		return Feature{}, err
	}
	return value.(Feature), nil
}

func (s *Service) refresh(key string, entityID int64) {
	refreshKey := "refresh:" + key
	_, _, _ = s.group.Do(refreshKey, func() (any, error) {
		callCtx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
		defer cancel()
		return nil, func() error {
			_, err := s.queryAndCache(callCtx, key, entityID)
			return err
		}()
	})
}

func (s *Service) queryAndCache(ctx context.Context, key string, entityID int64) (Feature, error) {
	select {
	case s.dbSlots <- struct{}{}:
		defer func() { <-s.dbSlots }()
	default:
		s.metrics.IncShed()
		return Feature{}, ErrLoadShed
	}
	s.metrics.BeginDBQuery()
	defer s.metrics.EndDBQuery()
	if delay := s.chaos.DBLatency(); delay > 0 {
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return Feature{}, ctx.Err()
		case <-timer.C:
		}
	}
	feature, err := s.store.Get(ctx, entityID)
	if err != nil {
		return Feature{}, err
	}
	value, err := json.Marshal(feature)
	if err != nil {
		return Feature{}, fmt.Errorf("encode feature: %w", err)
	}

	softTTL, hardTTL := s.ttls()
	_ = s.cache.Set(ctx, key, cache.Entry{Data: value, SoftExpires: time.Now().Add(softTTL)}, softTTL, hardTTL)
	return feature, nil
}

var ErrLoadShed = fmt.Errorf("database load shed")

func (s *Service) ChaosState() *chaos.State { return s.chaos }

func (s *Service) ttls() (time.Duration, time.Duration) {
	s.randMu.Lock()
	soft, hard := jitteredTTLs(s.rand)
	s.randMu.Unlock()
	return soft, hard
}

func jitteredTTLs(r *rand.Rand) (time.Duration, time.Duration) {
	jitter := r.Float64() * 0.2
	return 30*time.Second + time.Duration(float64(30*time.Second)*jitter),
		5*time.Minute + time.Duration(float64(5*time.Minute)*jitter)
}

func decode(data []byte) (Feature, error) {
	var feature Feature
	if err := json.Unmarshal(data, &feature); err != nil {
		return Feature{}, fmt.Errorf("decode cached feature: %w", err)
	}
	return feature, nil
}

func (s *Service) Observe(start time.Time) {
	s.metrics.Observe(start)
}

func (s *Service) RunBurst(rps, seconds, keys int) {
	interval := time.Second / time.Duration(rps)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	deadline := time.After(time.Duration(seconds) * time.Second)
	for {
		select {
		case <-deadline:
			return
		case <-ticker.C:
			entityID := int64(s.randIntn(keys) + 1)
			go func() { _, _ = s.Get(context.Background(), entityID) }()
		}
	}
}

func (s *Service) randIntn(max int) int {
	s.randMu.Lock()
	defer s.randMu.Unlock()
	return s.rand.Intn(max)
}
