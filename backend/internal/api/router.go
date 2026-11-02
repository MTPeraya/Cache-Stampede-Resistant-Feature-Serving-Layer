package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/MTPeraya/stampede-shield/backend/internal/cache"
	"github.com/MTPeraya/stampede-shield/backend/internal/features"
	"github.com/MTPeraya/stampede-shield/backend/internal/metrics"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewRouter(service *features.Service, redisCache *cache.Client, m *metrics.Metrics) http.Handler {
	r := chi.NewRouter()
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})
	r.Get("/features/{entity_id}", featureHandler(service))
	r.Post("/mode", modeHandler(service))
	r.Get("/metrics/stream", metricsStreamHandler(service, m))
	r.Post("/loadgen/burst", burstHandler(service))
	r.Post("/chaos/flush-cache", flushCacheHandler(redisCache))
	r.Post("/chaos/redis-down", redisDownHandler(service))
	r.Post("/chaos/db-latency", dbLatencyHandler(service))
	r.Get("/chaos/state", chaosStateHandler(service))
	r.Handle("/metrics", promhttp.Handler())
	return r
}

func metricsStreamHandler(service *features.Service, m *metrics.Metrics) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		encoder := json.NewEncoder(w)
		for {
			select {
			case <-r.Context().Done():
				return
			case <-ticker.C:
				_, _ = w.Write([]byte("data: "))
				snapshot := m.Snapshot(string(service.Mode()))
				chaosState := service.ChaosState().Snapshot()
				if chaosState["redis_down"] == true {
					snapshot.ChaosEvents = append(snapshot.ChaosEvents, "redis-down")
				}
				if chaosState["db_latency_ms"].(int64) > 0 {
					snapshot.ChaosEvents = append(snapshot.ChaosEvents, "db-latency")
				}
				if err := encoder.Encode(snapshot); err != nil {
					return
				}
				_, _ = w.Write([]byte("\n"))
				flusher.Flush()
			}
		}
	}
}

func burstHandler(service *features.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rps := queryInt(r, "rps", 100)
		seconds := queryInt(r, "seconds", 10)
		keys := queryInt(r, "keys", 10)
		if rps < 1 || rps > 10000 || seconds < 1 || seconds > 300 || keys < 1 || keys > 1000000 {
			http.Error(w, "rps, seconds, and keys are out of range", http.StatusBadRequest)
			return
		}
		go service.RunBurst(rps, seconds, keys)
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("burst started\n"))
	}
}

func queryInt(r *http.Request, name string, fallback int) int {
	value, err := strconv.Atoi(r.URL.Query().Get(name))
	if err != nil {
		return fallback
	}
	return value
}

func modeHandler(service *features.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mode := features.Mode(r.URL.Query().Get("value"))
		if mode != features.ModeNaive && mode != features.ModeHardened {
			http.Error(w, "value must be naive or hardened", http.StatusBadRequest)
			return
		}
		if err := service.SetMode(mode); err != nil {
			http.Error(w, "value must be naive or hardened", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func flushCacheHandler(redisCache *cache.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := redisCache.Flush(r.Context()); err != nil {
			http.Error(w, "cache flush failed", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func redisDownHandler(service *features.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		seconds := queryInt(r, "seconds", 10)
		if seconds < 1 || seconds > 3600 {
			http.Error(w, "seconds must be between 1 and 3600", http.StatusBadRequest)
			return
		}
		service.ChaosState().RedisDownFor(time.Duration(seconds) * time.Second)
		w.WriteHeader(http.StatusNoContent)
	}
}

func dbLatencyHandler(service *features.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ms := queryInt(r, "ms", 0)
		if ms < 0 || ms > 30000 {
			http.Error(w, "ms must be between 0 and 30000", http.StatusBadRequest)
			return
		}
		service.ChaosState().SetDBLatency(time.Duration(ms) * time.Millisecond)
		w.WriteHeader(http.StatusNoContent)
	}
}

func chaosStateHandler(service *features.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		state := service.ChaosState().Snapshot()
		state["mode"] = service.Mode()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(state)
	}
}

func featureHandler(service *features.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer service.Observe(start)
		entityID, err := strconv.ParseInt(chi.URLParam(r, "entity_id"), 10, 64)
		if err != nil || entityID < 1 {
			http.Error(w, "entity_id must be a positive integer", http.StatusBadRequest)
			return
		}
		feature, err := service.Get(r.Context(), entityID)
		if err != nil {
			if errors.Is(err, features.ErrLoadShed) {
				http.Error(w, "database capacity exhausted", http.StatusTooManyRequests)
				return
			}
			if errors.Is(err, context.DeadlineExceeded) {
				http.Error(w, "upstream timeout", http.StatusServiceUnavailable)
				return
			}
			http.Error(w, "feature unavailable", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(feature)
	}
}
