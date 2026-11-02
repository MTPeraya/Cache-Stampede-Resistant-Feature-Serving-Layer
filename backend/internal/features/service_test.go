package features

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/MTPeraya/stampede-shield/backend/internal/cache"
	"github.com/MTPeraya/stampede-shield/backend/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

type fakeStore struct {
	queries atomic.Int64
	delay   time.Duration
}

func (s *fakeStore) Get(ctx context.Context, entityID int64) (Feature, error) {
	s.queries.Add(1)
	select {
	case <-time.After(s.delay):
	case <-ctx.Done():
		return Feature{}, ctx.Err()
	}
	return Feature{EntityID: entityID, ClickRate: 0.12}, nil
}

type fakeCache struct {
	mu       sync.Mutex
	entries  map[string]cache.Entry
	getErr   error
	setErr   error
	setCount atomic.Int64
}

func newFakeCache() *fakeCache {
	return &fakeCache{entries: make(map[string]cache.Entry)}
}

func (c *fakeCache) Get(_ context.Context, key string) (cache.Entry, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.getErr != nil {
		return cache.Entry{}, c.getErr
	}
	entry, ok := c.entries[key]
	if !ok {
		return cache.Entry{}, cache.ErrMiss
	}
	return entry, nil
}

func (c *fakeCache) Set(_ context.Context, key string, entry cache.Entry, _, _ time.Duration) error {
	c.setCount.Add(1)
	if c.setErr != nil {
		return c.setErr
	}
	c.mu.Lock()
	c.entries[key] = entry
	c.mu.Unlock()
	return nil
}

func testService(t *testing.T, store StoreReader, cacheClient CacheReader) *Service {
	t.Helper()
	registry := prometheus.NewRegistry()
	return NewService(store, cacheClient, metrics.NewWithRegisterer(registry))
}

func TestHardenedCoalescesThousandRequests(t *testing.T) {
	store := &fakeStore{delay: 20 * time.Millisecond}
	service := testService(t, store, newFakeCache())
	if err := service.SetMode(ModeHardened); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	errs := make(chan error, 1000)
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := service.Get(context.Background(), 42)
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
	}
	if got := store.queries.Load(); got != 1 {
		t.Fatalf("database queries = %d, want 1", got)
	}
}

func TestTTLJitterBounds(t *testing.T) {
	randSource := rand.New(rand.NewSource(1))
	for i := 0; i < 1000; i++ {
		soft, hard := jitteredTTLs(randSource)
		if soft < 30*time.Second || soft > 36*time.Second {
			t.Fatalf("soft TTL %s outside [30s, 36s]", soft)
		}
		if hard < 5*time.Minute || hard > 6*time.Minute {
			t.Fatalf("hard TTL %s outside [5m, 6m]", hard)
		}
	}
}

func TestStaleReadsTriggerOneRefresh(t *testing.T) {
	store := &fakeStore{delay: 20 * time.Millisecond}
	c := newFakeCache()
	c.entries["feature:7"] = cache.Entry{
		Data:        []byte(`{"entity_id":7,"click_rate":0.1}`),
		SoftExpires: time.Now().Add(-time.Second),
	}
	service := testService(t, store, c)
	if err := service.SetMode(ModeHardened); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := service.Get(context.Background(), 7); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	time.Sleep(100 * time.Millisecond)
	if got := store.queries.Load(); got != 1 {
		t.Fatalf("refresh queries = %d, want 1", got)
	}
}

func TestRedisDownFallsBackToDatabase(t *testing.T) {
	store := &fakeStore{}
	c := newFakeCache()
	c.getErr = errors.New("redis unavailable")
	c.setErr = errors.New("redis unavailable")
	service := testService(t, store, c)
	if err := service.SetMode(ModeHardened); err != nil {
		t.Fatal(err)
	}
	feature, err := service.Get(context.Background(), 9)
	if err != nil {
		t.Fatal(err)
	}
	if feature.EntityID != 9 || store.queries.Load() != 1 {
		t.Fatalf("unexpected fallback result: %+v, queries=%d", feature, store.queries.Load())
	}
}
