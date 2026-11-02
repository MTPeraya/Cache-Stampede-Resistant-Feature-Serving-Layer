package metrics

import (
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Snapshot struct {
	Timestamp    time.Time `json:"timestamp"`
	Requests     uint64    `json:"requests"`
	CacheHits    uint64    `json:"cache_hits"`
	CacheMisses  uint64    `json:"cache_misses"`
	Coalesced    uint64    `json:"coalesced"`
	StaleServes  uint64    `json:"stale_serves"`
	DBQueries    uint64    `json:"db_queries"`
	DBQPS        float64   `json:"db_qps"`
	HitRate      float64   `json:"hit_rate"`
	P50LatencyMS float64   `json:"p50_latency_ms"`
	P95LatencyMS float64   `json:"p95_latency_ms"`
	P99LatencyMS float64   `json:"p99_latency_ms"`
	InFlightDB   int64     `json:"in_flight_db"`
	ShedRequests uint64    `json:"shed_requests"`
	Mode         string    `json:"mode"`
	ChaosEvents  []string  `json:"chaos_events"`
}

type Metrics struct {
	Requests    prometheus.Counter
	CacheHits   prometheus.Counter
	CacheMisses prometheus.Counter
	Coalesced   prometheus.Counter
	StaleServes prometheus.Counter
	DBQueries   prometheus.Counter
	Latency     prometheus.Histogram

	requests    atomic.Uint64
	cacheHits   atomic.Uint64
	cacheMisses atomic.Uint64
	coalesced   atomic.Uint64
	staleServes atomic.Uint64
	dbQueries   atomic.Uint64
	inFlightDB  atomic.Int64
	shed        atomic.Uint64
	mu          sync.Mutex
	latencies   []float64
	dbQueryAt   []time.Time
}

func New() *Metrics {
	return NewWithRegisterer(prometheus.DefaultRegisterer)
}

func NewWithRegisterer(registerer prometheus.Registerer) *Metrics {
	m := &Metrics{
		Requests:    prometheus.NewCounter(prometheus.CounterOpts{Name: "stampede_requests_total", Help: "Feature requests."}),
		CacheHits:   prometheus.NewCounter(prometheus.CounterOpts{Name: "stampede_cache_hits_total", Help: "Cache hits."}),
		CacheMisses: prometheus.NewCounter(prometheus.CounterOpts{Name: "stampede_cache_misses_total", Help: "Cache misses."}),
		Coalesced:   prometheus.NewCounter(prometheus.CounterOpts{Name: "stampede_coalesced_requests_total", Help: "Requests sharing an in-flight database read."}),
		StaleServes: prometheus.NewCounter(prometheus.CounterOpts{Name: "stampede_stale_serves_total", Help: "Requests served stale data."}),
		DBQueries:   prometheus.NewCounter(prometheus.CounterOpts{Name: "stampede_db_queries_total", Help: "Database queries."}),
		Latency:     prometheus.NewHistogram(prometheus.HistogramOpts{Name: "stampede_request_duration_seconds", Help: "Request latency.", Buckets: prometheus.DefBuckets}),
	}
	for _, collector := range []prometheus.Collector{m.Requests, m.CacheHits, m.CacheMisses, m.Coalesced, m.StaleServes, m.DBQueries, m.Latency} {
		if err := registerer.Register(collector); err != nil {
			panic(err)
		}
	}
	return m
}

func (m *Metrics) IncRequests()    { m.Requests.Inc(); m.requests.Add(1) }
func (m *Metrics) IncCacheHits()   { m.CacheHits.Inc(); m.cacheHits.Add(1) }
func (m *Metrics) IncCacheMisses() { m.CacheMisses.Inc(); m.cacheMisses.Add(1) }
func (m *Metrics) IncCoalesced()   { m.Coalesced.Inc(); m.coalesced.Add(1) }
func (m *Metrics) IncStaleServes() { m.StaleServes.Inc(); m.staleServes.Add(1) }

func (m *Metrics) BeginDBQuery() {
	m.DBQueries.Inc()
	m.dbQueries.Add(1)
	m.inFlightDB.Add(1)
	m.mu.Lock()
	m.dbQueryAt = append(m.dbQueryAt, time.Now())
	m.mu.Unlock()
}

func (m *Metrics) EndDBQuery() { m.inFlightDB.Add(-1) }
func (m *Metrics) IncShed()    { m.shed.Add(1) }

func (m *Metrics) Observe(start time.Time) {
	seconds := time.Since(start).Seconds()
	m.Latency.Observe(seconds)
	m.mu.Lock()
	m.latencies = append(m.latencies, seconds)
	if len(m.latencies) > 2000 {
		m.latencies = m.latencies[len(m.latencies)-2000:]
	}
	m.mu.Unlock()
}

func (m *Metrics) Snapshot(mode string) Snapshot {
	now := time.Now()
	m.mu.Lock()
	latencies := append([]float64(nil), m.latencies...)
	cutoff := now.Add(-time.Second)
	first := 0
	for first < len(m.dbQueryAt) && m.dbQueryAt[first].Before(cutoff) {
		first++
	}
	m.dbQueryAt = append([]time.Time(nil), m.dbQueryAt[first:]...)
	dbQPS := float64(len(m.dbQueryAt))
	m.mu.Unlock()
	sort.Float64s(latencies)
	requests := m.requests.Load()
	hits := m.cacheHits.Load()
	snapshot := Snapshot{
		Timestamp: now, Requests: requests, CacheHits: hits, CacheMisses: m.cacheMisses.Load(),
		Coalesced: m.coalesced.Load(), StaleServes: m.staleServes.Load(), DBQueries: m.dbQueries.Load(),
		DBQPS: dbQPS, InFlightDB: m.inFlightDB.Load(), ShedRequests: m.shed.Load(), Mode: mode,
		ChaosEvents: []string{},
	}
	if requests > 0 {
		snapshot.HitRate = float64(hits) / float64(requests)
	}
	snapshot.P50LatencyMS = quantile(latencies, 0.50) * 1000
	snapshot.P95LatencyMS = quantile(latencies, 0.95) * 1000
	snapshot.P99LatencyMS = quantile(latencies, 0.99) * 1000
	return snapshot
}

func quantile(values []float64, q float64) float64 {
	if len(values) == 0 {
		return 0
	}
	index := int(float64(len(values)-1) * q)
	return values[index]
}
